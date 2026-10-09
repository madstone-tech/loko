package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// Apply validates, plans, compiles and commits a batch of edits, following the
// state flow in data-model §5. A refusal is a result, not an error: it changes
// nothing on disk and says why (SC-004). Writes are serialised (FR-017).
func (s *AuthoringService) Apply(ctx context.Context, req ApplyRequest) (*EditResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	edits, refusal := toEdits(req.Edits)
	if refusal != nil {
		return s.refuse(ctx, refusal)
	}
	base, ok := s.baseRevision(ctx, req.BaseRevision)
	if !ok {
		return s.refuse(ctx, &Refusal{Reason: authoring.ReasonStaleRevision,
			Detail: "base_revision is unknown or expired; read again (describe or validate) and retry"})
	}
	batch, err := authoring.NewBatch(edits, req.Preview, base)
	if err != nil {
		return s.refuse(ctx, &Refusal{Reason: authoring.ReasonInvalidEdit, Detail: err.Error()})
	}
	return s.applyBatch(ctx, batch)
}

// Move renames an element: one rename edit (FR-022).
func (s *AuthoringService) Move(ctx context.Context, from, to, base string, preview bool) (*EditResult, error) {
	return s.Apply(ctx, ApplyRequest{
		Edits:        []EditInput{{Op: string(authoring.OpRename), Target: string(authoring.TargetElement), Address: from, To: to}},
		BaseRevision: base,
		Preview:      preview,
	})
}

// baseRevision finds the revision a write is based on: one this server
// handed out, or, after a restart has emptied that memory, the current one if
// the token still matches the files as they are now.
func (s *AuthoringService) baseRevision(ctx context.Context, token string) (authoring.Revision, bool) {
	if rev, ok := s.deps.Revisions.lookup(token); ok {
		return rev, true
	}
	cur, err := s.deps.Editor.Revision(ctx, s.deps.Root)
	if err != nil || cur.Token() != token {
		return authoring.Revision{}, false
	}
	return cur, true
}

func (s *AuthoringService) refuse(ctx context.Context, r *Refusal) (*EditResult, error) {
	rev, err := currentRevision(ctx, s.deps)
	if err != nil {
		return nil, err
	}
	return &EditResult{Refusal: r, Revision: rev}, nil
}

// toEdits validates each input into an entity edit; the first failure is an
// invalid_edit refusal naming its index in the batch.
func toEdits(in []EditInput) ([]authoring.Edit, *Refusal) {
	out := make([]authoring.Edit, 0, len(in))
	for i, e := range in {
		set, err := toAttrs(authoring.TargetKind(e.Target), e.Address, e.Set)
		if err == nil {
			var ed authoring.Edit
			ed, err = authoring.NewEdit(authoring.Edit{
				Op: authoring.Op(e.Op), Target: authoring.TargetKind(e.Target), Address: e.Address,
				Binding: bindingRef(e.Binding), Set: set, Clear: e.Clear, Cascade: e.Cascade, To: e.To, File: e.File,
			})
			out = append(out, ed)
		}
		if err != nil {
			reason := authoring.ReasonInvalidEdit
			if fe, ok := errors.AsType[*authoring.FieldError](err); ok && fe.Field == "file" {
				reason = authoring.ReasonPathRefused
			}
			return nil, &Refusal{Reason: reason, Edit: i, Detail: err.Error()}
		}
	}
	return out, nil
}

func bindingRef(b *BindingInput) authoring.BindingRef {
	if b == nil {
		return authoring.BindingRef{}
	}
	return authoring.BindingRef{Kind: b.Kind, Index: b.Index}
}

// toAttrs converts decoded JSON values to typed attribute values. A string
// becomes a reference, and a list of strings a reference list, where the
// language expects one (system, container, target, of; a view's include and
// exclude); NewEdit then checks every value against the schema.
func toAttrs(t authoring.TargetKind, address string, set map[string]any) ([]authoring.Attr, error) {
	legal := authoring.LegalAttrs(t, address)
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	slices.Sort(names)
	out := make([]authoring.Attr, 0, len(set))
	for _, n := range names {
		v, err := toValue(set[n], legal[n])
		if err != nil {
			return nil, fmt.Errorf("set.%s: %w", n, err)
		}
		out = append(out, authoring.Attr{Name: n, Value: v})
	}
	return out, nil
}

func toValue(v any, want authoring.ValueKind) (authoring.AttrValue, error) {
	switch x := v.(type) {
	case string:
		if want == authoring.ValueRef {
			return authoring.AttrValue{Kind: authoring.ValueRef, Ref: x}, nil
		}
		return authoring.AttrValue{Kind: authoring.ValueString, Str: x}, nil
	case float64:
		return authoring.AttrValue{Kind: authoring.ValueNumber, Num: x}, nil
	case bool:
		return authoring.AttrValue{Kind: authoring.ValueBool, Bool: x}, nil
	case []any:
		list := make([]string, 0, len(x))
		for _, item := range x {
			s, ok := item.(string)
			if !ok {
				return authoring.AttrValue{}, fmt.Errorf("list items must be strings")
			}
			list = append(list, s)
		}
		if want == authoring.ValueRefList {
			return authoring.AttrValue{Kind: authoring.ValueRefList, List: list}, nil
		}
		return authoring.AttrValue{Kind: authoring.ValueList, List: list}, nil
	case map[string]any:
		return toMap(x)
	}
	return authoring.AttrValue{}, fmt.Errorf("unsupported value %v", v)
}

func toMap(m map[string]any) (authoring.AttrValue, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := authoring.AttrValue{Kind: authoring.ValueMap}
	for _, k := range keys {
		v, err := toValue(m[k], "")
		if err != nil {
			return out, fmt.Errorf("%s: %w", k, err)
		}
		out.Map = append(out.Map, authoring.MapEntry{Key: k, Value: v})
	}
	return out, nil
}
