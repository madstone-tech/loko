package arch

// This file is the architecture language as data: every block, where it
// appears, and every attribute it accepts. Adding an attribute to the
// language starts here; the decoder derives its schema from this table and a
// parity test holds the edit schema to it.

var commonElementAttrs = []LanguageAttr{
	{Name: "description", Type: AttrString},
	{Name: "docs", Type: AttrString,
		Note: "markdown file, relative to the declaring file (docs/x.md next to architecture/x.loko.hcl is architecture/docs/x.md); the project root is a fallback"},
	{Name: "owner", Type: AttrString},
	{Name: "shape", Type: AttrString, Values: Shapes, Kinds: []string{string(KindContainer), string(KindExternal)}},
	{Name: "tags", Type: AttrStringList},
	{Name: "technology", Type: AttrString},
	{Name: "title", Type: AttrString, Note: "display title; the address keeps the name"},
}

var elementBlocks = []string{
	string(KindComponent), string(KindContainer), string(KindExternal), string(KindPerson), string(KindSystem),
}

// Language lists every block, sorted by name.
var Language = []LanguageBlock{
	{
		Name: "binding", Labels: []string{"kind"}, LabelValues: claimKindNames(), In: []string{"instance"},
		Attrs: []LanguageAttr{
			{Name: "address", Type: AttrString, Note: "one exact resource address; set address or addresses"},
			{Name: "addresses", Type: AttrStringList, Note: "several exact resource addresses"},
			{Name: "tags", Type: AttrMap, Note: "selects resources by tag"},
		},
		Example: "deployment \"prod\" {\n  instance \"api\" {\n    of = container.api\n\n" +
			"    binding \"terraform\" {\n      address = \"module.api.aws_lambda_function.this\"\n    }\n  }\n}\n",
	},
	{
		Name: string(KindComponent), Labels: []string{"name"}, In: []string{InFile},
		Attrs:   elementAttrs(LanguageAttr{Name: "container", Type: AttrRef, Required: true, Note: "the container it belongs to"}),
		Example: "component \"handler\" {\n  container = container.api\n}\n",
	},
	{
		Name: string(KindContainer), Labels: []string{"name"}, In: []string{InFile},
		Attrs:   elementAttrs(LanguageAttr{Name: "system", Type: AttrRef, Required: true, Note: "the system it belongs to"}),
		Example: "container \"api\" {\n  system = system.shop\n}\n",
	},
	{
		Name: "deployment", Labels: []string{"name"}, In: []string{InFile},
		Attrs: []LanguageAttr{
			{Name: "account", Type: AttrString},
			{Name: "provider", Type: AttrString},
			{Name: "region", Type: AttrString},
		},
		Example: "deployment \"prod\" {\n  provider = \"aws\"\n  region   = \"us-east-1\"\n}\n",
	},
	{
		Name: string(KindExternal), Labels: []string{"name"}, In: []string{InFile},
		Attrs:   elementAttrs(),
		Example: "external \"stripe\" {\n  description = \"Card payments, run by a vendor\"\n}\n",
	},
	{
		Name: "instance", Labels: []string{"name"}, In: []string{"deployment", "node"},
		Attrs: []LanguageAttr{
			{Name: "attributes", Type: AttrMap, Note: "environment-specific values such as sizing"},
			{Name: "of", Type: AttrRef, Required: true, Note: "the element this instance runs"},
			{Name: "planned", Type: AttrBool, Note: "infrastructure not built yet, so no binding is expected"},
			{Name: "technology", Type: AttrString, Note: "overrides the element's technology in this deployment"},
		},
		Example: "deployment \"prod\" {\n  instance \"api\" {\n    of = container.api\n  }\n}\n",
	},
	{
		Name: "locals", In: []string{InFile},
		Example: "locals {\n  team = \"platform\"\n}\n",
	},
	{
		Name: "moved", In: []string{InFile},
		Attrs: []LanguageAttr{
			{Name: "from", Type: AttrRef, Required: true, Note: "the old address"},
			{Name: "to", Type: AttrRef, Required: true, Note: "the new address"},
		},
		Example: "moved {\n  from = container.old_api\n  to   = container.api\n}\n",
	},
	{
		Name: "node", Labels: []string{"name"}, In: []string{"deployment", "node"},
		Attrs: []LanguageAttr{
			{Name: "description", Type: AttrString, Note: "listed on the deployment page, not drawn"},
			{Name: "tags", Type: AttrStringList, Note: "listed on the deployment page, not drawn"},
			{Name: "technology", Type: AttrString},
			{Name: "title", Type: AttrString, Note: "display title; the address keeps the name"},
		},
		Example: "deployment \"prod\" {\n  node \"vpc\" {\n    instance \"api\" {\n      of = container.api\n    }\n  }\n}\n",
	},
	{
		Name: string(KindPerson), Labels: []string{"name"}, In: []string{InFile},
		Attrs:   elementAttrs(),
		Example: "person \"customer\" {\n  description = \"Buys things\"\n}\n",
	},
	{
		Name: "project", Labels: []string{"name"}, In: []string{InFile},
		Attrs: []LanguageAttr{
			{Name: "description", Type: AttrString},
			{Name: "ignore_warnings", Type: AttrStringList, Note: ignoreWarningsNote()},
			{Name: "layout", Type: AttrString, Values: Layouts, Note: "layout engine for every diagram; dagre by default"},
			{Name: "loko_version", Type: AttrString, Note: "version constraint, for example \"~> 1.0\""},
		},
		Example: "project \"shop\" {\n  description = \"Online shop\"\n}\n",
	},
	{
		Name: "reconcile", In: []string{InFile},
		Attrs: []LanguageAttr{
			{Name: "ignore", Type: AttrStringList, Note: "resource address patterns reconciliation skips"},
		},
		Example: "reconcile {\n  ignore = [\"aws_iam_role_policy_attachment.*\"]\n}\n",
	},
	{
		Name: string(KindSystem), Labels: []string{"name"}, In: []string{InFile},
		Attrs:   elementAttrs(),
		Example: "system \"shop\" {\n  description = \"Storefront and orders\"\n}\n",
	},
	{
		Name: "uses", Labels: []string{"name"}, In: elementBlocks,
		Attrs: []LanguageAttr{
			{Name: "description", Type: AttrString},
			{Name: "kind", Type: AttrString, Values: RelationshipKinds, Note: "sync by default"},
			{Name: "tags", Type: AttrStringList},
			{Name: "target", Type: AttrRef, Required: true, Note: "the element depended on"},
			{Name: "technology", Type: AttrString},
		},
		Example: "container \"web\" {\n  system = system.shop\n\n  uses \"api\" {\n    target = container.api\n  }\n}\n",
	},
	{
		Name: "view", Labels: []string{"name"}, In: []string{InFile},
		Attrs: []LanguageAttr{
			{Name: "description", Type: AttrString, Note: "shown under the page heading, not drawn"},
			{Name: "direction", Type: AttrString, Values: Directions, Note: "down by default"},
			{Name: "exclude", Type: AttrRefList},
			{Name: "include", Type: AttrRefList},
			{Name: "layout", Type: AttrString, Values: Layouts, Note: "overrides the project's layout"},
			{Name: "tags", Type: AttrStringList},
			{Name: "title", Type: AttrString, Note: "page heading, navigation label and a heading on the diagram; the address keeps the name"},
		},
		Example: "view \"checkout\" {\n  include = [system.shop]\n}\n",
	},
}
