package notes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/inventory"
	"github.com/github/github-mcp-server/pkg/utils"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolsetMetadataNotes is the toolset metadata for the notes toolset.
var ToolsetMetadataNotes = inventory.ToolsetMetadata{
	ID:          "notes",
	Description: "Tools for creating and managing in-memory notes",
	Default:     true,
	Icon:        "note",
}

// ListNotes returns a tool that lists all notes in the store.
func ListNotes(store Store) inventory.ServerTool {
	return inventory.NewServerTool(
		mcp.Tool{
			Name:        "list_notes",
			Description: "List all notes",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Notes",
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
			},
		},
		ToolsetMetadataNotes,
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			notes := store.List()
			data, err := json.Marshal(notes)
			if err != nil {
				return utils.NewToolResultErrorFromErr("failed to marshal notes", err), nil
			}
			return utils.NewToolResultText(string(data)), nil
		},
	)
}

// GetNote returns a tool that retrieves a single note by ID.
func GetNote(store Store) inventory.ServerTool {
	return inventory.NewServerTool(
		mcp.Tool{
			Name:        "get_note",
			Description: "Get a note by ID",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Get Note",
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"id": {
						Type:        "string",
						Description: "The ID of the note to retrieve",
					},
				},
				Required: []string{"id"},
			},
		},
		ToolsetMetadataNotes,
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return utils.NewToolResultError(fmt.Sprintf("invalid arguments: %s", err)), nil
			}
			if args.ID == "" {
				return utils.NewToolResultError("id is required"), nil
			}

			note, err := store.Get(args.ID)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			data, err := json.Marshal(note)
			if err != nil {
				return utils.NewToolResultErrorFromErr("failed to marshal note", err), nil
			}
			return utils.NewToolResultText(string(data)), nil
		},
	)
}

// CreateNote returns a tool that creates a new note.
func CreateNote(store Store) inventory.ServerTool {
	return inventory.NewServerTool(
		mcp.Tool{
			Name:        "create_note",
			Description: "Create a new note with a title and content",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Create Note",
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"title": {
						Type:        "string",
						Description: "The title of the note",
					},
					"content": {
						Type:        "string",
						Description: "The body content of the note",
					},
				},
				Required: []string{"title", "content"},
			},
		},
		ToolsetMetadataNotes,
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return utils.NewToolResultError(fmt.Sprintf("invalid arguments: %s", err)), nil
			}
			if args.Title == "" {
				return utils.NewToolResultError("title is required"), nil
			}
			if args.Content == "" {
				return utils.NewToolResultError("content is required"), nil
			}

			note := store.Create(args.Title, args.Content)

			data, err := json.Marshal(note)
			if err != nil {
				return utils.NewToolResultErrorFromErr("failed to marshal note", err), nil
			}
			return utils.NewToolResultText(string(data)), nil
		},
	)
}

// DeleteNote returns a tool that deletes a note by ID.
func DeleteNote(store Store) inventory.ServerTool {
	return inventory.NewServerTool(
		mcp.Tool{
			Name:        "delete_note",
			Description: "Delete a note by ID",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Delete Note",
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"id": {
						Type:        "string",
						Description: "The ID of the note to delete",
					},
				},
				Required: []string{"id"},
			},
		},
		ToolsetMetadataNotes,
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return utils.NewToolResultError(fmt.Sprintf("invalid arguments: %s", err)), nil
			}
			if args.ID == "" {
				return utils.NewToolResultError("id is required"), nil
			}

			if err := store.Delete(args.ID); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			return utils.NewToolResultText(fmt.Sprintf("note %q deleted", args.ID)), nil
		},
	)
}

// AllTools returns all notes tools wired to the given store.
func AllTools(store Store) []inventory.ServerTool {
	return []inventory.ServerTool{
		ListNotes(store),
		GetNote(store),
		CreateNote(store),
		DeleteNote(store),
	}
}
