package encoding

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestPinnedArtifactStillReads is SC-011. export_v1.json was produced by the
// release that introduced schemaVersion 1 and is committed unchanged.
//
// It exists so that a later change to the export's shape cannot pass unnoticed:
// if this test starts failing, either the change was unintended, or it was
// intended and schemaVersion must be incremented. Without a pinned artefact the
// compatibility promise is only a claim.
func TestPinnedArtifactStillReads(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/export_v1.json")
	if err != nil {
		t.Fatalf("reading the pinned artefact: %v", err)
	}

	ir, err := DecodeIRJSON(data)
	if err != nil {
		t.Fatalf("this build cannot read an artefact it produced at schemaVersion 1: %v", err)
	}
	if ir.SchemaVersion != 1 {
		t.Errorf("schemaVersion = %d, want 1", ir.SchemaVersion)
	}
	if len(ir.Elements) == 0 {
		t.Error("the pinned artefact decoded to an empty architecture")
	}

	// Re-encoding must reproduce the file byte for byte. A shape change that
	// still parses would slip past a decode-only check.
	again, err := EncodeIRJSON(ir)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if !bytes.Equal(data, again) {
		t.Errorf("the export shape has changed but schemaVersion is still 1.\n" +
			"Either revert the change, or increment SchemaVersion and regenerate " +
			"testdata/export_v1.json.")
	}
}

// TestUnrecognisedVersionIsRefused covers FR-036b.
func TestUnrecognisedVersionIsRefused(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/export_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	future := bytes.Replace(data, []byte(`"schemaVersion": 1`), []byte(`"schemaVersion": 99`), 1)

	_, err = DecodeIRJSON(future)
	if err == nil {
		t.Fatal("an artefact from an unrecognised schema version was accepted")
	}
	for _, want := range []string{"99", "1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name version %q — the reader cannot tell "+
				"whether to upgrade the tool or regenerate the artefact", err.Error(), want)
		}
	}
}
