package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

type recordingProse struct{ read []string }

func (r *recordingProse) ReadProse(_ context.Context, _, docs string) (string, bool, error) {
	r.read = append(r.read, docs)
	return "# prose", true, nil
}

// TestProseReadByResolvedPath: the IR's docs is the path the source adapter
// resolved, relative to the root, so the site reads a file-relative path's
// document (feature 018, R9). Unresolved, the declared string stands.
func TestProseReadByResolvedPath(t *testing.T) {
	t.Parallel()
	near := elem(arch.KindSystem, "near", 1)
	near.Docs, near.DocsPath, near.DocsVia = "docs/near.md", "architecture/docs/near.md", arch.DocsFile
	lost := elem(arch.KindSystem, "lost", 2)
	lost.Docs, lost.DocsVia = "docs/lost.md", arch.DocsMissing
	model := &arch.SourceModel{Elements: []arch.ElementDecl{near, lost}}
	res, _ := ResolveModel(model)
	ir := BuildIR(model, res)
	if e, _ := ir.Element("system.near"); e.Docs != "architecture/docs/near.md" {
		t.Errorf("IR docs = %q, want the resolved path", e.Docs)
	}
	if e, _ := ir.Element("system.lost"); e.Docs != "docs/lost.md" {
		t.Errorf("IR docs of an unresolved path = %q, want the declared one", e.Docs)
	}
	reader := &recordingProse{}
	set, err := readProse(t.Context(), reader, "/root", ir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := set["architecture/docs/near.md"]; !ok {
		t.Errorf("prose read %v, want the resolved path", reader.read)
	}
}
