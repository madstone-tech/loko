package main

import (
	"regexp"
	"strings"
)

var (
	reBlankLine        = regexp.MustCompile(`^\s*$`)
	rePackageLine      = regexp.MustCompile(`^\s*package\s+`)
	reSingleImport     = regexp.MustCompile(`^\s*import\s+"`)
	reLineComment      = regexp.MustCompile(`^\s*//`)
	reImportBlockOpen  = regexp.MustCompile(`^\s*import\s*\(`)
	reImportBlockClose = regexp.MustCompile(`^\s*\)`)
)

// CountEffectiveLines counts the non-blank, non-comment, non-import lines in
// text, mirroring the logic of scripts/audit-constitution.sh count_effective_lines.
//
// Dropped categories (in order of evaluation):
//  1. Blank lines (^\s*$)
//  2. Package declaration lines (^\s*package\s+)
//  3. Single-line imports (^\s*import\s+")
//  4. Line comments (^\s*//)
//  5. Import blocks: from "import (" to the matching ")" (inclusive)
//  6. Block comments: any line containing "/*" starts block-comment mode;
//     any line containing "*/" ends it; all lines in between (and the
//     start/end lines themselves) are dropped.
func CountEffectiveLines(text string) int {
	lines := strings.Split(text, "\n")
	count := 0
	inImportBlock := false
	inBlockComment := false

	for _, line := range lines {
		// Step 5a: import block tracking (must happen before other filters)
		if !inImportBlock && reImportBlockOpen.MatchString(line) {
			inImportBlock = true
			continue // drop the "import (" line
		}
		if inImportBlock {
			if reImportBlockClose.MatchString(line) {
				inImportBlock = false
			}
			continue // drop all lines inside (and the closing ")")
		}

		// Step 6: block comment tracking.
		// Shell awk evaluates rules in order for each line:
		//   rule1: /\/\*/  → in_block=1   (falls through to next rules)
		//   rule2: in_block && /\*\// → in_block=0; next  (drops the line)
		//   rule3: in_block → next  (drops interior lines AND the opener)
		//   default: { print }
		// Net effect: every line touching "/*" or inside a block is dropped.
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}
		if strings.Contains(line, "/*") {
			inBlockComment = true
			if strings.Contains(line, "*/") {
				// Same-line open+close (e.g. "/* foo */"): flag set then
				// immediately cleared; line dropped (rule2 next fires).
				inBlockComment = false
			}
			// rule3 (in_block) fires on this same line → drop it.
			continue
		}

		// Steps 1-4: simple single-line filters
		if reBlankLine.MatchString(line) {
			continue
		}
		if rePackageLine.MatchString(line) {
			continue
		}
		if reSingleImport.MatchString(line) {
			continue
		}
		if reLineComment.MatchString(line) {
			continue
		}

		count++
	}

	return count
}
