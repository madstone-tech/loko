// Package outputdir implements the ArtifactStore port: it commits an artifact
// set to an output directory through a staging directory, skips byte-identical
// files, and prunes only files a previous build recorded in its ownership
// manifest (FR-022, FR-023). The manifest lives in the state directory beside
// the output directory (.dist.loko/manifest for dist), never inside it, so a
// published copy of the output carries no tool state. Which files to write and
// prune is decided by viewmodel.PlanCommit; this package performs the I/O.
package outputdir
