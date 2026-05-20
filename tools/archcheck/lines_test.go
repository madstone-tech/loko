package main

import (
	"testing"
)

func TestCountEffectiveLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "only blank lines",
			input: "\n\n\n",
			want:  0,
		},
		{
			name:  "single statement",
			input: "x := 1",
			want:  1,
		},
		{
			name:  "package declaration and statement",
			input: "package main\n\nx := 1\n",
			want:  1,
		},
		{
			name:  "single line comment dropped",
			input: "// this is a comment\nx := 1\n",
			want:  1,
		},
		{
			name:  "multi-line block comment dropped",
			input: "/*\nthis is\na block comment\n*/\nx := 1\n",
			want:  1,
		},
		{
			// "/* inline */" — awk sets flag then immediately fires the
			// in_block_comment && "*/" rule → next; line is dropped.
			name:  "single-line block comment dropped",
			input: "/* inline */\nx := 1\n",
			want:  1, // only x := 1 counts
		},
		{
			name:  "single-line import dropped",
			input: `import "fmt"` + "\nx := 1\n",
			want:  1,
		},
		{
			name:  "import block single entry dropped",
			input: "import (\n\t\"fmt\"\n)\nx := 1\n",
			want:  1,
		},
		{
			name:  "import block multi entry dropped",
			input: "import (\n\t\"fmt\"\n\t\"os\"\n\t\"strings\"\n)\nx := 1\ny := 2\n",
			want:  2,
		},
		{
			name:  "backtick string not treated as comment",
			input: "s := `hello\nworld`\n",
			// Note: backtick strings split across lines — each line is evaluated
			// independently. "s := `hello" counts, "world`" counts.
			want: 2,
		},
		{
			name: "mixed: package + imports + body + comments",
			input: `package main

import (
	"fmt"
	"os"
)

// doTheThing does a thing.
func doTheThing() {
	fmt.Println("hello")
	/* block
	   comment */
	os.Exit(0)
}
`,
			want: 4, // func decl, fmt.Println, os.Exit, closing brace
		},
		{
			// The shell awk script: when "/*" is seen it sets in_block_comment=1
			// (falls through), then on the SAME line "in_block_comment && */"
			// fires and does `next` — so the entire line is dropped, even if it
			// contains real code before the comment.
			name:  "inline block comment on code line is dropped (mirrors shell awk)",
			input: "x := 1 /* set x */\n",
			want:  0,
		},
		{
			name:  "package line with qualifier dropped",
			input: "package foo\n\nfunc F() {}\n",
			want:  1,
		},
		{
			name: "import with backtick path (single-line form) dropped",
			// single-line `import "..."` regex only matches double-quote form;
			// backtick imports in single-line form are unusual but we test the
			// import-block form instead
			input: "import (\n\t\"github.com/foo/bar\"\n)\nfunc F() {}\n",
			want:  1,
		},
		{
			name: "only package and imports — zero effective lines",
			input: `package main

import (
	"fmt"
)
`,
			want: 0,
		},
		{
			name: "function body with comments",
			input: `package main

// F is a function.
func F() {
	// do nothing
}
`,
			want: 2, // func F() { and }
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CountEffectiveLines(tc.input)
			if got != tc.want {
				t.Errorf("CountEffectiveLines(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
