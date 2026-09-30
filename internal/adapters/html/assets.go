package html

import "embed"

// builtinTheme holds the default presentation. It lives in files rather than
// Go string constants, so presentation is edited as HTML and CSS and the Go
// files stay within the adapter size budget (research R11).
//
//go:embed theme/*
var builtinTheme embed.FS
