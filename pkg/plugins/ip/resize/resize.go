package resize

import (
    _ "embed"
    "context"
    "encoding/json"
    "github.com/disintegration/imaging"
    "github.com/pprunty/magikarp/pkg/agent"
    "path/filepath"
    "strings"
)

//go:embed tool.json
var schema []byte

type input struct {
    Src        string `json:"src,omitempty"`
    ImagePath  string `json:"image_path,omitempty"`
    Dst        string `json:"dst,omitempty"`
    Width      int    `json:"width,omitempty"`
    Height     int    `json:"height,omitempty"`
}

func Definition() agent.ToolDefinition {
    // schema.json still contains the wrapper (name, description, input_schema)
    var file struct {
        Name        string                 `json:"name"`
        Description string                 `json:"description"`
        InputSchema map[string]any         `json:"input_schema"`
    }
    _ = json.Unmarshal(schema, &file)

    return agent.ToolDefinition{
        Name:        file.Name,            // "resize_image"
        Description: file.Description,
        InputSchema: file.InputSchema,     // <-- ONLY the parameters object
        Function:    run,
    }
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    var in input
    inputBytes, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("resize_image", err.Error(), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("resize_image", err.Error(), true), nil
    }

    // Get source path from either src or image_path
    srcPath := in.Src
    if srcPath == "" {
        srcPath = in.ImagePath
    }
    if srcPath == "" {
        return agent.NewToolResult("resize_image", "src or image_path is required", true), nil
    }

    // Set default dimensions if none specified
    width := in.Width
    height := in.Height
    if width == 0 && height == 0 {
        width = 512
        height = 512
    }

    // Handle relative paths
    if !filepath.IsAbs(srcPath) {
        srcPath = filepath.Join(".", srcPath)
    }

    // Generate default output filename if not provided
    dstPath := in.Dst
    if dstPath == "" {
        ext := filepath.Ext(srcPath)
        base := strings.TrimSuffix(filepath.Base(srcPath), ext)
        dstPath = filepath.Join(filepath.Dir(srcPath), base+"_resized"+ext)
    } else if !filepath.IsAbs(dstPath) {
        dstPath = filepath.Join(".", dstPath)
    }

    img, err := imaging.Open(srcPath)
    if err != nil {
        return agent.NewToolResult("resize_image", err.Error(), true), nil
    }

    resized := imaging.Resize(img, width, height, imaging.Lanczos)
    if err := imaging.Save(resized, dstPath); err != nil {
        return agent.NewToolResult("resize_image", err.Error(), true), nil
    }

    return agent.NewToolResult("resize_image", "image resized to "+dstPath, false), nil
}
