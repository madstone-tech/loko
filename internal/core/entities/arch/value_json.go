package arch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// MarshalJSON renders a Value as its natural JSON shape rather than as the
// tagged union it is internally: a string becomes "x", a number 1024, a list
// an array, a map an object.
//
// Without this, every instance attribute in the export would be a struct with
// six mostly-empty fields — technically lossless, unusable in practice, and
// not what contracts/ir.schema.json describes.
//
// Map entries are emitted in the slice's existing order, which the parser
// already sorted by key. Go's encoder would sort a real map's keys anyway, but
// relying on that would leave the TOON backend free to disagree.
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.Kind {
	case ValueNull:
		return []byte("null"), nil
	case ValueString:
		return json.Marshal(v.Str)
	case ValueBool:
		return json.Marshal(v.Bool)
	case ValueNumber:
		// Integral values are emitted without a trailing ".0" so that
		// memory = 1024 round-trips as it was written.
		if v.Num == float64(int64(v.Num)) {
			return json.Marshal(int64(v.Num))
		}
		return json.Marshal(v.Num)
	case ValueList:
		items := v.List
		if items == nil {
			items = []Value{}
		}
		return json.Marshal(items)
	case ValueMap:
		var buf []byte
		buf = append(buf, '{')
		for i, kv := range v.MapKV {
			if i > 0 {
				buf = append(buf, ',')
			}
			key, err := json.Marshal(kv.Key)
			if err != nil {
				return nil, err
			}
			val, err := json.Marshal(kv.Value)
			if err != nil {
				return nil, err
			}
			buf = append(buf, key...)
			buf = append(buf, ':')
			buf = append(buf, val...)
		}
		return append(buf, '}'), nil
	default:
		return nil, fmt.Errorf("value: unknown kind %d", v.Kind)
	}
}

// UnmarshalJSON reads a Value back from its natural JSON shape, so an exported
// artefact can be read again (FR-036b).
//
// Object keys come back through a decoder token stream rather than a map, so
// the original key order is preserved and a re-export is byte-identical to the
// original.
func (v *Value) UnmarshalJSON(data []byte) error {
	var decoded any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&decoded); err != nil {
		return err
	}
	parsed, err := fromAny(decoded)
	if err != nil {
		return err
	}
	*v = parsed
	return nil
}

func fromAny(x any) (Value, error) {
	switch t := x.(type) {
	case nil:
		return Value{Kind: ValueNull}, nil
	case string:
		return Value{Kind: ValueString, Str: t}, nil
	case bool:
		return Value{Kind: ValueBool, Bool: t}, nil
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return Value{}, err
		}
		return Value{Kind: ValueNumber, Num: f}, nil
	case []any:
		out := Value{Kind: ValueList}
		for _, item := range t {
			iv, err := fromAny(item)
			if err != nil {
				return Value{}, err
			}
			out.List = append(out.List, iv)
		}
		return out, nil
	case map[string]any:
		out := Value{Kind: ValueMap}
		for _, k := range sortedKeys(t) {
			mv, err := fromAny(t[k])
			if err != nil {
				return Value{}, err
			}
			out.MapKV = append(out.MapKV, KeyValue{Key: k, Value: mv})
		}
		return out, nil
	default:
		return Value{}, fmt.Errorf("value: unsupported JSON type %T", x)
	}
}

// sortedKeys returns a map's keys in byte order, so a decoded map re-exports
// in the same order it was written.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
