package convert

import (
    _ "embed"
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

// Input represents the parameters for the convert_image tool
type input struct {
    Src     string `json:"src"`
    Dst     string `json:"dst"`
    Quality int    `json:"quality,omitempty"`
    Resize  string `json:"resize,omitempty"`
}

// Definition returns the tool definition for the convert_image tool
func Definition() agent.ToolDefinition {
    var sch map[string]interface{}
    if err := json.Unmarshal(schema, &sch); err != nil {
        fmt.Printf("Error unmarshaling schema: %v\n", err)
    }
    return agent.ToolDefinition{
        Name:        "convert_image",
        Description: "Convert an image between various formats including PNG, JPEG, WEBP, HEIC, TIFF, BMP, and GIF",
        InputSchema: sch,
        Function:    run,
    }
}

// run executes the image conversion
func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    // Parse input parameters
    var in input
    inputBytes, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("convert_image", fmt.Sprintf("Error processing input: %v", err), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("convert_image", fmt.Sprintf("Error parsing input: %v", err), true), nil
    }

    // Validate input
    if in.Src == "" || in.Dst == "" {
        return agent.NewToolResult("convert_image", "Source and destination paths are required", true), nil
    }

    // Ensure paths are valid
    if !filepath.IsAbs(in.Src) {
        in.Src = filepath.Clean(in.Src)
    }
    if !filepath.IsAbs(in.Dst) {
        in.Dst = filepath.Clean(in.Dst)
    }

    // Set default quality if not specified
    if in.Quality <= 0 {
        in.Quality = 95 // High quality default
    }

    // Create destination directory if it doesn't exist
    dstDir := filepath.Dir(in.Dst)
    if err := os.MkdirAll(dstDir, 0755); err != nil {
        return agent.NewToolResult("convert_image", fmt.Sprintf("Failed to create destination directory: %v", err), true), nil
    }

    // Check if ImageMagick is installed
    _, err = exec.LookPath("convert")
    if err != nil {
        _, err = exec.LookPath("magick")
        if err != nil {
            return agent.NewToolResult("convert_image", "ImageMagick is not installed (required for image conversion)", true), nil
        }
    }

    // Build the ImageMagick command
    var args []string

    // Check which command is available (newer versions use "magick convert", older just "convert")
    magickCmd := "convert"
    if _, err := exec.LookPath("magick"); err == nil {
        magickCmd = "magick"
        args = append(args, "convert")
    }

    // Add source file
    args = append(args, in.Src)

    // Add quality parameter for lossy formats
    if in.Quality > 0 {
        args = append(args, "-quality", fmt.Sprintf("%d", in.Quality))
    }

    // Add resize parameter if specified
    if in.Resize != "" {
        args = append(args, "-resize", in.Resize)
    }

    // Add destination file
    args = append(args, in.Dst)

    // Execute the command
    cmd := exec.CommandContext(ctx, magickCmd, args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return agent.NewToolResult("convert_image",
            fmt.Sprintf("Conversion failed: %v\nOutput: %s", err, string(output)),
            true), nil
    }

    return agent.NewToolResult("convert_image",
        fmt.Sprintf("Successfully converted image from %s to %s",
            filepath.Ext(in.Src), filepath.Ext(in.Dst)),
        false), nil
}