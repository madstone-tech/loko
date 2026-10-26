package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// This file holds the argument decoding and result encoding every tool
// shares, so each handler stays a thin decode → use case → encode adapter
// (Principle III). It is a helper file, exempt from the per-handler budget.

// argError is a malformed request: a JSON-RPC error, not a refusal.
func argError(key, format string, args ...any) error {
	return fmt.Errorf("argument %q: %s", key, fmt.Sprintf(format, args...))
}

// str returns a required string argument.
func str(args map[string]any, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", argError(key, "required")
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", argError(key, "must be a non-empty string")
	}
	return s, nil
}

// optStr returns an optional string argument, or def when absent.
func optStr(args map[string]any, key, def string) (string, error) {
	if _, ok := args[key]; !ok {
		return def, nil
	}
	return str(args, key)
}

// boolArg returns an optional boolean argument, false when absent.
func boolArg(args map[string]any, key string) (bool, error) {
	v, ok := args[key]
	if !ok {
		return false, nil
	}
	b, ok := v.(bool)
	if !ok {
		return false, argError(key, "must be a boolean")
	}
	return b, nil
}

// intArg returns an optional non-negative integer argument, 0 when absent.
// JSON numbers arrive as float64.
func intArg(args map[string]any, key string) (int, error) {
	v, ok := args[key]
	if !ok {
		return 0, nil
	}
	f, ok := v.(float64)
	if !ok || f < 0 || f != float64(int(f)) {
		return 0, argError(key, "must be a non-negative integer")
	}
	return int(f), nil
}

// formatArg returns the output format, defaulting to TOON (FR-008, ADR-0011).
func formatArg(args map[string]any) (string, error) {
	f, err := optStr(args, "format", "toon")
	if err != nil {
		return "", err
	}
	if !slices.Contains([]string{"toon", "json"}, f) {
		return "", argError("format", "must be toon or json")
	}
	return f, nil
}

// decodeEdits decodes the edits array strictly: an unknown field is a typo the
// caller should hear about, not silently drop.
func decodeEdits(args map[string]any) ([]usecases.EditInput, error) {
	raw, ok := args["edits"].([]any)
	if !ok || len(raw) == 0 {
		return nil, argError("edits", "must be a non-empty array of edits")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, argError("edits", "%v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var edits []usecases.EditInput
	if err := dec.Decode(&edits); err != nil {
		return nil, argError("edits", "%v", err)
	}
	return edits, nil
}

// encodeResult renders a use-case result as the tool's text content. The
// encoder is injected by cmd/wiring.go; tools never construct an adapter.
func encodeResult(enc usecases.OutputEncoder, v any, format string) (any, error) {
	var (
		b   []byte
		err error
	)
	if format == "json" {
		b, err = enc.EncodeJSON(v)
	} else {
		b, err = enc.EncodeTOON(v)
	}
	if err != nil {
		return nil, fmt.Errorf("encoding result: %w", err)
	}
	// A string result is sent verbatim by the server rather than JSON-quoted.
	return string(b), nil
}

// queryRequest decodes the query tool's arguments.
func queryRequest(args map[string]any) (usecases.QueryRequest, error) {
	var (
		req usecases.QueryRequest
		err error
	)
	if req.Kind, err = str(args, "kind"); err != nil {
		return req, err
	}
	if req.Address, err = optStr(args, "address", ""); err != nil {
		return req, err
	}
	if req.To, err = optStr(args, "to", ""); err != nil {
		return req, err
	}
	if req.Transitive, err = boolArg(args, "transitive"); err != nil {
		return req, err
	}
	req.Limit, err = intArg(args, "limit")
	return req, err
}

// fullRanges reads the ranges argument: "line" (the default) gives each
// diagnostic as file:line, "full" its whole source range (feature 017).
func fullRanges(args map[string]any) (bool, error) {
	r, err := optStr(args, "ranges", "line")
	if err != nil {
		return false, err
	}
	if r != "line" && r != "full" {
		return false, argError("ranges", "must be line or full")
	}
	return r == "full", nil
}

// strList returns an optional list-of-strings argument.
func strList(args map[string]any, key string) ([]string, error) {
	v, ok := args[key]
	if !ok {
		return nil, nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil, argError(key, "must be a list of strings")
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, argError(key, "must be a list of strings")
		}
		out = append(out, s)
	}
	return out, nil
}

// shaper is a result whose diagnostics are shaped before encoding.
type shaper interface{ ShapeDiagnostics(full bool) }

// encodeShaped shapes a result's diagnostics as the ranges argument asks and
// encodes it in the requested format.
func encodeShaped(enc usecases.OutputEncoder, res shaper, args map[string]any) (any, error) {
	format, err := formatArg(args)
	if err != nil {
		return nil, err
	}
	full, err := fullRanges(args)
	if err != nil {
		return nil, err
	}
	res.ShapeDiagnostics(full)
	return encodeResult(enc, res, format)
}

// validateRequest decodes the validate tool's options.
func validateRequest(args map[string]any) (usecases.ValidateRequest, error) {
	var (
		req usecases.ValidateRequest
		err error
	)
	if req.Summary, err = boolArg(args, "summary"); err != nil {
		return req, err
	}
	if req.Severity, err = optStr(args, "severity", ""); err != nil {
		return req, err
	}
	if req.Ignore, err = strList(args, "ignore"); err != nil {
		return req, err
	}
	req.FullRanges, err = fullRanges(args)
	return req, err
}
