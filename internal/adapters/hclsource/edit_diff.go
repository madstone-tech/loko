package hclsource

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// diffContext is the number of unchanged lines shown around each change.
const diffContext = 3

// Diff renders a unified diff for every changed file in the plan, sorted by
// path (research R9).
func (e *Editor) Diff(plan authoring.Plan) []authoring.FileChange {
	var out []authoring.FileChange
	for _, f := range plan.Files {
		if d := unifiedDiff(f.Path, f.Old, f.New); d != "" {
			out = append(out, authoring.FileChange{Path: f.Path, Diff: d})
		}
	}
	slices.SortFunc(out, func(a, b authoring.FileChange) int { return strings.Compare(a.Path, b.Path) })
	return out
}

type diffOp struct {
	kind byte // ' ', '-', '+'
	line string
}

// unifiedDiff is a line-based unified diff of old and new. old nil means the
// file is new. Identical inputs give "".
func unifiedDiff(path string, old, new []byte) string {
	if old != nil && bytes.Equal(old, new) {
		return ""
	}
	ops := myers(splitLines(old), splitLines(new))
	var b strings.Builder
	from := "a/" + path
	if old == nil {
		from = "/dev/null"
	}
	fmt.Fprintf(&b, "--- %s\n+++ b/%s\n", from, path)
	for _, h := range hunks(ops) {
		writeHunk(&b, ops, h)
	}
	return b.String()
}

func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	ls := strings.SplitAfter(string(b), "\n")
	if ls[len(ls)-1] == "" {
		ls = ls[:len(ls)-1]
	}
	return ls
}

// myers computes a shortest edit script (Myers 1986), O((N+M)·D).
func myers(a, b []string) []diffOp {
	n, m := len(a), len(b)
	max := n + m
	v := make([]int, 2*max+2)
	var trace [][]int
	for d := 0; d <= max; d++ {
		trace = append(trace, slices.Clone(v))
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[max+k-1] < v[max+k+1]) {
				x = v[max+k+1]
			} else {
				x = v[max+k-1] + 1
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x, y = x+1, y+1
			}
			v[max+k] = x
			if x >= n && y >= m {
				return backtrack(a, b, trace, d, max)
			}
		}
	}
	return nil
}

func backtrack(a, b []string, trace [][]int, d, max int) []diffOp {
	var ops []diffOp
	x, y := len(a), len(b)
	for ; d > 0; d-- {
		v := trace[d]
		k := x - y
		var pk int
		if k == -d || (k != d && v[max+k-1] < v[max+k+1]) {
			pk = k + 1
		} else {
			pk = k - 1
		}
		px := v[max+pk]
		py := px - pk
		for x > px && y > py {
			x, y = x-1, y-1
			ops = append(ops, diffOp{' ', a[x]})
		}
		if x == px {
			y--
			ops = append(ops, diffOp{'+', b[y]})
		} else {
			x--
			ops = append(ops, diffOp{'-', a[x]})
		}
	}
	for x > 0 {
		x--
		ops = append(ops, diffOp{' ', a[x]})
	}
	slices.Reverse(ops)
	return ops
}

// hunks groups changes with their context; changes separated by no more than
// twice the context share a hunk.
func hunks(ops []diffOp) [][2]int {
	var out [][2]int
	for i := 0; i < len(ops); i++ {
		if ops[i].kind == ' ' {
			continue
		}
		start := max(0, i-diffContext)
		end := i
		for j := i; j < len(ops); j++ {
			if ops[j].kind != ' ' {
				end = j
			} else if j-end > 2*diffContext {
				break
			}
		}
		end = min(len(ops), end+1+diffContext)
		if n := len(out); n > 0 && start <= out[n-1][1] {
			out[n-1][1] = end
		} else {
			out = append(out, [2]int{start, end})
		}
		i = end - 1
	}
	return out
}

func writeHunk(b *strings.Builder, ops []diffOp, h [2]int) {
	oldStart, newStart := 1, 1
	for _, op := range ops[:h[0]] {
		if op.kind != '+' {
			oldStart++
		}
		if op.kind != '-' {
			newStart++
		}
	}
	var oldN, newN int
	for _, op := range ops[h[0]:h[1]] {
		if op.kind != '+' {
			oldN++
		}
		if op.kind != '-' {
			newN++
		}
	}
	fmt.Fprintf(b, "@@ -%s +%s @@\n", hunkRange(oldStart, oldN), hunkRange(newStart, newN))
	for _, op := range ops[h[0]:h[1]] {
		b.WriteByte(op.kind)
		b.WriteString(op.line)
		if !strings.HasSuffix(op.line, "\n") {
			b.WriteString("\n\\ No newline at end of file\n")
		}
	}
}

// hunkRange follows GNU diff: a count of 1 is omitted, and an empty range
// names the line before it.
func hunkRange(start, n int) string {
	switch n {
	case 0:
		return fmt.Sprintf("%d,0", start-1)
	case 1:
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d,%d", start, n)
}
