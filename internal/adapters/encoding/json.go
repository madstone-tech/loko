package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// EncodeIRJSON serialises the compiled IR as JSON conforming to
// contracts/ir.schema.json (FR-036).
//
// The IR already holds ordered slices and no maps, so this encoder sorts
// nothing: it walks the value as given. That is the point of ordering at
// construction — a second backend cannot disagree with the first about order
// (FR-040, research R6).
//
// SetEscapeHTML(false) matters for byte-stability of a different kind: Go's
// default would rewrite & < > inside descriptions as & and friends, which
// is valid JSON but makes a committed artefact needlessly unreadable and
// diff-noisy the first time someone writes "A & B" in a description.
func EncodeIRJSON(ir *arch.IR) ([]byte, error) {
	if ir == nil {
		return nil, fmt.Errorf("encode: nil IR")
	}
	if err := arch.CheckSchemaVersion(ir.SchemaVersion); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(irDoc(ir)); err != nil {
		return nil, fmt.Errorf("encode IR: %w", err)
	}
	return buf.Bytes(), nil
}

// DecodeIRJSON reads a previously exported artefact, refusing an unrecognised
// schema version outright rather than reading it best-effort (FR-036b).
//
// The version is checked BEFORE any other field is interpreted: a partial read
// of an incompatible shape is worse than a clear stop, because it fails later
// and somewhere less obvious.
func DecodeIRJSON(data []byte) (*arch.IR, error) {
	var probe struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("reading export: %w", err)
	}
	if err := arch.CheckSchemaVersion(probe.SchemaVersion); err != nil {
		return nil, err
	}

	var doc irWire
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("reading export: %w", err)
	}
	return arch.NewIR(doc.Project, doc.Elements, doc.Relationships,
		doc.Environments, doc.Views, doc.Ignores), nil
}

// irWire is the on-the-wire shape, shared by both encodings so they cannot
// drift apart. It gives the document a stable field order and keeps the
// collection fields non-nil.
//
// The `toon` tags duplicate the `json` ones because toon-go reads its own tag
// and would otherwise emit Go field names — SchemaVersion rather than
// schemaVersion — leaving consumers with two vocabularies for one contract.
type irWire struct {
	SchemaVersion int                 `json:"schemaVersion" toon:"schemaVersion"`
	Project       arch.Project        `json:"project"       toon:"project"`
	Elements      []arch.Element      `json:"elements"      toon:"elements"`
	Relationships []arch.Relationship `json:"relationships" toon:"relationships"`
	Environments  []arch.Environment  `json:"environments"  toon:"environments"`
	Views         []arch.View         `json:"views"         toon:"views"`
	Ignores       []string            `json:"ignores"       toon:"ignores"`
}

func irDoc(ir *arch.IR) irWire {
	return irWire{
		SchemaVersion: ir.SchemaVersion,
		Project:       ir.Project,
		// Non-nil so an empty architecture emits [] rather than null. A
		// consumer should not have to special-case the empty project, and the
		// two encodings must agree on what "nothing" looks like.
		Elements:      nonNilElements(ir.Elements),
		Relationships: nonNilRelationships(ir.Relationships),
		Environments:  nonNilEnvironments(ir.Environments),
		Views:         nonNilViews(ir.Views),
		Ignores:       nonNilStrings(ir.Ignores),
	}
}

func nonNilElements(in []arch.Element) []arch.Element {
	if in == nil {
		return []arch.Element{}
	}
	return in
}

func nonNilRelationships(in []arch.Relationship) []arch.Relationship {
	if in == nil {
		return []arch.Relationship{}
	}
	return in
}

func nonNilEnvironments(in []arch.Environment) []arch.Environment {
	if in == nil {
		return []arch.Environment{}
	}
	return in
}

func nonNilViews(in []arch.View) []arch.View {
	if in == nil {
		return []arch.View{}
	}
	return in
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
