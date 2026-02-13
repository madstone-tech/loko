package tools

import (
	"context"
	"fmt"
	"testing"
)

// mockTool is a test implementation of the Tool interface.
type mockTool struct {
	name        string
	description string
	schema      map[string]any
}

func (t *mockTool) Name() string {
	return t.name
}

func (t *mockTool) Description() string {
	return t.description
}

func (t *mockTool) InputSchema() map[string]any {
	return t.schema
}

func (t *mockTool) Call(ctx context.Context, args map[string]any) (any, error) {
	return map[string]any{"result": "success"}, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	tool1 := &mockTool{name: "test_tool_1", description: "Test tool 1"}
	tool2 := &mockTool{name: "test_tool_2", description: "Test tool 2"}

	// Register tools
	if err := registry.Register(tool1); err != nil {
		t.Fatalf("Failed to register tool1: %v", err)
	}

	if err := registry.Register(tool2); err != nil {
		t.Fatalf("Failed to register tool2: %v", err)
	}

	// Try to register a duplicate tool
	duplicateTool := &mockTool{name: "test_tool_1", description: "Duplicate tool"}
	if err := registry.Register(duplicateTool); err == nil {
		t.Fatal("Expected error when registering duplicate tool, got nil")
	}

	// Retrieve tools
	retrievedTool1, ok := registry.Get("test_tool_1")
	if !ok {
		t.Fatal("Failed to get tool1")
	}
	if retrievedTool1.Name() != "test_tool_1" {
		t.Errorf("Expected tool name 'test_tool_1', got %q", retrievedTool1.Name())
	}

	retrievedTool2, ok := registry.Get("test_tool_2")
	if !ok {
		t.Fatal("Failed to get tool2")
	}
	if retrievedTool2.Name() != "test_tool_2" {
		t.Errorf("Expected tool name 'test_tool_2', got %q", retrievedTool2.Name())
	}

	// Try to get non-existent tool
	_, ok = registry.Get("non_existent_tool")
	if ok {
		t.Error("Expected false when getting non-existent tool, got true")
	}
}

func TestRegistry_ListAndNames(t *testing.T) {
	registry := NewRegistry()

	tool1 := &mockTool{name: "alpha", description: "Alpha tool"}
	tool2 := &mockTool{name: "beta", description: "Beta tool"}
	tool3 := &mockTool{name: "gamma", description: "Gamma tool"}

	// Register tools
	registry.Register(tool1)
	registry.Register(tool2)
	registry.Register(tool3)

	// Check list
	tools := registry.List()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}

	// Check names
	names := registry.Names()
	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	// Verify all names are present
	expectedNames := map[string]bool{
		"alpha": true,
		"beta":  true,
		"gamma": true,
	}

	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("Unexpected name in list: %s", name)
		}
		delete(expectedNames, name)
	}

	if len(expectedNames) != 0 {
		t.Errorf("Missing names: %v", expectedNames)
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	if count := registry.Count(); count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	registry.Register(&mockTool{name: "tool1"})
	if count := registry.Count(); count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	registry.Register(&mockTool{name: "tool2"})
	if count := registry.Count(); count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	done := make(chan bool)

	// Start multiple goroutines registering tools
	for i := 0; i < 10; i++ {
		go func(i int) {
			tool := &mockTool{
				name:        fmt.Sprintf("concurrent_tool_%d", i),
				description: fmt.Sprintf("Concurrent tool %d", i),
			}
			registry.Register(tool)
			done <- true
		}(i)
	}

	// Wait for all registrations
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all tools were registered
	if count := registry.Count(); count != 10 {
		t.Errorf("Expected count 10, got %d", count)
	}

	// Test concurrent reads
	readDone := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = registry.List()
			_ = registry.Names()
			_, _ = registry.Get("concurrent_tool_0")
			readDone <- true
		}()
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-readDone
	}
}
