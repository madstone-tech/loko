package hclsource

import (
	"bytes"
	"slices"
	"strings"
	"unicode"
)

// meaningDiff is unifiedDiff with the whitespace classification of feature
// 023 (contract W-4). A line whose only change is whitespace outside string
// literals becomes a context line with its new text, and its new line number
// is returned. The diff is "" when no line changed in meaning.
func meaningDiff(path string, old, new []byte) (string, []int) {
	if old != nil && bytes.Equal(old, new) {
		return "", nil
	}
	ops, ws := classifyWhitespace(myers(splitLines(old), splitLines(new)))
	if !slices.ContainsFunc(ops, func(op diffOp) bool { return op.kind != ' ' }) {
		return "", ws
	}
	return renderDiff(path, old == nil, ops), ws
}

// classifyWhitespace re-diffs each maximal run of removed and added lines on
// copies without whitespace outside string literals. A pair that matches
// there becomes a context op with the new text.
func classifyWhitespace(ops []diffOp) ([]diffOp, []int) {
	out := make([]diffOp, 0, len(ops))
	var ws []int
	newLine := 0
	for i := 0; i < len(ops); {
		if ops[i].kind == ' ' {
			out = append(out, ops[i])
			newLine++
			i++
			continue
		}
		j := i
		var removed, added []string
		for ; j < len(ops) && ops[j].kind != ' '; j++ {
			if ops[j].kind == '-' {
				removed = append(removed, ops[j].line)
			} else {
				added = append(added, ops[j].line)
			}
		}
		run, lines := reclassifyRun(removed, added, newLine)
		out, ws = append(out, run...), append(ws, lines...)
		newLine += len(added)
		i = j
	}
	return out, ws
}

// reclassifyRun diffs one run of removed and added lines on their stripped
// forms. first is the number of new lines before the run.
func reclassifyRun(removed, added []string, first int) ([]diffOp, []int) {
	inner := myers(stripAll(removed), stripAll(added))
	out := make([]diffOp, 0, len(inner))
	var ws []int
	r, a := 0, 0
	for _, op := range inner {
		switch op.kind {
		case ' ':
			out = append(out, diffOp{' ', added[a]})
			if removed[r] != added[a] {
				ws = append(ws, first+a+1)
			}
			r, a = r+1, a+1
		case '-':
			out = append(out, diffOp{'-', removed[r]})
			r++
		default:
			out = append(out, diffOp{'+', added[a]})
			a++
		}
	}
	return out, ws
}

func stripAll(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = stripWhitespace(l)
	}
	return out
}

// stripWhitespace removes whitespace outside "…" string literals. A \" inside
// a literal does not end it.
func stripWhitespace(line string) string {
	var b strings.Builder
	inString, escaped := false, false
	for _, r := range line {
		switch {
		case inString && escaped:
			escaped = false
		case inString && r == '\\':
			escaped = true
		case r == '"':
			inString = !inString
		case !inString && unicode.IsSpace(r):
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
