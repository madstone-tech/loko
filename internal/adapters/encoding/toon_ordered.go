package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	toon "github.com/toon-format/toon-go"
)

// jsonToOrdered converts encoded JSON into toon-go's ordered Object/Field
// representation, preserving the order the JSON was written in.
//
// TOON is derived from the JSON bytes rather than from the IR directly, which
// makes FR-036's "equivalent information" true by construction instead of by
// two hand-written mappings staying in step. Two earlier attempts each lost
// something: encoding the struct directly made toon-go reflect past
// arch.Value's custom MarshalJSON and choke on its internal tag, and decoding
// into map[string]any sorted the keys, burying schemaVersion in the middle of
// the document when FR-036b requires it be readable first.
//
// A plain json.Unmarshal into map[string]any cannot preserve order, so the
// token stream is walked instead.
func jsonToOrdered(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("reading JSON: %w", err)
	}
	return readValue(dec, tok)
}

func readValue(dec *json.Decoder, tok json.Token) (any, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return readObject(dec)
		case '[':
			return readArray(dec)
		default:
			return nil, fmt.Errorf("unexpected delimiter %v", t)
		}
	case json.Number:
		// Handed through as a Go number so TOON renders it unquoted.
		if i, intErr := t.Int64(); intErr == nil {
			return i, nil
		}
		f, err := t.Float64()
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		// string, bool, or nil.
		return tok, nil
	}
}

func readObject(dec *json.Decoder) (any, error) {
	obj := toon.Object{}
	for {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if d, isDelim := keyTok.(json.Delim); isDelim && d == '}' {
			return obj, nil
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("object key is %T, want string", keyTok)
		}

		valTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		val, err := readValue(dec, valTok)
		if err != nil {
			return nil, err
		}
		obj.Fields = append(obj.Fields, toon.Field{Key: key, Value: val})
	}
}

func readArray(dec *json.Decoder) (any, error) {
	items := []any{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("unterminated array")
		}
		if err != nil {
			return nil, err
		}
		if d, isDelim := tok.(json.Delim); isDelim && d == ']' {
			return items, nil
		}
		item, err := readValue(dec, tok)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
}
