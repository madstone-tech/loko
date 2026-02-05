package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

// TestServerInitialization tests creating a new MCP server.
func TestServerInitialization(t *testing.T) {
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer([]byte{})

	server := NewServer("test", input, output)
	if server == nil {
		t.Fatal("server should not be nil")
	}

	if server.ProjectRoot != "test" {
		t.Errorf("expected ProjectRoot='test', got '%s'", server.ProjectRoot)
	}
}

// TestToolRegistration tests registering tools with the server.
func TestToolRegistration(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	tool := &MockTool{
		NameValue:        "test_tool",
		DescriptionValue: "A test tool",
		InputSchemaValue: map[string]any{
			"type": "object",
		},
	}

	err := server.RegisterTool(tool)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(server.tools) != 1 {
		t.Errorf("expected 1 tool registered, got %d", len(server.tools))
	}

	if server.tools["test_tool"] != tool {
		t.Error("tool not found in registry")
	}
}

// TestDuplicateToolRegistration tests that duplicate tool names are rejected.
func TestDuplicateToolRegistration(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	tool1 := &MockTool{NameValue: "duplicate"}
	tool2 := &MockTool{NameValue: "duplicate"}

	server.RegisterTool(tool1)
	err := server.RegisterTool(tool2)

	if err == nil {
		t.Fatal("expected error for duplicate tool name")
	}
}

// TestToolsListRequest tests the tools/list request (MCP protocol).
func TestToolsListRequest(t *testing.T) {
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer([]byte{})

	server := NewServer("test", input, output)
	server.RegisterTool(&MockTool{
		NameValue:        "tool1",
		DescriptionValue: "First tool",
	})
	server.RegisterTool(&MockTool{
		NameValue:        "tool2",
		DescriptionValue: "Second tool",
	})

	// Build tools/list request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}

	response := server.handleRequest(request)

	// Verify response structure
	if response["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %v", response["jsonrpc"])
	}

	if response["id"] != 1 {
		t.Errorf("expected id=1, got %v", response["id"])
	}

	// Check result
	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("result should be a map")
	}

	toolsVal := result["tools"]
	if toolsVal == nil {
		t.Fatal("tools should be present in result")
	}

	tools, ok := toolsVal.([]map[string]any)
	if !ok {
		// Try converting from []any
		toolsIface, ok := toolsVal.([]any)
		if !ok {
			t.Fatalf("tools should be an array, got %T", toolsVal)
		}
		if len(toolsIface) != 2 {
			t.Errorf("expected 2 tools, got %d", len(toolsIface))
		}
		return
	}

	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

// TestCallToolRequest tests calling a specific tool.
func TestCallToolRequest(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	mockTool := &MockTool{
		NameValue:        "echo",
		DescriptionValue: "Echo tool",
		CallFunc: func(ctx context.Context, args map[string]any) (any, error) {
			return map[string]any{
				"echo": args["message"],
			}, nil
		},
	}

	server.RegisterTool(mockTool)

	// Build call request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "echo",
			"arguments": map[string]any{"message": "hello"},
		},
	}

	response := server.handleRequest(request)

	// Verify success response
	if response["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %v", response["jsonrpc"])
	}

	if response["id"] != 2 {
		t.Errorf("expected id=2, got %v", response["id"])
	}

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("result should be a map")
	}

	if result["echo"] != "hello" {
		t.Errorf("expected echo='hello', got %v", result["echo"])
	}
}

// TestCallNonexistentTool tests calling a tool that doesn't exist.
func TestCallNonexistentTool(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "nonexistent",
		},
	}

	response := server.handleRequest(request)

	// Should return error
	if _, hasError := response["error"]; !hasError {
		t.Error("expected error response for nonexistent tool")
	}
}

// TestInitializeRequest tests the initialize request (MCP protocol).
func TestInitializeRequest(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      0,
		"method":  "initialize",
		"params": map[string]any{
			"protocol_version": "2024-11-05",
			"capabilities":     map[string]any{},
			"client_info": map[string]any{
				"name":    "Claude",
				"version": "3.5",
			},
		},
	}

	response := server.handleRequest(request)

	// Verify response
	if response["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %v", response["jsonrpc"])
	}

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("result should be a map")
	}

	// Should have server_info
	serverInfo, ok := result["server_info"].(map[string]any)
	if !ok {
		t.Fatal("server_info should be a map")
	}

	if serverInfo["name"] != "loko" {
		t.Errorf("expected name='loko', got %v", serverInfo["name"])
	}
}

// TestInvalidRequest tests handling invalid requests.
func TestInvalidRequest(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	// Request missing required fields
	request := map[string]any{
		"jsonrpc": "2.0",
		// missing id
		"method": "some_method",
	}

	response := server.handleRequest(request)

	// Should return error
	if _, hasError := response["error"]; !hasError {
		t.Error("expected error response for invalid request")
	}
}

// TestJSONParsing tests parsing JSON-RPC requests from stdin.
func TestJSONParsing(t *testing.T) {
	jsonData := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	input := bytes.NewBufferString(jsonData + "\n")

	// Parse request
	req := make(map[string]any)
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&req)

	if err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if req["method"] != "tools/list" {
		t.Errorf("expected method='tools/list', got %v", req["method"])
	}
}

// TestResponseWriting tests writing JSON-RPC responses to stdout.
func TestResponseWriting(t *testing.T) {
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer([]byte{})

	server := NewServer("test", input, output)

	response := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]any{
			"key": "value",
		},
	}

	err := server.writeResponse(response)
	if err != nil {
		t.Fatalf("failed to write response: %v", err)
	}

	// Verify output contains JSON
	outputStr := output.String()
	if outputStr == "" {
		t.Fatal("output should not be empty")
	}

	// Parse response to verify it's valid JSON
	var parsedResp map[string]any
	decoder := json.NewDecoder(bytes.NewBufferString(outputStr))
	err = decoder.Decode(&parsedResp)
	if err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsedResp["id"] != float64(1) { // JSON decodes numbers as float64
		t.Errorf("expected id=1, got %v", parsedResp["id"])
	}
}

// TestErrorResponse tests error response format (MCP spec).
func TestErrorResponse(t *testing.T) {
	server := NewServer("test", bytes.NewBufferString(""), bytes.NewBuffer([]byte{}))

	response := server.errorResponse(42, -32600, "Invalid Request", nil)

	if response["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc=2.0")
	}

	if response["id"] != 42 {
		t.Errorf("expected id=42, got %v", response["id"])
	}

	errorObj, ok := response["error"].(map[string]any)
	if !ok {
		t.Fatal("error should be a map")
	}

	if errorObj["code"] != -32600 {
		t.Errorf("expected code=-32600, got %v", errorObj["code"])
	}

	if errorObj["message"] != "Invalid Request" {
		t.Errorf("expected message='Invalid Request', got %v", errorObj["message"])
	}
}

// Mock implementations for testing

type MockTool struct {
	NameValue        string
	DescriptionValue string
	InputSchemaValue map[string]any
	CallFunc         func(ctx context.Context, args map[string]any) (any, error)
}

func (m *MockTool) Name() string {
	return m.NameValue
}

func (m *MockTool) Description() string {
	return m.DescriptionValue
}

func (m *MockTool) InputSchema() map[string]any {
	if m.InputSchemaValue != nil {
		return m.InputSchemaValue
	}
	return map[string]any{"type": "object"}
}

func (m *MockTool) Call(ctx context.Context, args map[string]any) (any, error) {
	if m.CallFunc != nil {
		return m.CallFunc(ctx, args)
	}
	return nil, nil
}
