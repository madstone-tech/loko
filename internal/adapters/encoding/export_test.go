package encoding

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

func sampleIR() *arch.IR {
	return arch.NewIR(
		arch.Project{Name: "acme", Description: "A & B <test>", LokoVersion: "~> 1.0"},
		[]arch.Element{
			{Address: "container.api", Kind: arch.KindContainer, Name: "api",
				Parent: "system.payments", Tags: []string{"pci", "public"},
				Range: arch.SourceRange{File: "a.loko.hcl", StartLine: 1, StartColumn: 1, EndLine: 1, EndColumn: 9}},
			{Address: "system.payments", Kind: arch.KindSystem, Name: "payments",
				Range: arch.SourceRange{File: "a.loko.hcl", StartLine: 10, StartColumn: 1, EndLine: 10, EndColumn: 9}},
		},
		[]arch.Relationship{
			{Address: "container.api.uses.db", Source: "container.api", Target: "system.payments", LocalName: "db"},
		},
		[]arch.Environment{{
			Address: "deployment.prod", Name: "prod", Provider: "aws",
			Groups: []arch.Group{{Address: "deployment.prod.node.vpc", Name: "vpc",
				Contains: []arch.Address{"deployment.prod.instance.api"}}},
			Instances: []arch.Instance{{
				Address: "deployment.prod.instance.api", Name: "api", Of: "container.api",
				PlacedIn: "deployment.prod.node.vpc",
				Attributes: []arch.Attribute{
					{Key: "memory", Value: arch.Value{Kind: arch.ValueNumber, Num: 1024}},
					{Key: "public", Value: arch.Value{Kind: arch.ValueBool, Bool: true}},
					{Key: "zone", Value: arch.Value{Kind: arch.ValueString, Str: "a"}},
				},
				Claims: []arch.Claim{{Kind: arch.ClaimTerraform, Address: "module.api.this"}},
			}},
		}},
		[]arch.View{{Address: "view.v", Name: "v", Include: []arch.Address{"container.api"}}},
		[]string{"aws_iam_role.*"},
	)
}

// TestExportIsByteIdenticalAcrossRuns covers SC-003 for both encodings.
func TestExportIsByteIdenticalAcrossRuns(t *testing.T) {
	t.Parallel()

	enc := NewEncoder()
	for _, format := range []usecases.ExportFormat{usecases.ExportJSON, usecases.ExportTOON} {
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()
			first, err := enc.EncodeIR(sampleIR(), format)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			for i := range 20 {
				got, err := enc.EncodeIR(sampleIR(), format)
				if err != nil {
					t.Fatalf("encode run %d: %v", i, err)
				}
				if !bytes.Equal(first, got) {
					t.Fatalf("run %d differs from the first", i)
				}
			}
		})
	}
}

// TestExportNoRunVaryingValues covers FR-036c. A timestamp, hostname, tool
// version, or absolute path would break byte-stability across machines while
// still passing a repeat-run test on one machine.
func TestExportNoRunVaryingValues(t *testing.T) {
	t.Parallel()

	out, err := NewEncoder().EncodeIR(sampleIR(), usecases.ExportJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	text := string(out)

	host, _ := os.Hostname()
	banned := []string{"timestamp", "generatedAt", "hostname", "/Users/", "/home/", "C:\\"}
	if host != "" {
		banned = append(banned, host)
	}
	for _, b := range banned {
		if strings.Contains(text, b) {
			t.Errorf("export contains run-varying value %q", b)
		}
	}
	// The tool version must not appear: it would change the bytes on every
	// release even when the shape has not changed (FR-036a).
	if strings.Contains(text, `"toolVersion"`) {
		t.Error("export embeds the tool version")
	}
}

// TestJSONDoesNotEscapeHTML: descriptions routinely contain & < >, and
// escaping them makes a committed artefact needlessly diff-noisy.
func TestJSONDoesNotEscapeHTML(t *testing.T) {
	t.Parallel()

	out, err := NewEncoder().EncodeIR(sampleIR(), usecases.ExportJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(string(out), "A & B <test>") {
		t.Errorf("HTML-escaped output: %s", firstLines(string(out), 6))
	}
}

// TestEncodingsCarryEquivalentInformation is FR-036. It is checked
// mechanically because the two backends are easy to let drift: a Stringer on
// one type silently collapsed the whole source range in TOON during
// development, and only a comparison like this caught it.
func TestEncodingsCarryEquivalentInformation(t *testing.T) {
	t.Parallel()

	enc := NewEncoder()
	jsonBytes, err := enc.EncodeIR(sampleIR(), usecases.ExportJSON)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	toonBytes, err := enc.EncodeIR(sampleIR(), usecases.ExportTOON)
	if err != nil {
		t.Fatalf("toon: %v", err)
	}

	var jsonDoc map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonDoc); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	// Every leaf key path in the JSON document must appear in the TOON output.
	toonText := string(toonBytes)
	var missing []string
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, sub := range t {
				if !toonHasKey(toonText, k) {
					missing = append(missing, prefix+"."+k)
					continue
				}
				walk(prefix+"."+k, sub)
			}
		case []any:
			for _, item := range t {
				walk(prefix+"[]", item)
			}
		}
	}
	walk("", jsonDoc)

	if len(missing) > 0 {
		t.Errorf("TOON is missing keys present in JSON: %v", missing)
	}
}

// TestExportRoundTrip: an artefact can be read back, which the diff stage
// depends on for --from-export.
func TestExportRoundTrip(t *testing.T) {
	t.Parallel()

	enc := NewEncoder()
	original, err := enc.EncodeIR(sampleIR(), usecases.ExportJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := DecodeIRJSON(original)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	again, err := enc.EncodeIR(decoded, usecases.ExportJSON)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if !bytes.Equal(original, again) {
		t.Errorf("round-trip is not byte-identical:\n--- first ---\n%s\n--- again ---\n%s",
			firstLines(string(original), 20), firstLines(string(again), 20))
	}
}

// TestSchemaVersionRefusal covers FR-036b and SC-011.
func TestSchemaVersionRefusal(t *testing.T) {
	t.Parallel()

	good, err := NewEncoder().EncodeIR(sampleIR(), usecases.ExportJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if _, err := DecodeIRJSON(good); err != nil {
		t.Errorf("a current-version artefact was refused: %v", err)
	}

	bumped := bytes.Replace(good, []byte(`"schemaVersion": 1`), []byte(`"schemaVersion": 2`), 1)
	_, err = DecodeIRJSON(bumped)
	if err == nil {
		t.Fatal("an unrecognised schema version was accepted")
	}
	// The message must name both versions, or the reader cannot tell whether
	// to upgrade the tool or regenerate the artefact.
	for _, want := range []string{"2", "1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name version %q", err.Error(), want)
		}
	}
}

func TestEncodeNilIR(t *testing.T) {
	t.Parallel()
	if _, err := NewEncoder().EncodeIR(nil, usecases.ExportJSON); err == nil {
		t.Error("encoding a nil IR returned no error")
	}
}

// TestEmptyArchitectureEmitsEmptyArrays: an empty project must not emit null
// collections, or every consumer needs a special case for it.
func TestEmptyArchitectureEmitsEmptyArrays(t *testing.T) {
	t.Parallel()

	ir := arch.NewIR(arch.Project{Name: "empty"}, nil, nil, nil, nil, nil)
	out, err := NewEncoder().EncodeIR(ir, usecases.ExportJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	for _, field := range []string{"elements", "relationships", "environments", "views", "ignores"} {
		if !strings.Contains(string(out), `"`+field+`": []`) {
			t.Errorf("%s is not an empty array:\n%s", field, out)
		}
	}
}

// toonHasKey reports whether TOON names a key. TOON has three shapes for one:
//
//	key: value                       a scalar
//	key[#3]:                         a list
//	key[#3]{col,col}:                a tabular array, where the field names
//	                                 appear once in the header instead of on
//	                                 every row — the token-efficiency win
//
// The third is why a naive "key:" check under-reports: the fields of a uniform
// array are hoisted into the header and never appear followed by a colon.
func toonHasKey(toonText, key string) bool {
	for _, form := range []string{key + ":", key + "[", key + ",", key + "}"} {
		if strings.Contains(toonText, form) {
			return true
		}
	}
	return false
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
