package arch

// DocsVia records how a docs path resolved (feature 018). Empty means it
// was not resolved, as for a model built in memory.
type DocsVia string

const (
	DocsFile    DocsVia = "file"    // next to the declaring file
	DocsRoot    DocsVia = "root"    // only from the project root, the 1.1.0 rule
	DocsBoth    DocsVia = "both"    // different files at both; the file-relative one is used
	DocsMissing DocsVia = "missing" // at neither, or outside the root
)
