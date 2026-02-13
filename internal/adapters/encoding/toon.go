// Package encoding provides serialization adapters for loko.
// It implements OutputEncoder for JSON and TOON (Token-Optimized Object Notation) formats.
package encoding

import (
	"encoding/json"
	"fmt"

	toon "github.com/toon-format/toon-go"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Ensure Encoder implements usecases.OutputEncoder interface.
var _ usecases.OutputEncoder = (*Encoder)(nil)

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

// ArchitectureSummary is a TOON-optimized structure for architecture queries.
// It uses short field names and flat structure for maximum token efficiency.
type ArchitectureSummary struct {
	Name        string   `json:"n"        toon:"name"`
	Description string   `json:"d,omitempty" toon:"description,omitempty"`
	Version     string   `json:"v,omitempty" toon:"version,omitempty"`
	Systems     int      `json:"s"        toon:"systems"`
	Containers  int      `json:"c"        toon:"containers"`
	Components  int      `json:"k"        toon:"components"`
	SystemNames []string `json:"sn,omitempty" toon:"system_names,omitempty"`
}

// ArchitectureStructure is a TOON-optimized structure for structure-level queries.
type ArchitectureStructure struct {
	Name        string          `json:"n"            toon:"name"`
	Description string          `json:"d,omitempty" toon:"description,omitempty"`
	Systems     []SystemCompact `json:"s"            toon:"systems"`
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

// FormatArchitectureTOON creates a TOON-formatted string for architecture data.
// This is the primary function for token-efficient LLM responses.
// Compliant with TOON v3.0 specification.
func FormatArchitectureTOON(summary ArchitectureSummary) string {
	data, err := toon.Marshal(summary, toon.WithLengthMarkers(true))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(data)
}

// FormatStructureTOON creates a TOON-formatted string for structure data.
// Compliant with TOON v3.0 specification.
func FormatStructureTOON(structure ArchitectureStructure) string {
	data, err := toon.Marshal(structure, toon.WithLengthMarkers(true))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(data)
}
