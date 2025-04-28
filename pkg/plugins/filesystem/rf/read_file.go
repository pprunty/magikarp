package read_file

import (
    _ "embed"
    "context"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
    "unicode/utf8"

    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var wrapper []byte

// FileCache is a simple in-memory cache for file chunks
var FileCache = make(map[string][]byte)

/* ------------------------------------------------------------------ */

// FileID identifies a specific file in the system
type FileID struct {
    Path       string    `json:"path"`
    ContentHash string    `json:"content_hash"`
    ModTime    time.Time `json:"mod_time"`
}

// input for read_file represents the parameters for the read file tool
type input struct {
    Path           string `json:"path"`
    PreviewMode    bool   `json:"preview_mode,omitempty"`    // Return preview + metadata instead of full content
    PreviewLines   int    `json:"preview_lines,omitempty"`   // Number of lines to include in preview
    PreviewBytes   int    `json:"preview_bytes,omitempty"`   // Max bytes to include in preview
    DetectEncoding bool   `json:"detect_encoding,omitempty"` // Handle non-UTF8 files
}

// input for read_file_chunk represents the parameters for reading a specific chunk
type chunkInput struct {
    FileID     string `json:"file_id"`       // ID of previously previewed file
    StartLine  int    `json:"start_line"`    // First line to read (0-indexed)
    EndLine    int    `json:"end_line"`      // Last line to read (exclusive)
    StartByte  int    `json:"start_byte"`    // Alternative: start at byte offset
    ByteLength int    `json:"byte_length"`   // Alternative: read specified bytes
}

// PreviewResponse contains the preview and metadata for a file
type PreviewResponse struct {
    FileID        string `json:"file_id"`           // Stable ID for future chunk requests
    Preview       string `json:"preview"`           // Content preview (first few lines/bytes)
    ByteSize      int64  `json:"byte_size"`         // Total size of the file in bytes
    Lines         int    `json:"lines"`             // Total line count
    HasMore       bool   `json:"has_more"`          // Whether there's more content beyond preview
    MIME          string `json:"mime_type"`         // Best-guess MIME type
    ModTime       string `json:"mod_time"`          // Last modification timestamp
    IsBinary      bool   `json:"is_binary"`         // Whether file appears to be binary
    Language      string `json:"language,omitempty"` // Programming language detection (optional)
    ContainsHTML  bool   `json:"contains_html,omitempty"` // Whether file contains HTML markup
}

func Definition() agent.ToolDefinition {
    var w map[string]any
    if err := json.Unmarshal(wrapper, &w); err != nil {
        fmt.Printf("Error unmarshaling read_file schema: %v\n", err)
    }

    schema := w["input_schema"].(map[string]any)

    params := map[string]agent.Parameter{}
    if props, ok := schema["properties"].(map[string]any); ok {
        for name, raw := range props {
            m := raw.(map[string]any)
            params[name] = agent.Parameter{
                Type:        fmt.Sprint(m["type"]),
                Description: fmt.Sprint(m["description"]),
                Required:    contains(schema["required"], name),
            }
        }
    }

    return agent.ToolDefinition{
        Name:        "read_file",
        Description: w["description"].(string),
        InputSchema: schema,
        Parameters:  params,
        Function:    run,
    }
}

/* ------------------------------------------------------------------ */

func run(ctx context.Context, inMap map[string]any) (*agent.ToolResult, error) {
    // Parse input parameters
    var in input
    inputBytes, err := json.Marshal(inMap)
    if err != nil {
        return agent.NewToolResult("read_file", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("read_file", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
    }

    // Set default preview sizes if not specified
    if in.PreviewLines <= 0 {
        in.PreviewLines = 10 // Default to 10 lines in preview
    }
    if in.PreviewBytes <= 0 {
        in.PreviewBytes = 1000 // Default to 1000 bytes in preview
    }
    if in.PreviewBytes > 4000 {
        in.PreviewBytes = 4000 // Cap preview at 4000 bytes (~1000 tokens)
    }

    // Validate path
    if in.Path == "" {
        return agent.NewToolResult("read_file", "Path parameter is required", true), nil
    }

    if !filepath.IsLocal(in.Path) {
        return agent.NewToolResult("read_file", "Path must be local for security reasons", true), nil
    }

    // Clean path and handle case-insensitive search if needed
    path := filepath.Clean(in.Path)

    if _, err := os.Stat(path); os.IsNotExist(err) {
        // Try case-insensitive match
        dir := filepath.Dir(path)
        base := filepath.Base(path)

        // Make sure the directory exists
        entries, err := os.ReadDir(dir)
        if err != nil {
            return agent.NewToolResult("read_file",
                fmt.Sprintf("Directory not found or not accessible: %s", dir), true), nil
        }

        // Look for case-insensitive match
        found := false
        for _, e := range entries {
            if strings.EqualFold(e.Name(), base) {
                path = filepath.Join(dir, e.Name())
                found = true
                break
            }
        }

        if !found {
            return agent.NewToolResult("read_file",
                fmt.Sprintf("File not found: %s (no case-insensitive match found)", in.Path), true), nil
        }
    }

    // Check if path exists and is a file
    fileInfo, err := os.Stat(path)
    if err != nil {
        return agent.NewToolResult("read_file", fmt.Sprintf("Error accessing file: %v", err), true), nil
    }

    if fileInfo.IsDir() {
        return agent.NewToolResult("read_file", "Path points to a directory, not a file", true), nil
    }

    // Always use preview mode for large files to prevent context bloat
    if !in.PreviewMode && fileInfo.Size() > 100000 { // 100KB
        in.PreviewMode = true
    }

    // Read file and generate file ID for cache
    file, err := os.Open(path)
    if err != nil {
        return agent.NewToolResult("read_file", fmt.Sprintf("Error opening file: %v", err), true), nil
    }
    defer file.Close()

    // Calculate file hash for ID
    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return agent.NewToolResult("read_file", fmt.Sprintf("Error calculating file hash: %v", err), true), nil
    }
    contentHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

    // Create stable file ID
    fileID := fmt.Sprintf("file-%s", contentHash[:12])

    // Reset file position
    file.Seek(0, 0)

    // If preview mode, create a preview response
    if in.PreviewMode {
        // Read preview bytes
        preview := make([]byte, in.PreviewBytes)
        n, _ := file.Read(preview)
        preview = preview[:n]

        // Count total lines
        file.Seek(0, 0)
        lineCount := 0
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            lineCount++
        }

        // Check if preview has whole file
        hasMore := int64(n) < fileInfo.Size()

        // Ensure preview is valid UTF-8
        previewStr := string(preview)
        if !utf8.ValidString(previewStr) && !in.DetectEncoding {
            previewStr = "[Binary data - set detect_encoding=true to force display]"
        }

        // Trim preview to requested number of lines
        lines := strings.Split(previewStr, "\n")
        if len(lines) > in.PreviewLines {
            lines = lines[:in.PreviewLines]
            previewStr = strings.Join(lines, "\n")
            hasMore = true
        }

        // Guess MIME type
        mimeType := guessMIMEType(path, preview)

        // Detect if file contains HTML
        containsHTML := strings.Contains(previewStr, "<html") ||
                        strings.Contains(previewStr, "<body") ||
                        strings.Contains(previewStr, "<div")

        // Create preview response
        resp := PreviewResponse{
            FileID:       fileID,
            Preview:      previewStr,
            ByteSize:     fileInfo.Size(),
            Lines:        lineCount,
            HasMore:      hasMore,
            MIME:         mimeType,
            ModTime:      fileInfo.ModTime().Format(time.RFC3339),
            IsBinary:     !utf8.ValidString(previewStr) && in.DetectEncoding,
            ContainsHTML: containsHTML,
            Language:     detectLanguage(path, previewStr),
        }

        // Cache the file path and content hash for chunk requests
        metadataKey := fmt.Sprintf("metadata:%s", fileID)
        metadata := FileID{
            Path:        path,
            ContentHash: contentHash,
            ModTime:     fileInfo.ModTime(),
        }
        metadataBytes, _ := json.Marshal(metadata)
        FileCache[metadataKey] = metadataBytes

        // Return preview response
        respJSON, _ := json.MarshalIndent(resp, "", "  ")
        return agent.NewToolResult("read_file", string(respJSON), false), nil
    }

    // For small files, just return the content directly
    content, err := io.ReadAll(file)
    if err != nil {
        return agent.NewToolResult("read_file", fmt.Sprintf("Error reading file: %v", err), true), nil
    }

    // Check if content is valid UTF-8
    if !utf8.Valid(content) && !in.DetectEncoding {
        return agent.NewToolResult("read_file",
            "File contains invalid UTF-8 sequences. Set detect_encoding=true to attempt conversion.", true), nil
    }

    // Cache the file for later chunk requests
    FileCache[fileID] = content

    // Return the full content
    return agent.NewToolResult("read_file", string(content), false), nil
}

// ChunkDefinition provides the tool definition for read_file_chunk
func ChunkDefinition() agent.ToolDefinition {
    return agent.ToolDefinition{
        Name:        "read_file_chunk",
        Description: "Reads a specific chunk from a previously previewed file, either by line range or byte range.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "file_id": map[string]any{
                    "type": "string",
                    "description": "The file_id from a previous read_file preview response",
                },
                "start_line": map[string]any{
                    "type": "integer",
                    "description": "First line to read (0-indexed)",
                },
                "end_line": map[string]any{
                    "type": "integer",
                    "description": "Last line to read (exclusive)",
                },
                "start_byte": map[string]any{
                    "type": "integer",
                    "description": "Alternative: start at byte offset",
                },
                "byte_length": map[string]any{
                    "type": "integer",
                    "description": "Alternative: read specified bytes",
                },
            },
            "required": []string{"file_id"},
            "additionalProperties": false,
        },
        Function: runChunk,
    }
}

// runChunk handles requests for specific chunks of a previously accessed file
func runChunk(ctx context.Context, inMap map[string]any) (*agent.ToolResult, error) {
    // Parse input parameters
    var in chunkInput
    inputBytes, err := json.Marshal(inMap)
    if err != nil {
        return agent.NewToolResult("read_file_chunk", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("read_file_chunk", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
    }

    // Validate file_id
    if in.FileID == "" {
        return agent.NewToolResult("read_file_chunk", "file_id is required", true), nil
    }

    // Get file metadata from cache
    metadataKey := fmt.Sprintf("metadata:%s", in.FileID)
    metadataBytes, ok := FileCache[metadataKey]
    if !ok {
        return agent.NewToolResult("read_file_chunk", "Invalid or expired file_id", true), nil
    }

    var metadata FileID
    if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
        return agent.NewToolResult("read_file_chunk", "Error parsing file metadata", true), nil
    }

    // Read file content (from cache if available)
    var content []byte
    if cachedContent, ok := FileCache[in.FileID]; ok {
        content = cachedContent
    } else {
        // Verify file still exists and is unchanged
        fileInfo, err := os.Stat(metadata.Path)
        if err != nil {
            return agent.NewToolResult("read_file_chunk", "File no longer accessible", true), nil
        }

        if !fileInfo.ModTime().Equal(metadata.ModTime) {
            return agent.NewToolResult("read_file_chunk", "File has been modified since preview", true), nil
        }

        // Read file
        content, err = os.ReadFile(metadata.Path)
        if err != nil {
            return agent.NewToolResult("read_file_chunk", fmt.Sprintf("Error reading file: %v", err), true), nil
        }

        // Cache for future requests
        FileCache[in.FileID] = content
    }

    // Extract the requested chunk
    var chunk []byte

    if in.StartLine >= 0 && in.EndLine > in.StartLine {
        // Extract by line range
        lines := strings.Split(string(content), "\n")

        if in.StartLine >= len(lines) {
            return agent.NewToolResult("read_file_chunk", "Start line out of range", true), nil
        }

        endLine := in.EndLine
        if endLine > len(lines) {
            endLine = len(lines)
        }

        selectedLines := lines[in.StartLine:endLine]
        chunk = []byte(strings.Join(selectedLines, "\n"))
    } else if in.StartByte >= 0 && in.ByteLength > 0 {
        // Extract by byte range
        if in.StartByte >= len(content) {
            return agent.NewToolResult("read_file_chunk", "Start byte out of range", true), nil
        }

        endByte := in.StartByte + in.ByteLength
        if endByte > len(content) {
            endByte = len(content)
        }

        chunk = content[in.StartByte:endByte]
    } else {
        return agent.NewToolResult("read_file_chunk",
            "Either start_line/end_line or start_byte/byte_length must be provided", true), nil
    }

    // Return the chunk
    return agent.NewToolResult("read_file_chunk", string(chunk), false), nil
}

/* ------------------------------------------------------------------ */

// Helper functions

// guessMIMEType makes a simple guess at the MIME type based on file extension and content
func guessMIMEType(path string, preview []byte) string {
    ext := strings.ToLower(filepath.Ext(path))

    // Common text formats
    switch ext {
    case ".txt":
        return "text/plain"
    case ".html", ".htm":
        return "text/html"
    case ".md":
        return "text/markdown"
    case ".csv":
        return "text/csv"
    case ".json":
        return "application/json"
    case ".xml":
        return "application/xml"
    }

    // Programming languages
    switch ext {
    case ".py":
        return "text/x-python"
    case ".js":
        return "text/javascript"
    case ".go":
        return "text/x-go"
    case ".java":
        return "text/x-java"
    case ".c", ".cpp", ".h":
        return "text/x-c"
    }

    // Check if content is ASCII text
    if isProbablyText(preview) {
        return "text/plain"
    }

    return "application/octet-stream"
}

// detectLanguage makes a simple guess at the programming language
func detectLanguage(path string, preview string) string {
    ext := strings.ToLower(filepath.Ext(path))

    switch ext {
    case ".py":
        return "python"
    case ".js":
        return "javascript"
    case ".go":
        return "go"
    case ".java":
        return "java"
    case ".c":
        return "c"
    case ".cpp":
        return "c++"
    case ".rb":
        return "ruby"
    case ".php":
        return "php"
    case ".cs":
        return "csharp"
    case ".ts":
        return "typescript"
    case ".rs":
        return "rust"
    case ".sh":
        return "shell"
    }

    // Try to determine from content
    lowerPreview := strings.ToLower(preview)

    if strings.Contains(lowerPreview, "def ") && strings.Contains(lowerPreview, "import ") {
        return "python"
    }
    if strings.Contains(lowerPreview, "function ") && strings.Contains(lowerPreview, "const ") {
        return "javascript"
    }
    if strings.Contains(lowerPreview, "package ") && strings.Contains(lowerPreview, "func ") {
        return "go"
    }
    if strings.Contains(lowerPreview, "public class ") || strings.Contains(lowerPreview, "import java.") {
        return "java"
    }

    return ""
}

// isProbablyText returns true if the content is likely text rather than binary
func isProbablyText(content []byte) bool {
    // Simple heuristic: count the ratio of control characters and high-bit characters
    if len(content) == 0 {
        return true
    }

    controlCount := 0
    for _, b := range content {
        if (b < 32 && b != '\n' && b != '\r' && b != '\t') || b >= 127 {
            controlCount++
        }
    }

    ratio := float64(controlCount) / float64(len(content))
    return ratio < 0.1 // Arbitrary threshold
}

// bufio is necessary for counting lines
// Add a minimal implementation here to avoid adding dependency
type bufio struct{}

func NewScanner(r io.Reader) *Scanner {
    return &Scanner{r: r}
}

type Scanner struct {
    r   io.Reader
    buf []byte
    err error
}

func (s *Scanner) Scan() bool {
    if s.err != nil {
        return false
    }

    s.buf = make([]byte, 0, 4096)
    var b [1]byte
    for {
        n, err := s.r.Read(b[:])
        if n > 0 {
            s.buf = append(s.buf, b[0])
            if b[0] == '\n' {
                return true
            }
        }
        if err != nil {
            s.err = err
            if err == io.EOF {
                return len(s.buf) > 0
            }
            return false
        }
    }
}

func contains(raw any, key string) bool {
    if arr, ok := raw.([]any); ok {
        for _, v := range arr {
            if v == key {
                return true
            }
        }
    }
    return false
}