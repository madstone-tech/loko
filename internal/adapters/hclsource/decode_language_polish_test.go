package hclsource

import (
	"slices"
	"testing"
)

// TestDecodeNodeAndInstanceAttributes: feature 018's node and instance
// attributes reach the source model, each node with its own (nothing is
// inherited).
func TestDecodeNodeAndInstanceAttributes(t *testing.T) {
	t.Parallel()
	m, diags, err := New().Load(t.Context(), "testdata/golden/language_polish/input")
	if err != nil || len(diags) != 0 {
		t.Fatal(err, diags)
	}
	prod, dev := m.Environments[0], m.Environments[1]
	aws := prod.Groups[0]
	if aws.Title != "AWS account 1234" || aws.Technology != "AWS" || aws.Description != "" || aws.Tags != nil ||
		aws.AttrRanges["title"].StartLine != 12 {
		t.Errorf("aws: %+v", aws)
	}
	vpc := aws.Groups[0]
	if vpc.Title != "Production VPC" || vpc.Technology != "Amazon VPC" ||
		vpc.Description != "Private subnets in eu-west-1a and 1b" || !slices.Equal(vpc.Tags, []string{"network", "private"}) {
		t.Errorf("vpc: %+v", vpc)
	}
	if got := vpc.Instances[0].Technology; got != "AWS Lambda (arm64)" {
		t.Errorf("prod instance technology = %q", got)
	}
	if got := dev.Instances[0].Technology; got != "local process" {
		t.Errorf("dev instance technology = %q", got)
	}
	v := m.Views[0]
	if v.Title != "Phase 1: ingestion only" || v.Description != "What runs before the pricing engine exists." || v.TitleRange.StartLine != 37 {
		t.Errorf("view: %+v", v)
	}
}
