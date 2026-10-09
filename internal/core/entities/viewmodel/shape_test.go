package viewmodel

import "testing"

func TestWithShape(t *testing.T) {
	t.Parallel()
	base := StyleFor(RoleElement, KindContainer, nil)
	for shape, want := range map[string]Shape{"database": ShapeDatabase, "queue": ShapeQueue, "topic": ShapeTopic,
		"function": ShapeFunction, "bucket": ShapeBucket} {
		got := WithShape(base, shape)
		if got.Shape != want || got.Fill != base.Fill || got.Stroke != base.Stroke || got.FontColor != base.FontColor {
			t.Errorf("WithShape(%s) = %+v, want shape %s with the kind's colours", shape, got, want)
		}
	}
	if got := WithShape(base, ""); got.Shape != base.Shape {
		t.Errorf("no shape keeps the kind's default, got %s", got.Shape)
	}
}
