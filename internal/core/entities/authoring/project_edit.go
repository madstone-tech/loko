package authoring

// checkProject holds the rules only the project target has: its name is the
// block's label, chosen once when it is added, and it lives in a file the
// caller names, because an empty project has no file to place it beside.
func checkProject(e Edit) error {
	if e.Target != TargetProject {
		return nil
	}
	var name *AttrValue
	for i := range e.Set {
		if e.Set[i].Name == "name" {
			name = &e.Set[i].Value
		}
	}
	switch {
	case e.Op == OpRemove:
		return fieldError("op", "a project always has exactly one declaration; it cannot be removed")
	case e.Op == OpUpdate && name != nil:
		return fieldError("set.name", "the project name is its label, chosen when it is added")
	case e.Op == OpAdd && e.File == "":
		return fieldError("file", "required when adding the project, for example architecture.loko.hcl")
	case e.Op == OpAdd && name != nil && !ValidName(name.Str):
		return fieldError("set.name", "%q is not a valid name: letters, digits, _ and -", name.Str)
	}
	return nil
}

// ProjectSyntax is the smallest valid project declaration.
const ProjectSyntax = "project \"name\" {\n  description = \"...\"\n}\n"

// BootstrapEdit is the apply_edit edit that creates the project declaration,
// with placeholders to fill in. describe offers it in an empty directory, and
// the no_project refusal repeats it.
const BootstrapEdit = `{"op":"add","target":"project","address":"project","file":"architecture.loko.hcl",` +
	`"set":{"name":"<name>","description":"<description>"}}`
