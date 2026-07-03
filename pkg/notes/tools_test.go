package notes

import (
	"encoding/json"
	"testing"

	"github.com/github/github-mcp-server/internal/toolsnaps"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCallRequest builds a mcp.CallToolRequest with the given JSON arguments.
func newCallRequest(t *testing.T, args any) *mcp.CallToolRequest {
	t.Helper()
	raw, err := json.Marshal(args)
	require.NoError(t, err)
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: json.RawMessage(raw),
		},
	}
}

// resultText extracts the first text content from a CallToolResult.
func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	return text.Text
}

func Test_ListNotes(t *testing.T) {
	store := NewMemoryStore()
	serverTool := ListNotes(store)
	tool := serverTool.Tool

	require.NoError(t, toolsnaps.Test(tool.Name, tool))
	assert.Equal(t, "list_notes", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.True(t, tool.Annotations.ReadOnlyHint, "list_notes should be read-only")

	t.Run("empty store", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{}))
		require.NoError(t, err)
		assert.False(t, result.IsError)
		assert.Equal(t, "[]", resultText(t, result))
	})

	t.Run("with notes", func(t *testing.T) {
		store := NewMemoryStore()
		store.Create("Hello", "World")
		store.Create("Foo", "Bar")
		serverTool := ListNotes(store)

		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{}))
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var got []*Note
		require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got))
		assert.Len(t, got, 2)
	})
}

func Test_GetNote(t *testing.T) {
	store := NewMemoryStore()
	serverTool := GetNote(store)
	tool := serverTool.Tool

	require.NoError(t, toolsnaps.Test(tool.Name, tool))
	assert.Equal(t, "get_note", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.True(t, tool.Annotations.ReadOnlyHint, "get_note should be read-only")

	t.Run("missing id", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{}))
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("note not found", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{"id": "999"}))
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "not found")
	})

	t.Run("note found", func(t *testing.T) {
		store := NewMemoryStore()
		created := store.Create("My Note", "Some content")
		serverTool := GetNote(store)

		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{"id": created.ID}))
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var got Note
		require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got))
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "My Note", got.Title)
		assert.Equal(t, "Some content", got.Content)
	})
}

func Test_CreateNote(t *testing.T) {
	store := NewMemoryStore()
	serverTool := CreateNote(store)
	tool := serverTool.Tool

	require.NoError(t, toolsnaps.Test(tool.Name, tool))
	assert.Equal(t, "create_note", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.False(t, tool.Annotations.ReadOnlyHint, "create_note should not be read-only")

	t.Run("missing title", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{"content": "hello"}))
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "title is required")
	})

	t.Run("missing content", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{"title": "hi"}))
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "content is required")
	})

	t.Run("successful creation", func(t *testing.T) {
		store := NewMemoryStore()
		serverTool := CreateNote(store)

		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{
			"title":   "Test Note",
			"content": "Test content",
		}))
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var got Note
		require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got))
		assert.Equal(t, "Test Note", got.Title)
		assert.Equal(t, "Test content", got.Content)
		assert.NotEmpty(t, got.ID)
	})
}

func Test_DeleteNote(t *testing.T) {
	store := NewMemoryStore()
	serverTool := DeleteNote(store)
	tool := serverTool.Tool

	require.NoError(t, toolsnaps.Test(tool.Name, tool))
	assert.Equal(t, "delete_note", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.False(t, tool.Annotations.ReadOnlyHint, "delete_note should not be read-only")

	t.Run("missing id", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{}))
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "id is required")
	})

	t.Run("note not found", func(t *testing.T) {
		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{"id": "999"}))
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "not found")
	})

	t.Run("successful deletion", func(t *testing.T) {
		store := NewMemoryStore()
		created := store.Create("To Delete", "content")
		serverTool := DeleteNote(store)

		result, err := serverTool.Handler(nil)(t.Context(), newCallRequest(t, map[string]any{"id": created.ID}))
		require.NoError(t, err)
		assert.False(t, result.IsError)
		assert.Contains(t, resultText(t, result), "deleted")

		// Verify the note is gone
		_, err = store.Get(created.ID)
		assert.Error(t, err)
	})
}
