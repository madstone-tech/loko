// Package outputdir implements the ArtifactStore port: it commits an artifact
// set to an output directory through a staging directory, skips byte-identical
// files, and prunes only files a previous build recorded in .loko-manifest
// (FR-022, FR-023). Which files to write and prune is decided by
// viewmodel.PlanCommit; this package performs the I/O.
package outputdir
