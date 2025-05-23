package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ToolDefinition represents a tool that can be used by the agent
type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

// ReadFile tool
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", fmt.Errorf("invalid input format for read_file: %w", err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", readFileInput.Path, err)
	}
	return string(content), nil
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

// ListFiles tool
type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

func ListFiles(input json.RawMessage) (string, error) {
	listFilesInput := ListFilesInput{}
	err := json.Unmarshal(input, &listFilesInput)
	if err != nil {
		return "", fmt.Errorf("invalid input format for list_files: %w", err)
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to list files in '%s': %w", dir, err)
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", fmt.Errorf("failed to marshal file list: %w", err)
	}

	return string(result), nil
}

var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

// EditFile tool
type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

var EditFileInputSchema = GenerateSchema[EditFileInput]()

func EditFile(input json.RawMessage) (string, error) {
	editFileInput := EditFileInput{}
	err := json.Unmarshal(input, &editFileInput)
	if err != nil {
		return "", fmt.Errorf("invalid input format for edit_file: %w", err)
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s' for editing: %w", editFileInput.Path, err)
	}

	contentStr := string(content)
	newContentStr := strings.Replace(contentStr, editFileInput.OldStr, editFileInput.NewStr, 1)
	if newContentStr == contentStr {
		return "", fmt.Errorf("string '%s' not found in file '%s'", editFileInput.OldStr, editFileInput.Path)
	}

	err = os.WriteFile(editFileInput.Path, []byte(newContentStr), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write changes to file '%s': %w", editFileInput.Path, err)
	}

	return "File edited successfully", nil
}

var EditFileDefinition = ToolDefinition{
	Name:        "edit_file",
	Description: "Edit a file by replacing a specific string with another string. The old string must match exactly and must only have one match in the file.",
	InputSchema: EditFileInputSchema,
	Function:    EditFile,
}

// RipGrepSearch tool
type RipGrepInput struct {
	Query   string `json:"query" jsonschema_description:"The ripgrep compatible regex pattern to search for."`
	Path    string `json:"path,omitempty" jsonschema_description:"Optional file or directory path to search within. Defaults to current directory if empty."`
	IgnoreCase bool `json:"ignore_case,omitempty" jsonschema_description:"Perform case-insensitive search."`
	MaxCount   int    `json:"max_count,omitempty" jsonschema_description:"Limit the number of matches per file."`
}

var RipGrepInputSchema = GenerateSchema[RipGrepInput]()

func RipGrepSearch(input json.RawMessage) (string, error) {
	rgInput := RipGrepInput{}
	err := json.Unmarshal(input, &rgInput)
	if err != nil {
		return "", fmt.Errorf("invalid input format for ripgrep_search: %w", err)
	}

	args := []string{"--no-heading", "--with-filename", "--line-number"}
	if rgInput.IgnoreCase {
		args = append(args, "--ignore-case")
	}
	if rgInput.MaxCount > 0 {
		args = append(args, fmt.Sprintf("--max-count=%d", rgInput.MaxCount))
	}
	args = append(args, "--", rgInput.Query)
	if rgInput.Path != "" {
		args = append(args, rgInput.Path)
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command("rg", args...)
	out, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return "No matches found.", nil
			} else {
				stderr := string(exitErr.Stderr)
				if stderr != "" {
					return "", fmt.Errorf("ripgrep failed with exit code %d: %s", exitErr.ExitCode(), stderr)
				} else {
					return "", fmt.Errorf("ripgrep failed with exit code %d", exitErr.ExitCode())
				}
			}
		} else {
			return "", fmt.Errorf("failed to execute ripgrep: %w", err)
		}
	}

	if len(out) == 0 {
		return "No matches found.", nil
	}

	return string(out), nil
}

var RipGrepToolDefinition = ToolDefinition{
	Name:        "ripgrep_search",
	Description: "Search for a regex pattern in files using ripgrep. Provides filename and line number for matches.",
	InputSchema: RipGrepInputSchema,
	Function:    RipGrepSearch,
}

// GetTools returns all available tools
func GetTools() []ToolDefinition {
	return []ToolDefinition{
		ReadFileDefinition,
		ListFilesDefinition,
		EditFileDefinition,
		RipGrepToolDefinition,
	}
}

type ToolError struct {
	ToolName string
	Err      error
}

func (e *ToolError) Error() string {
	return fmt.Sprintf("tool %s: %v", e.ToolName, e.Err)
}

type MessageHandler func() (string, bool)