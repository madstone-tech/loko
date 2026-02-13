// Package mcp provides Model Context Protocol (MCP) server implementation.
// MCP allows LLMs to interact with loko via a standardized protocol.
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// Tool represents an MCP tool that can be called by the client.
type Tool interface {
	// Name returns the tool's name (unique identifier)
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// InputSchema returns JSON Schema for the tool's input parameters
	InputSchema() map[string]any

	// Call executes the tool with the given arguments
	Call(ctx context.Context, args map[string]any) (any, error)
}

// Server implements the MCP server for loko.
// It communicates with clients via JSON-RPC 2.0 over stdio.
type Server struct {
	ProjectRoot string
	input       io.Reader
	output      io.Writer
	tools       map[string]Tool
	toolsMutex  sync.RWMutex
	graphCache  *GraphCache // Cache for architecture graphs
}

// NewServer creates a new MCP server.
func NewServer(projectRoot string, input io.Reader, output io.Writer) *Server {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}

	return &Server{
		ProjectRoot: projectRoot,
		input:       input,
		output:      output,
		tools:       make(map[string]Tool),
		graphCache:  NewGraphCache(),
	}
}

// GetGraphCache returns the server's graph cache for tool access.
func (s *Server) GetGraphCache() *GraphCache {
	return s.graphCache
}

// RegisterTool adds a tool to the server's registry.
// Returns error if a tool with the same name is already registered.
func (s *Server) RegisterTool(tool Tool) error {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	if _, exists := s.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name())
	}

	s.tools[tool.Name()] = tool
	return nil
}

// Run starts the MCP server, reading JSON-RPC requests from input
// and writing responses to output.
func (s *Server) Run(ctx context.Context) error {
	decoder := json.NewDecoder(s.input)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var request map[string]any
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				return nil
			}
			// Send error response for malformed JSON
			_ = s.writeResponse(s.errorResponse(nil, -32700, "Parse error", nil))
			continue
		}

		response := s.handleRequest(request)
		if err := s.writeResponse(response); err != nil {
			fmt.Fprintf(os.Stderr, "error writing response: %v\n", err)
		}
	}
}

// handleRequest processes a single JSON-RPC request and returns the response.
func (s *Server) handleRequest(request map[string]any) map[string]any {
	// Validate request structure
	id, ok := request["id"]
	if !ok {
		return s.errorResponse(nil, -32600, "Invalid Request: missing id", nil)
	}

	method, ok := request["method"].(string)
	if !ok {
		return s.errorResponse(id, -32600, "Invalid Request: missing method", nil)
	}

	// Handle MCP protocol methods
	switch method {
	case "initialize":
		return s.handleInitialize(id, request)
	case "tools/list":
		return s.handleToolsList(id)
	case "tools/call":
		return s.handleToolCall(id, request)
	default:
		return s.errorResponse(id, -32601, fmt.Sprintf("Method not found: %s", method), nil)
	}
}

// handleInitialize handles the initialize request (MCP protocol handshake).
func (s *Server) handleInitialize(id any, request map[string]any) map[string]any {
	// Log client info if available
	if params, ok := request["params"].(map[string]any); ok {
		if clientInfo, ok := params["clientInfo"].(map[string]any); ok {
			fmt.Fprintf(os.Stderr, "MCP client connected: %v\n", clientInfo)
		}
	}

	result := map[string]any{
		"protocolVersion": "2025-06-18",
		"serverInfo": map[string]any{
			"name":    "loko",
			"version": "0.1.0",
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
}

// handleToolsList handles the tools/list request.
func (s *Server) handleToolsList(id any) map[string]any {
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	tools := make([]map[string]any, 0, len(s.tools))
	for _, tool := range s.tools {
		toolDesc := map[string]any{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		}
		tools = append(tools, toolDesc)
	}

	result := map[string]any{
		"tools": tools,
	}

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
}

// handleToolCall handles the tools/call request.
func (s *Server) handleToolCall(id any, request map[string]any) map[string]any {
	params, ok := request["params"].(map[string]any)
	if !ok {
		return s.errorResponse(id, -32602, "Invalid params", nil)
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return s.errorResponse(id, -32602, "Missing tool name", nil)
	}

	arguments, ok := params["arguments"].(map[string]any)
	if !ok {
		arguments = make(map[string]any)
	}

	// Look up and call the tool
	s.toolsMutex.RLock()
	tool, exists := s.tools[toolName]
	s.toolsMutex.RUnlock()

	if !exists {
		return s.errorResponse(id, -32601, fmt.Sprintf("Tool not found: %s", toolName), nil)
	}

	// Call the tool
	result, err := tool.Call(context.Background(), arguments)
	if err != nil {
		return s.errorResponse(id, -32000, fmt.Sprintf("Tool error: %v", err), nil)
	}

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  wrapToolResult(result),
	}
}

// wrapToolResult wraps a tool result in the MCP content array format.
// MCP protocol requires tool results as {"content": [{"type": "text", "text": "..."}]}.
func wrapToolResult(result any) map[string]any {
	textContent, err := json.Marshal(result)
	if err != nil {
		textContent = []byte(fmt.Sprintf("%v", result))
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": string(textContent),
			},
		},
	}
}

// errorResponse creates a JSON-RPC error response.
func (s *Server) errorResponse(id any, code int, message string, data any) map[string]any {
	errorObj := map[string]any{
		"code":    code,
		"message": message,
	}

	if data != nil {
		errorObj["data"] = data
	}

	response := map[string]any{
		"jsonrpc": "2.0",
		"error":   errorObj,
	}

	if id != nil {
		response["id"] = id
	}

	return response
}

// writeResponse writes a JSON-RPC response to output.
func (s *Server) writeResponse(response map[string]any) error {
	// Create a buffer to ensure the entire response is written at once
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	// Write to output
	if _, err := io.Copy(s.output, &buf); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	// Flush if the output supports it (important for stdio)
	if flusher, ok := s.output.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
	}

	return nil
}
