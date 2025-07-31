package utils

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func init() {
	if err := validateToolSchemas(); err != nil {
		// Fail build / runtime early
		panic(err)
	}
}

func validateToolSchemas() error {
	compiler := jsonschema.NewCompiler()

	meta, err := compiler.Compile("https://json-schema.org/draft/2020-12/schema")
	if err != nil {
		return fmt.Errorf("compile metaschema: %w", err)
	}

	// Walk internal/tools directory for tool.json
	root := "internal/tools"
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "tool.json" {
			if err := validateFile(meta, path); err != nil {
				return err
			}
		}
		return nil
	})
}

func validateFile(meta *jsonschema.Schema, path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var v any
	if err := json.Unmarshal(f, &v); err != nil {
		return fmt.Errorf("%s: invalid json: %w", path, err)
	}
	if err := meta.Validate(v); err != nil {
		return fmt.Errorf("%s: schema invalid: %w", path, err)
	}
	return nil
}
