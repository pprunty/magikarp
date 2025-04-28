package cli

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/manifoldco/promptui"
    "github.com/urfave/cli/v2"
)

func NewCreateAgentCmd() *cli.Command {
    return &cli.Command{
        Name:  "create-agent",
        Usage: "Create new agent definition",
        Flags: []cli.Flag{
            &cli.StringFlag{Name: "name", Usage: "agent name"},
        },
        Action: func(c *cli.Context) error {
            // Get agent name interactively if not provided via flag
            name := c.String("name")
            if name == "" {
                prompt := promptui.Prompt{
                    Label: "Agent name (use lowercase)",
                    Validate: func(input string) error {
                        if len(input) == 0 {
                            return fmt.Errorf("name cannot be empty")
                        }
                        if input != strings.ToLower(input) {
                            return fmt.Errorf("name must be lowercase")
                        }
                        // Check if agent already exists
                        if _, err := os.Stat(fmt.Sprintf("pkg/agents/%s", input)); err == nil {
                            return fmt.Errorf("agent '%s' already exists", input)
                        }
                        return nil
                    },
                }

                var err error
                name, err = prompt.Run()
                if err != nil {
                    return fmt.Errorf("prompt failed: %v", err)
                }
            }

            // Get available plugins by listing directories in pkg/plugins
            pluginsDir := "pkg/plugins"
            entries, err := os.ReadDir(pluginsDir)
            if err != nil {
                return fmt.Errorf("failed to read plugins directory: %v", err)
            }

            // Filter out 'all' which is a special case for imports
            var availablePlugins []string
            for _, entry := range entries {
                if entry.IsDir() && entry.Name() != "all" {
                    availablePlugins = append(availablePlugins, entry.Name())
                }
            }

            selectedPlugins := make(map[string]bool)
            
            // Initial state - all plugins are unselected
            pluginOptions := []string{"[ Done - Create Agent ]"}
            for _, plugin := range availablePlugins {
                pluginOptions = append(pluginOptions, "[ ] "+plugin)
            }

            fmt.Println("Select plugins for your agent:")
            fmt.Println("Use ↑/↓ to navigate, space to toggle selection, enter when done.")

            for {
                prompt := promptui.Select{
                    Label: "Available plugins",
                    Items: pluginOptions,
                    Size:  len(pluginOptions),
                }

                idx, _, err := prompt.Run()
                if err != nil {
                    return fmt.Errorf("plugin selection failed: %v", err)
                }

                if idx == 0 {
                    // User selected "Done"
                    break
                }

                // Toggle selection
                pluginIdx := idx - 1
                plugin := availablePlugins[pluginIdx]
                
                if selectedPlugins[plugin] {
                    // Unselect
                    delete(selectedPlugins, plugin)
                    pluginOptions[idx] = "[ ] " + plugin
                } else {
                    // Select
                    selectedPlugins[plugin] = true
                    pluginOptions[idx] = "[✓] " + plugin
                }
            }

            // Convert selected plugins map to slice
            plugins := make([]string, 0, len(selectedPlugins))
            for plugin := range selectedPlugins {
                plugins = append(plugins, plugin)
            }

            // Create agent definition
            var description string
            if len(plugins) == 0 {
                description = "Custom agent with no plugins"
            } else {
                description = "Custom agent with " + strings.Join(plugins, ", ") + " plugins"
            }
            
            def := map[string]any{
                "name":        name,
                "description": description,
                "plugins":     plugins,
            }

            // Write to file
            data, _ := json.MarshalIndent(def, "", "  ")
            dir := fmt.Sprintf("pkg/agents/%s", name)
            if err := os.MkdirAll(dir, 0755); err != nil {
                return err
            }

            agentFile := filepath.Join(dir, "agent.json")
            if err := os.WriteFile(agentFile, data, 0644); err != nil {
                return err
            }

            if len(plugins) == 0 {
                fmt.Printf("Created agent '%s' with no plugins\n", name)
            } else {
                fmt.Printf("Created agent '%s' with plugins: %s\n", name, strings.Join(plugins, ", "))
            }
            return nil
        },
    }
}
