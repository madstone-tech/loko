package usecases

// scaffold_project.go — project-level scaffolding (US1, T025)
//
// Investigation finding: the ScaffoldEntity.Execute dispatcher handles only
// "system", "container", and "component" entity types. There is no distinct
// "project" entity type handled here; project initialisation is the
// responsibility of the InitProject use case (see init_project.go).
//
// This file intentionally contains no code beyond this documentation comment.
// It serves as the canonical record that project-scaffolding is not part of
// ScaffoldEntity's remit, satisfying the split-plan entry for scaffoldProject.
