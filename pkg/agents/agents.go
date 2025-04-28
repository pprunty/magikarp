
package agents

import (
    "embed"
    "encoding/json"
    "fmt"
    "io/fs"
    "path/filepath"
)

// Definition describes an agent composed of plugins.
type Definition struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Plugins     []string `json:"plugins"`
}

//go:embed */agent.json
var agentsFS embed.FS

// All returns the list of embedded agent definitions.
func All() ([]Definition, error) {
    var defs []Definition
    err := fs.WalkDir(agentsFS, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if filepath.Base(path) != "agent.json" {
            return nil
        }

        data, err := agentsFS.ReadFile(path)
        if err != nil {
            return err
        }
        var def Definition
        if err := json.Unmarshal(data, &def); err != nil {
            return fmt.Errorf("parse %s: %w", path, err)
        }
        if def.Plugins == nil {
            def.Plugins = []string{}
        }
        defs = append(defs, def)
        return nil
    })
    if err != nil {
        return nil, err
    }
    // move auto agent to front if present
    for i, d := range defs {
        if d.Name == "auto" && i != 0 {
            defs[0], defs[i] = defs[i], defs[0]
            break
        }
    }
    return defs, nil
}
