package cmd

import (
	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/adapters/markdown"
	"github.com/madstone-tech/loko/internal/adapters/outputdir"
	"github.com/madstone-tech/loko/internal/adapters/projectfs"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// newBuildDeps is the composition root for `build` and `serve`: the one
// place concrete adapters are chosen for the render ports. Constitution
// v1.4.0 (Principles I and II) places wiring in main.go or cmd/.
//
// Backends is the format registry. The default --format set is derived from
// it, so registering a backend here is all adding a format takes.
func newBuildDeps() usecases.BuildDeps {
	return usecases.BuildDeps{
		Source: hclsource.New(),
		Prose:  newProseReader(),
		Theme:  projectfs.Theme{},
		Backends: []usecases.Backend{
			d2.NewSourceBackend(),
			d2.NewSVGBackend(),
			markdown.New(),
			html.New(),
		},
		Store: outputdir.New(),
	}
}

// newProseReader returns the ProseReader for build, serve and the projection
// goldens.
func newProseReader() usecases.ProseReader { return projectfs.Prose{} }
