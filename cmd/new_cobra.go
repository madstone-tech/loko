package cmd

import (
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new C4 entity",
	Long:    "Create a new system, container, or component in the current project.",
	GroupID: "scaffolding",
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"system\tCreate a new system",
			"container\tCreate a new container",
			"component\tCreate a new component",
		}, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// new system
	newCmd.AddCommand(newSystemCmd)
	newSystemCmd.Flags().StringP("description", "d", "", "system description")
	newSystemCmd.Flags().String("technology", "", "technology stack")
	newSystemCmd.Flags().StringP("template", "t", "", "template override")

	// new system flag completion
	_ = newSystemCmd.RegisterFlagCompletionFunc("template", completeTemplates)

	// new container
	newCmd.AddCommand(newContainerCmd)
	newContainerCmd.Flags().StringP("description", "d", "", "container description")
	newContainerCmd.Flags().String("technology", "", "technology stack")
	newContainerCmd.Flags().String("parent", "", "parent system name (required)")
	newContainerCmd.Flags().StringP("template", "t", "", "template override")
	_ = newContainerCmd.MarkFlagRequired("parent")
	_ = newContainerCmd.RegisterFlagCompletionFunc("parent", completeParentSystems)
	_ = newContainerCmd.RegisterFlagCompletionFunc("template", completeTemplates)

	// new component
	newCmd.AddCommand(newComponentCmd)
	newComponentCmd.Flags().StringP("description", "d", "", "component description")
	newComponentCmd.Flags().String("technology", "", "technology stack")
	newComponentCmd.Flags().String("parent", "", "parent container name (required)")
	newComponentCmd.Flags().StringP("template", "t", "", "template override")
	_ = newComponentCmd.MarkFlagRequired("parent")
	_ = newComponentCmd.RegisterFlagCompletionFunc("parent", completeParentContainers)
	_ = newComponentCmd.RegisterFlagCompletionFunc("template", completeTemplates)
}

var newSystemCmd = &cobra.Command{
	Use:   "system <name>",
	Short: "Create a new system",
	Args:  cobra.ExactArgs(1),
	RunE:  runNewSystem,
}

var newContainerCmd = &cobra.Command{
	Use:   "container <name>",
	Short: "Create a new container",
	Args:  cobra.ExactArgs(1),
	RunE:  runNewContainer,
}

var newComponentCmd = &cobra.Command{
	Use:   "component <name>",
	Short: "Create a new component",
	Args:  cobra.ExactArgs(1),
	RunE:  runNewComponent,
}

func runNewSystem(cmd *cobra.Command, args []string) error {
	newCommand := NewNewCommand("system", args[0])
	newCommand.WithProjectRoot(ProjectRoot)

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		newCommand.WithDescription(desc)
	}
	if tech, _ := cmd.Flags().GetString("technology"); tech != "" {
		newCommand.WithTechnology(tech)
	}
	if tmpl, _ := cmd.Flags().GetString("template"); tmpl != "" {
		newCommand.WithTemplate(tmpl)
	}

	return newCommand.Execute(cmd.Context())
}

func runNewContainer(cmd *cobra.Command, args []string) error {
	newCommand := NewNewCommand("container", args[0])
	newCommand.WithProjectRoot(ProjectRoot)

	parent, _ := cmd.Flags().GetString("parent")
	newCommand.WithParent(parent)

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		newCommand.WithDescription(desc)
	}
	if tech, _ := cmd.Flags().GetString("technology"); tech != "" {
		newCommand.WithTechnology(tech)
	}
	if tmpl, _ := cmd.Flags().GetString("template"); tmpl != "" {
		newCommand.WithTemplate(tmpl)
	}

	return newCommand.Execute(cmd.Context())
}

func runNewComponent(cmd *cobra.Command, args []string) error {
	newCommand := NewNewCommand("component", args[0])
	newCommand.WithProjectRoot(ProjectRoot)

	parent, _ := cmd.Flags().GetString("parent")
	newCommand.WithParent(parent)

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		newCommand.WithDescription(desc)
	}
	if tech, _ := cmd.Flags().GetString("technology"); tech != "" {
		newCommand.WithTechnology(tech)
	}
	if tmpl, _ := cmd.Flags().GetString("template"); tmpl != "" {
		newCommand.WithTemplate(tmpl)
	}

	return newCommand.Execute(cmd.Context())
}

// completeTemplates returns available template names from the filesystem.
func completeTemplates(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	var templates []string
	seen := make(map[string]bool)

	// Check relative path.
	if entries, err := os.ReadDir(filepath.Join(".", "templates")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !seen[entry.Name()] {
				templates = append(templates, entry.Name())
				seen[entry.Name()] = true
			}
		}
	}

	// Check relative to executable.
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if entries, err := os.ReadDir(filepath.Join(exeDir, "..", "templates")); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !seen[entry.Name()] {
					templates = append(templates, entry.Name())
					seen[entry.Name()] = true
				}
			}
		}
	}

	if len(templates) == 0 {
		templates = []string{"standard-3layer"}
	}
	return templates, cobra.ShellCompDirectiveNoFileComp
}

// completeParentSystems returns system names from the current project for --parent completion on containers.
func completeParentSystems(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	repo := filesystem.NewProjectRepository()
	systems, err := repo.ListSystems(cmd.Context(), ProjectRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range systems {
		names = append(names, s.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeParentContainers returns container names from the current project for --parent completion on components.
func completeParentContainers(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	repo := filesystem.NewProjectRepository()
	systems, err := repo.ListSystems(cmd.Context(), ProjectRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, sys := range systems {
		for _, cont := range sys.ListContainers() {
			names = append(names, cont.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
