package authoring

import (
	"errors"
	"strings"
	"testing"
)

const h1 = "0000000000000000000000000000000000000000000000000000000000000001"
const h2 = "0000000000000000000000000000000000000000000000000000000000000002"

func TestRevisionToken(t *testing.T) {
	t.Parallel()
	a := NewRevision([]FileHash{{"b.loko.hcl", h2}, {"a.loko.hcl", h1}})
	b := NewRevision([]FileHash{{"a.loko.hcl", h1}, {"b.loko.hcl", h2}})
	if a.Token() != b.Token() || len(a.Token()) != 19 {
		t.Fatalf("token %q must be short and independent of input order", a.Token())
	}
	if c := NewRevision([]FileHash{{"a.loko.hcl", h2}, {"b.loko.hcl", h2}}); c.Token() == a.Token() {
		t.Error("a changed hash must change the token")
	}
	if h, ok := a.Hash("b.loko.hcl"); !ok || h != h2 {
		t.Errorf("Hash(b) = %q, %v", h, ok)
	}
	if _, ok := a.Hash("c.loko.hcl"); ok {
		t.Error("Hash of an unknown file reported ok")
	}
	if NewRevision(nil).IsZero() || !(Revision{}).IsZero() {
		t.Error("IsZero: an empty project still has a revision; only the zero value is unset")
	}
}

func TestNewBatch(t *testing.T) {
	t.Parallel()
	rev := NewRevision([]FileHash{{"a.loko.hcl", h1}})
	e := Edit{Op: OpRemove, Target: TargetElement, Address: "system.s"}
	if _, err := NewBatch([]Edit{e}, false, rev); err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		n   int
		rev Revision
	}{"empty": {0, rev}, "too many": {MaxBatch + 1, rev}, "no revision": {1, Revision{}}} {
		edits := make([]Edit, tc.n)
		if _, err := NewBatch(edits, false, tc.rev); err == nil {
			t.Errorf("%s: NewBatch accepted it", name)
		}
	}
}

func TestPlanChanged(t *testing.T) {
	t.Parallel()
	same := Plan{Files: []FileContent{{Path: "a", Old: []byte("x"), New: []byte("x")}}}
	if same.Changed() {
		t.Error("identical content reported as changed")
	}
	created := Plan{Files: []FileContent{{Path: "a", New: []byte("")}}}
	if !created.Changed() {
		t.Error("a created file is a change")
	}
	var err error = &EditError{Index: 2, Reason: ReasonNotFound, Detail: "x"}
	var ee *EditError
	if !errors.As(err, &ee) || !strings.Contains(err.Error(), "edit 2: not_found") {
		t.Errorf("EditError: %v", err)
	}
}

func TestSplitAddress(t *testing.T) {
	t.Parallel()
	p, ok := SplitAddress(TargetInstance, "deployment.prod.node.vpc.subnet-a.instance.api")
	if !ok || p.Env != "prod" || p.Local != "api" || strings.Join(p.Groups, "/") != "vpc/subnet-a" {
		t.Errorf("nested instance: %+v %v", p, ok)
	}
	if got := CanonicalAddress(TargetInstance, "deployment.prod.node.vpc.instance.api"); got != "deployment.prod.instance.api" {
		t.Errorf("CanonicalAddress = %q", got)
	}
	_, err := NewEdit(Edit{Op: OpRemove, Target: TargetInstance, Address: "deployment.prod.node.vpc.instance.api"})
	if err == nil {
		t.Error("an existing instance must be addressed without its placement path")
	}
}
