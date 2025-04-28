package convert

import (
    _ "embed"
    "context"
    "encoding/json"
    "fmt"
    "image"
    "image/jpeg"
    "image/png"
    "os"
    "path/filepath"
    "strings"

    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

type input struct {
    Src string `json:"src"`
    Dst string `json:"dst"`
}

func Definition() agent.ToolDefinition {
    var sch map[string]interface{}
    _ = json.Unmarshal(schema, &sch)
    return agent.ToolDefinition{
        Name:        "convert_image",
        Description: "Convert an image between PNG and JPEG",
        InputSchema: sch,
        Function:    run,
    }
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    var in input
    inputBytes, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("convert_image", err.Error(), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("convert_image", err.Error(), true), nil
    }

    file, err := os.Open(in.Src)
    if err != nil {
        return agent.NewToolResult("convert_image", err.Error(), true), nil
    }
    defer file.Close()

    img, _, err := image.Decode(file)
    if err != nil {
        return agent.NewToolResult("convert_image", err.Error(), true), nil
    }

    out, err := os.Create(in.Dst)
    if err != nil {
        return agent.NewToolResult("convert_image", err.Error(), true), nil
    }
    defer out.Close()

    switch strings.ToLower(filepath.Ext(in.Dst)) {
    case ".png":
        err = png.Encode(out, img)
    case ".jpg", ".jpeg":
        err = jpeg.Encode(out, img, &jpeg.Options{Quality: 92})
    default:
        err = fmt.Errorf("unsupported destination format")
    }
    if err != nil {
        return agent.NewToolResult("convert_image", err.Error(), true), nil
    }
    return agent.NewToolResult("convert_image", "image converted", false), nil
}
