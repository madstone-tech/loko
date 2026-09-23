// Package encoding provides serialization adapters for loko.
// It implements OutputEncoder for JSON and TOON (Token-Optimized Object Notation) formats.
package encoding

import (
	"encoding/json"
	"fmt"

	toon "github.com/toon-format/toon-go"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Ensure Encoder implements usecases.OutputEncoder interface.
var (
	_ usecases.OutputEncoder = (*Encoder)(nil)
	_ usecases.IREncoder     = (*Encoder)(nil)
)

// Encoder provides JSON and TOON encoding/decoding.
type Encoder struct{}

// NewEncoder creates a new Encoder instance.
func NewEncoder() *Encoder {
	return &Encoder{}
}

// EncodeJSON serializes a value to JSON bytes.
func (e *Encoder) EncodeJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

// DecodeJSON deserializes JSON bytes to a value.
func (e *Encoder) DecodeJSON(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

// EncodeTOON serializes a value to TOON format (token-efficient).
// TOON (Token-Optimized Object Notation) achieves reduced token usage for LLM consumption
// by using compact delimiters and abbreviated keys.
func (e *Encoder) EncodeTOON(value any) ([]byte, error) {
	return toon.Marshal(value, toon.WithLengthMarkers(true))
}

// DecodeTOON deserializes TOON format to a value.
// Fully compliant with TOON v3.0 specification.
func (e *Encoder) DecodeTOON(data []byte, value any) error {
	return toon.Unmarshal(data, value)
}

// SystemCompact is a compact system representation.
type SystemCompact struct {
	ID          string           `json:"id"          toon:"id"`
	Name        string           `json:"n"           toon:"name"`
	Description string           `json:"d,omitempty" toon:"description,omitempty"`
	Containers  []ContainerBrief `json:"c,omitempty" toon:"containers,omitempty"`
}

// ContainerBrief is a brief container representation.
type ContainerBrief struct {
	ID         string `json:"id"         toon:"id"`
	Name       string `json:"n"          toon:"name"`
	Technology string `json:"t,omitempty" toon:"technology,omitempty"`
}

// EncodeIR serialises a compiled IR in the requested format, implementing the
// usecases.IREncoder port.
//
// TOON is produced by encoding the same wire document the JSON path uses, so
// the two carry equivalent information by construction rather than by two
// hand-maintained mappings drifting apart (FR-036).
func (e *Encoder) EncodeIR(ir *arch.IR, format usecases.ExportFormat) ([]byte, error) {
	if format == usecases.ExportJSON {
		return EncodeIRJSON(ir)
	}
	if ir == nil {
		return nil, fmt.Errorf("encode: nil IR")
	}
	if err := arch.CheckSchemaVersion(ir.SchemaVersion); err != nil {
		return nil, err
	}

	// TOON is derived from the JSON encoding so the two carry equivalent
	// information by construction. See jsonToOrdered for why neither encoding
	// the struct nor decoding into a map was adequate.
	jsonBytes, err := EncodeIRJSON(ir)
	if err != nil {
		return nil, err
	}
	doc, err := jsonToOrdered(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("encode TOON: %w", err)
	}
	out, err := e.EncodeTOON(doc)
	if err != nil {
		return nil, fmt.Errorf("encode TOON: %w", err)
	}
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out, nil
}
