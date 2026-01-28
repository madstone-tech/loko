// Package markdown provides a Markdown documentation builder adapter.
// It implements the MarkdownBuilder interface by producing a single README.md
// file with complete architecture documentation.
package markdown

import (
	"context"
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// Builder implements the MarkdownBuilder interface by generating Markdown documentation.
// It produces a single README.md with complete architecture hierarchy.
type Builder struct{}

// NewBuilder creates a new Markdown builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildMarkdown generates a single README.md with complete architecture.
func (b *Builder) BuildMarkdown(ctx context.Context, project *entities.Project, systems []*entities.System) (string, error) {
	if project == nil {
		return "", fmt.Errorf("project cannot be nil")
	}

	var sb strings.Builder

	// Project header
	sb.WriteString(fmt.Sprintf("# %s\n\n", project.Name))

	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", project.Description))
	}

	if project.Version != "" {
		sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", project.Version))
	}

	// Table of contents
	sb.WriteString("## Table of Contents\n\n")
	for _, sys := range systems {
		if sys == nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("- [%s](#%s)\n", sys.Name, slugify(sys.Name)))
		for _, container := range sys.ListContainers() {
			if container == nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("  - [%s](#%s)\n", container.Name, slugify(container.Name)))
		}
	}
	sb.WriteString("\n---\n\n")

	// Systems
	for _, sys := range systems {
		if sys == nil {
			continue
		}
		content, err := b.BuildSystemMarkdown(ctx, sys, sys.ListContainers())
		if err != nil {
			return "", fmt.Errorf("failed to build system markdown for %s: %w", sys.Name, err)
		}
		sb.WriteString(content)
		sb.WriteString("\n---\n\n")
	}

	return sb.String(), nil
}

// BuildSystemMarkdown generates Markdown for a single system.
func (b *Builder) BuildSystemMarkdown(ctx context.Context, system *entities.System, containers []*entities.Container) (string, error) {
	if system == nil {
		return "", fmt.Errorf("system cannot be nil")
	}

	var sb strings.Builder

	// System header (Level 2)
	sb.WriteString(fmt.Sprintf("## %s\n\n", system.Name))

	if system.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", system.Description))
	}

	// System metadata
	if len(system.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(system.Tags, ", ")))
	}

	if len(system.Responsibilities) > 0 {
		sb.WriteString("**Responsibilities:**\n")
		for _, resp := range system.Responsibilities {
			sb.WriteString(fmt.Sprintf("- %s\n", resp))
		}
		sb.WriteString("\n")
	}

	if len(system.Dependencies) > 0 {
		sb.WriteString("**Dependencies:**\n")
		for _, dep := range system.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		sb.WriteString("\n")
	}

	// Containers
	if len(containers) > 0 {
		sb.WriteString("### Containers\n\n")

		for _, container := range containers {
			if container == nil {
				continue
			}
			content, err := b.buildContainerMarkdown(ctx, container)
			if err != nil {
				return "", fmt.Errorf("failed to build container markdown for %s: %w", container.Name, err)
			}
			sb.WriteString(content)
		}
	}

	return sb.String(), nil
}

// buildContainerMarkdown generates Markdown for a single container.
func (b *Builder) buildContainerMarkdown(_ context.Context, container *entities.Container) (string, error) {
	if container == nil {
		return "", fmt.Errorf("container cannot be nil")
	}

	var sb strings.Builder

	// Container header (Level 4)
	sb.WriteString(fmt.Sprintf("#### %s\n\n", container.Name))

	if container.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", container.Description))
	}

	// Container metadata
	if container.Technology != "" {
		sb.WriteString(fmt.Sprintf("**Technology:** %s\n\n", container.Technology))
	}

	if len(container.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(container.Tags, ", ")))
	}

	// Components
	components := container.ListComponents()
	if len(components) > 0 {
		sb.WriteString("**Components:**\n\n")
		sb.WriteString("| Name | Description | Technology |\n")
		sb.WriteString("|------|-------------|------------|\n")
		for _, comp := range components {
			if comp == nil {
				continue
			}
			desc := comp.Description
			if desc == "" {
				desc = "-"
			}
			tech := comp.Technology
			if tech == "" {
				tech = "-"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", comp.Name, desc, tech))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
