package hclsource

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

const sampleProject = `
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}

locals {
  team = "platform"
  tier = join("-", ["public", "edge"])
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = local.team
  docs        = "./docs/payments.md"
  tags        = ["pci"]
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"
  tags       = [local.tier, "pci"]

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "PostgreSQL wire protocol"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
}

component "handler" {
  container   = container.api
  description = "HTTP entry point"
}

deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    node "subnet-a" {
      instance "api" {
        of         = container.api
        attributes = { memory = 1024, timeout = 30 }

        binding "terraform" {
          address = "module.api.aws_lambda_function.this"
        }
      }
    }
  }
}

view "payment-path" {
  include = [system.payments, container.api]
  exclude = [container.orders_db]
  tags    = ["pci"]
}

reconcile {
  ignore = [
    "aws_iam_role_policy_attachment.*",
    "aws_cloudwatch_log_group.*",
  ]
}
`

func loadSample(t *testing.T, files map[string]string) (*arch.SourceModel, arch.Diagnostics) {
	t.Helper()
	root := writeTree(t, files)
	model, diags, err := New().Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return model, diags
}

func TestLoadFullProject(t *testing.T) {
	t.Parallel()

	model, diags := loadSample(t, map[string]string{"arch.loko.hcl": sampleProject})

	for _, d := range diags {
		if d.Severity == arch.SeverityError {
			t.Errorf("unexpected error: %s at %s — %s", d.Code, d.Range, d.Summary)
		}
	}

	if got, want := model.Project.Name, "acme-payments"; got != want {
		t.Errorf("project name = %q, want %q", got, want)
	}
	if got, want := model.Project.Version, "~> 1.0"; got != want {
		t.Errorf("loko_version = %q, want %q", got, want)
	}
	if got, want := len(model.Elements), 4; got != want {
		t.Fatalf("got %d elements, want %d", got, want)
	}

	byName := map[string]arch.ElementDecl{}
	for _, e := range model.Elements {
		byName[e.Name] = e
	}

	// Locals resolve, and one local feeds a function call feeding a list.
	if got, want := byName["payments"].Owner, "platform"; got != want {
		t.Errorf("owner from local = %q, want %q", got, want)
	}
	if got, want := byName["api"].Tags, []string{"public-edge", "pci"}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("tags = %v, want %v — join() through a local must evaluate", got, want)
	}

	// References are captured raw, unresolved, with a range.
	api := byName["api"]
	if got, want := api.Parent.Raw, "system.payments"; got != want {
		t.Errorf("container parent = %q, want %q", got, want)
	}
	if api.Parent.Range.StartLine == 0 {
		t.Error("parent reference has no source range")
	}
	if got, want := len(api.Relations), 1; got != want {
		t.Fatalf("got %d relations, want %d", got, want)
	}
	if got, want := api.Relations[0].Target.Raw, "container.orders_db"; got != want {
		t.Errorf("uses target = %q, want %q", got, want)
	}
	if got, want := api.Relations[0].LocalName, "orders"; got != want {
		t.Errorf("uses local name = %q, want %q", got, want)
	}
	if got, want := byName["handler"].Parent.Raw, "container.api"; got != want {
		t.Errorf("component parent = %q, want %q", got, want)
	}
}

func TestLoadDeploymentTree(t *testing.T) {
	t.Parallel()

	model, _ := loadSample(t, map[string]string{"arch.loko.hcl": sampleProject})

	if got, want := len(model.Environments), 1; got != want {
		t.Fatalf("got %d environments, want %d", got, want)
	}
	env := model.Environments[0]
	if env.Provider != "aws" || env.Account != "123456789012" || env.Region != "us-east-1" {
		t.Errorf("environment attrs wrong: %+v", env)
	}

	// AllInstances flattens the node tree and records the path.
	placed := env.AllInstances()
	if got, want := len(placed), 1; got != want {
		t.Fatalf("got %d instances, want %d", got, want)
	}
	inst := placed[0]
	if got, want := inst.Instance.Name, "api"; got != want {
		t.Errorf("instance name = %q, want %q", got, want)
	}
	if got, want := inst.Path, []string{"vpc-main", "subnet-a"}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("placement path = %v, want %v", got, want)
	}
	if got, want := inst.Instance.Of.Raw, "container.api"; got != want {
		t.Errorf("instance of = %q, want %q", got, want)
	}

	// Attributes are sorted by key at the boundary (FR-040).
	attrs := inst.Instance.Attributes
	if len(attrs) != 2 || attrs[0].Key != "memory" || attrs[1].Key != "timeout" {
		t.Errorf("attributes = %+v, want memory then timeout", attrs)
	}
	if got, want := attrs[0].Value.Num, 1024.0; got != want {
		t.Errorf("memory = %v, want %v", got, want)
	}

	if got, want := len(inst.Instance.Claims), 1; got != want {
		t.Fatalf("got %d claims, want %d", got, want)
	}
	claim := inst.Instance.Claims[0]
	if claim.Kind != arch.ClaimTerraform {
		t.Errorf("claim kind = %q, want terraform", claim.Kind)
	}
	if got, want := claim.Address, "module.api.aws_lambda_function.this"; got != want {
		t.Errorf("claim address = %q, want %q", got, want)
	}
}

func TestLoadViewsAndIgnores(t *testing.T) {
	t.Parallel()

	model, _ := loadSample(t, map[string]string{"arch.loko.hcl": sampleProject})

	if got, want := len(model.Views), 1; got != want {
		t.Fatalf("got %d views, want %d", got, want)
	}
	v := model.Views[0]
	if got, want := len(v.Include), 2; got != want {
		t.Errorf("view include = %d refs, want %d", got, want)
	}
	if got, want := v.Include[0].Raw, "system.payments"; got != want {
		t.Errorf("view include[0] = %q, want %q", got, want)
	}
	if got, want := len(v.Exclude), 1; got != want {
		t.Errorf("view exclude = %d refs, want %d", got, want)
	}

	if got, want := len(model.Ignores), 2; got != want {
		t.Fatalf("got %d ignore patterns, want %d", got, want)
	}
	if got, want := model.Ignores[0].Pattern, "aws_iam_role_policy_attachment.*"; got != want {
		t.Errorf("ignore[0] = %q, want %q", got, want)
	}
}

// TestLoadMergesAcrossFiles covers FR-001 and FR-019 together: an element
// declared in one file is referenced from another, and order does not matter.
func TestLoadMergesAcrossFiles(t *testing.T) {
	t.Parallel()

	model, diags := loadSample(t, map[string]string{
		"z_gateway.loko.hcl": `container "gateway" {
  system = system.payments
  uses "api" { target = container.api }
}`,
		"a_core.loko.hcl": `project "p" {}
system "payments" {}
container "api" { system = system.payments }`,
	})

	for _, d := range diags {
		if d.Severity == arch.SeverityError {
			t.Errorf("unexpected error: %s — %s", d.Code, d.Summary)
		}
	}
	if got, want := len(model.Elements), 3; got != want {
		t.Errorf("got %d elements, want %d", got, want)
	}
	if got, want := len(model.Files), 2; got != want {
		t.Errorf("got %d files, want %d", got, want)
	}
}

func TestUnknownConstructs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		code string
		// wantDetail, when set, must appear in the diagnostic's detail. This is
		// what distinguishes a message that teaches from one that merely
		// refuses — the whole point of naming the excluded constructs.
		wantDetail string
	}{
		{"unknown block", `widget "x" {}`, arch.CodeUnknownBlock, `"widget" is not a block`},
		{"for_each named", `for_each "x" {}`, arch.CodeUnknownBlock, "iteration is deliberately excluded"},
		{"dynamic named", `dynamic "x" {}`, arch.CodeUnknownBlock, "dynamic block generation is deliberately excluded"},
		{"variable named", `variable "x" {}`, arch.CodeUnknownBlock, "input variables are deliberately excluded"},
		{"module named", `module "x" {}`, arch.CodeUnknownBlock, "modules are deliberately excluded"},
		{"unknown attribute", `system "a" { colour = "red" }`, arch.CodeUnknownAttribute, "Supported attributes are"},
		{"unknown function", `system "a" { description = trimspace("  x  ") }`, arch.CodeUnknownFunction, "not a function in this language"},
		{"quoted reference", `container "a" { system = "system.payments" }`, arch.CodeWrongReferenceKind, "written without quotes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := loadSample(t, map[string]string{"arch.loko.hcl": tt.src})

			var match *arch.Diagnostic
			for i, d := range diags {
				if d.Code == tt.code {
					match = &diags[i]
					break
				}
			}
			if match == nil {
				var got []string
				for _, d := range diags {
					got = append(got, d.Code+": "+d.Summary)
				}
				t.Fatalf("want a %s diagnostic, got %v", tt.code, got)
			}
			if match.Range.StartLine < 1 {
				t.Errorf("%s has no source position", tt.code)
			}
			if tt.wantDetail != "" && !contains(match.Detail, tt.wantDetail) {
				t.Errorf("detail = %q, want it to contain %q", match.Detail, tt.wantDetail)
			}
		})
	}
}

// TestUnknownFunctionNamesTheFive is FR-017a: the message itself documents the
// available set, so the author does not have to find the manual.
func TestUnknownFunctionNamesTheFive(t *testing.T) {
	t.Parallel()

	_, diags := loadSample(t, map[string]string{
		"arch.loko.hcl": `system "a" { description = format("%s", "x") }`,
	})

	for _, d := range diags {
		if d.Code != arch.CodeUnknownFunction {
			continue
		}
		for _, fn := range []string{"join", "split", "lower", "upper", "replace"} {
			if !contains(d.Detail, fn) {
				t.Errorf("unknown_function detail does not name %q: %s", fn, d.Detail)
			}
		}
		return
	}
	t.Fatal("no unknown_function diagnostic")
}

func TestDuplicateProjectBlock(t *testing.T) {
	t.Parallel()

	_, diags := loadSample(t, map[string]string{
		"a.loko.hcl": `project "one" {}`,
		"b.loko.hcl": `project "two" {}`,
	})

	for _, d := range diags {
		if d.Code == arch.CodeDuplicateDeclaration {
			if len(d.Related) == 0 {
				t.Error("duplicate project diagnostic does not name the first declaration")
			}
			return
		}
	}
	t.Fatal("two project blocks produced no duplicate_declaration diagnostic")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
