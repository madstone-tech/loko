// Package encoding provides serialization adapters for loko.
// It implements OutputEncoder for JSON and TOON (Token-Optimized Object Notation) formats.
package encoding

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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
// TOON (Token-Optimized Object Notation) achieves 30-40% fewer tokens than JSON
// by using abbreviated keys and compact delimiters.
//
// Format rules:
//   - Objects: {k1:v1;k2:v2} (semicolon-delimited)
//   - Arrays: [v1;v2;v3] (semicolon-delimited)
//   - Strings: unquoted if simple alphanumeric, quoted otherwise
//   - Numbers: as-is
//   - Booleans: T/F
//   - Null: -
//
// Key abbreviations for architecture data:
//   - n=name, d=description, t=technology, tg=tags
//   - s=systems, c=containers, k=components
//   - id=id, v=version, e=external
func (e *Encoder) EncodeTOON(value any) ([]byte, error) {
	result := encodeTOONValue(reflect.ValueOf(value), 0)
	return []byte(result), nil
}

// DecodeTOON deserializes TOON format to a value.
// Note: For v0.1.0, TOON is primarily an output format for LLM consumption.
// Full decode support is planned for v0.2.0.
func (e *Encoder) DecodeTOON(data []byte, value any) error {
	// For now, fall back to JSON parsing if the data looks like JSON
	if len(data) > 0 && data[0] == '{' || data[0] == '[' {
		return json.Unmarshal(data, value)
	}
	return fmt.Errorf("TOON decode not fully implemented in v0.1.0")
}

// Key abbreviations for common architecture fields
var keyAbbreviations = map[string]string{
	"name":         "n",
	"description":  "d",
	"technology":   "t",
	"tags":         "tg",
	"systems":      "s",
	"containers":   "c",
	"components":   "k",
	"id":           "id",
	"version":      "v",
	"external":     "e",
	"path":         "p",
	"type":         "ty",
	"source":       "src",
	"target":       "tgt",
	"relationship": "rel",
}

// encodeTOONValue recursively encodes a value to TOON format.
func encodeTOONValue(v reflect.Value, depth int) string {
	// Handle nil/invalid
	if !v.IsValid() {
		return "-"
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "-"
		}
		return encodeTOONValue(v.Elem(), depth)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return "-"
		}
		// Use unquoted if simple alphanumeric
		if isSimpleString(s) {
			return s
		}
		// Quote and escape otherwise
		return fmt.Sprintf("%q", s)

	case reflect.Bool:
		if v.Bool() {
			return "T"
		}
		return "F"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return "[]"
		}
		var parts []string
		for i := 0; i < v.Len(); i++ {
			parts = append(parts, encodeTOONValue(v.Index(i), depth+1))
		}
		return "[" + strings.Join(parts, ";") + "]"

	case reflect.Map:
		if v.Len() == 0 {
			return "{}"
		}
		var parts []string
		iter := v.MapRange()
		for iter.Next() {
			key := abbreviateKey(fmt.Sprintf("%v", iter.Key().Interface()))
			val := encodeTOONValue(iter.Value(), depth+1)
			parts = append(parts, key+":"+val)
		}
		return "{" + strings.Join(parts, ";") + "}"

	case reflect.Struct:
		t := v.Type()
		var parts []string
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get the JSON tag name if present
			jsonTag := field.Tag.Get("json")
			name := field.Name
			if jsonTag != "" {
				tagParts := strings.Split(jsonTag, ",")
				if tagParts[0] != "" && tagParts[0] != "-" {
					name = tagParts[0]
				}
				// Skip omitempty fields that are empty
				if len(tagParts) > 1 && tagParts[1] == "omitempty" {
					if isEmptyValue(v.Field(i)) {
						continue
					}
				}
			}

			fieldVal := encodeTOONValue(v.Field(i), depth+1)
			// Skip empty values
			if fieldVal == "-" || fieldVal == "[]" || fieldVal == "{}" {
				continue
			}

			key := abbreviateKey(name)
			parts = append(parts, key+":"+fieldVal)
		}
		if len(parts) == 0 {
			return "{}"
		}
		return "{" + strings.Join(parts, ";") + "}"

	default:
		// Fallback to JSON for unknown types
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return "-"
		}
		return string(data)
	}
}

// isSimpleString checks if a string can be represented without quotes.
func isSimpleString(s string) bool {
	if len(s) == 0 || len(s) > 50 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

// abbreviateKey returns abbreviated key if available.
func abbreviateKey(key string) string {
	lower := strings.ToLower(key)
	if abbr, ok := keyAbbreviations[lower]; ok {
		return abbr
	}
	return lower
}

// isEmptyValue checks if a value is empty (for omitempty support).
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

// ArchitectureSummary is a TOON-optimized structure for architecture queries.
// It uses short field names and flat structure for maximum token efficiency.
type ArchitectureSummary struct {
	Name        string   `json:"n"`
	Description string   `json:"d,omitempty"`
	Version     string   `json:"v,omitempty"`
	Systems     int      `json:"s"`
	Containers  int      `json:"c"`
	Components  int      `json:"k"`
	SystemNames []string `json:"sn,omitempty"`
}

// ArchitectureStructure is a TOON-optimized structure for structure-level queries.
type ArchitectureStructure struct {
	Name        string          `json:"n"`
	Description string          `json:"d,omitempty"`
	Systems     []SystemCompact `json:"s"`
}

// SystemCompact is a compact system representation.
type SystemCompact struct {
	ID          string            `json:"id"`
	Name        string            `json:"n"`
	Description string            `json:"d,omitempty"`
	Containers  []ContainerBrief  `json:"c,omitempty"`
}

// ContainerBrief is a brief container representation.
type ContainerBrief struct {
	ID          string `json:"id"`
	Name        string `json:"n"`
	Technology  string `json:"t,omitempty"`
}

// FormatArchitectureTOON creates a TOON-formatted string for architecture data.
// This is the primary function for token-efficient LLM responses.
func FormatArchitectureTOON(summary ArchitectureSummary) string {
	var sb strings.Builder

	// Header line
	sb.WriteString(fmt.Sprintf("@%s", summary.Name))
	if summary.Description != "" {
		sb.WriteString(fmt.Sprintf(":%s", truncate(summary.Description, 60)))
	}
	sb.WriteString("\n")

	// Stats line (very compact)
	sb.WriteString(fmt.Sprintf("S%d/C%d/K%d", summary.Systems, summary.Containers, summary.Components))

	// System names if provided
	if len(summary.SystemNames) > 0 {
		sb.WriteString("\n")
		sb.WriteString(strings.Join(summary.SystemNames, ","))
	}

	return sb.String()
}

// FormatStructureTOON creates a TOON-formatted string for structure data.
func FormatStructureTOON(structure ArchitectureStructure) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("@%s\n", structure.Name))

	// Systems
	for _, sys := range structure.Systems {
		sb.WriteString(fmt.Sprintf("S:%s", sys.Name))
		if sys.Description != "" {
			sb.WriteString(fmt.Sprintf(":%s", truncate(sys.Description, 40)))
		}
		sb.WriteString("\n")

		// Containers
		for _, c := range sys.Containers {
			sb.WriteString(fmt.Sprintf("  C:%s", c.Name))
			if c.Technology != "" {
				sb.WriteString(fmt.Sprintf("[%s]", c.Technology))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
