package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestNewSearchElements tests creating a SearchElements use case.
func TestNewSearchElements(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	uc := NewSearchElements(mockRepo)

	if uc == nil {
		t.Error("NewSearchElements() returned nil")
	}

	if uc.repo != mockRepo {
		t.Error("NewSearchElements() did not set repo correctly")
	}

	if uc.buildGraph == nil {
		t.Error("NewSearchElements() did not initialize buildGraph")
	}
}

// TestSearchElementsExecute tests the Execute method of SearchElements.
func TestSearchElementsExecute(t *testing.T) {
	// Create test systems
	sys1, _ := entities.NewSystem("Payment Service")
	sys1.Description = "Handles payment processing"
	sys1.Tags = []string{"finance", "critical"}

	cont1, _ := entities.NewContainer("API Server")
	cont1.Description = "REST API endpoints"
	cont1.Technology = "Go + gRPC"
	cont1.Tags = []string{"api", "backend"}

	cont2, _ := entities.NewContainer("Database")
	cont2.Description = "PostgreSQL database"
	cont2.Technology = "PostgreSQL"
	cont2.Tags = []string{"database", "storage"}

	comp1, _ := entities.NewComponent("Auth Handler")
	comp1.Description = "Handles authentication"
	comp1.Technology = "Go"
	comp1.Tags = []string{"security", "core"}

	comp2, _ := entities.NewComponent("Payment Processor")
	comp2.Description = "Processes payments"
	comp2.Technology = "Go"
	comp2.Tags = []string{"finance", "core"}

	// Build hierarchy
	cont1.AddComponent(comp1)
	cont1.AddComponent(comp2)
	sys1.AddContainer(cont1)
	sys1.AddContainer(cont2)

	tests := []struct {
		name       string
		request    entities.SearchElementsRequest
		setupMocks func(*MockProjectRepository)
		wantErr    bool
		validate   func(t *testing.T, result *entities.SearchElementsResponse)
	}{
		{
			name: "search all elements",
			request: entities.SearchElementsRequest{
				ProjectRoot: "/test/project",
				Query:       "*",
				Limit:       10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.SearchElementsResponse) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				// Should find 4 elements: 1 system, 2 containers, 2 components
				if result.TotalMatched < 4 {
					t.Errorf("expected at least 4 total matched, got %d", result.TotalMatched)
				}
				if len(result.Results) < 4 {
					t.Errorf("expected at least 4 results, got %d", len(result.Results))
				}
				if result.Message == "" {
					t.Error("expected non-empty message")
				}
			},
		},
		{
			name: "search by name pattern",
			request: entities.SearchElementsRequest{
				ProjectRoot: "/test/project",
				Query:       "*",
				Limit:       10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.SearchElementsResponse) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				// Should find at least Payment Service system and Payment Processor component
				if result.TotalMatched < 2 {
					t.Errorf("expected at least 2 total matched, got %d", result.TotalMatched)
				}
			},
		},
		{
			name: "search by element type",
			request: entities.SearchElementsRequest{
				ProjectRoot: "/test/project",
				Query:       "*",
				Type:        "component",
				Limit:       10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.SearchElementsResponse) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				// Should find only components
				if result.TotalMatched != 2 {
					t.Errorf("expected 2 total matched components, got %d", result.TotalMatched)
				}
				for _, elem := range result.Results {
					if elem.Type != "component" {
						t.Errorf("expected component type, got %s", elem.Type)
					}
				}
			},
		},
		{
			name: "search by technology",
			request: entities.SearchElementsRequest{
				ProjectRoot: "/test/project",
				Query:       "*",
				Technology:  "Go",
				Limit:       10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.SearchElementsResponse) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				// Should find elements with Go technology
				if result.TotalMatched < 2 {
					t.Errorf("expected at least 2 total matched with Go tech, got %d", result.TotalMatched)
				}
			},
		},
		{
			name: "search by tag",
			request: entities.SearchElementsRequest{
				ProjectRoot: "/test/project",
				Query:       "*",
				Tag:         "core",
				Limit:       10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.SearchElementsResponse) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				// Should find elements with core tag
				if result.TotalMatched < 2 {
					t.Errorf("expected at least 2 total matched with core tag, got %d", result.TotalMatched)
				}
			},
		},
		{
			name: "no elements found",
			request: entities.SearchElementsRequest{
				ProjectRoot: "/test/project",
				Query:       "nonexistent",
				Limit:       10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.SearchElementsResponse) {
				if result.TotalMatched != 0 {
					t.Errorf("expected 0 total matched, got %d", result.TotalMatched)
				}
				if len(result.Results) != 0 {
					t.Errorf("expected 0 results, got %d", len(result.Results))
				}
				if result.Message != "No elements found matching query" {
					t.Errorf("expected 'No elements found matching query' message, got %q", result.Message)
				}
			},
		},
		{
			name:       "nil request validation",
			request:    entities.SearchElementsRequest{}, // Invalid: no ProjectRoot
			setupMocks: func(m *MockProjectRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProjectRepository{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			uc := NewSearchElements(mockRepo)
			result, err := uc.Execute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestSearchElementsMatchesElement tests the matchesElement helper function.
func TestSearchElementsMatchesElement(t *testing.T) {
	uc := NewSearchElements(&MockProjectRepository{})
	matcher := entities.NewGlobMatcher("*auth*")

	tests := []struct {
		name        string
		id          string
		nameVal     string
		elemType    string
		description string
		technology  string
		tags        []string
		request     entities.SearchElementsRequest
		expected    bool
	}{
		{
			name:        "match by ID",
			id:          "auth-handler",
			nameVal:     "Auth Handler",
			elemType:    "component",
			description: "Handles authentication",
			technology:  "Go",
			tags:        []string{"security"},
			request:     entities.SearchElementsRequest{Query: "*auth*"},
			expected:    true,
		},
		{
			name:        "match by name",
			id:          "auth-handler",
			nameVal:     "Auth Handler",
			elemType:    "component",
			description: "Handles authentication",
			technology:  "Go",
			tags:        []string{"security"},
			request:     entities.SearchElementsRequest{Query: "*auth*"},
			expected:    true,
		},
		{
			name:        "no match",
			id:          "db",
			nameVal:     "Database",
			elemType:    "component",
			description: "Database component",
			technology:  "PostgreSQL",
			tags:        []string{"storage"},
			request:     entities.SearchElementsRequest{Query: "*auth*"},
			expected:    false,
		},
		{
			name:        "match with technology filter",
			id:          "auth-handler",
			nameVal:     "Auth Handler",
			elemType:    "component",
			description: "Handles authentication",
			technology:  "Go",
			tags:        []string{"security"},
			request:     entities.SearchElementsRequest{Query: "*", Technology: "Go"},
			expected:    true,
		},
		{
			name:        "fail technology filter",
			id:          "auth-handler",
			nameVal:     "Auth Handler",
			elemType:    "component",
			description: "Handles authentication",
			technology:  "Go",
			tags:        []string{"security"},
			request:     entities.SearchElementsRequest{Query: "*", Technology: "Java"},
			expected:    false,
		},
		{
			name:        "match with tag filter",
			id:          "auth-handler",
			nameVal:     "Auth Handler",
			elemType:    "component",
			description: "Handles authentication",
			technology:  "Go",
			tags:        []string{"security", "core"},
			request:     entities.SearchElementsRequest{Query: "*", Tag: "core"},
			expected:    true,
		},
		{
			name:        "fail tag filter",
			id:          "auth-handler",
			nameVal:     "Auth Handler",
			elemType:    "component",
			description: "Handles authentication",
			technology:  "Go",
			tags:        []string{"security"},
			request:     entities.SearchElementsRequest{Query: "*", Tag: "database"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.matchesElement(matcher, tt.id, tt.nameVal, tt.elemType, tt.description, tt.technology, tt.tags, tt.request)
			if result != tt.expected {
				t.Errorf("matchesElement() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSearchElementsBuildMessage tests the buildMessage helper function.
func TestSearchElementsBuildMessage(t *testing.T) {
	uc := NewSearchElements(&MockProjectRepository{})

	tests := []struct {
		name         string
		totalMatched int
		returned     int
		request      entities.SearchElementsRequest
		expectedMsg  string
	}{
		{
			name:         "no elements found",
			totalMatched: 0,
			returned:     0,
			request:      entities.SearchElementsRequest{Query: "test"},
			expectedMsg:  "No elements found matching query",
		},
		{
			name:         "all elements returned",
			totalMatched: 5,
			returned:     5,
			request:      entities.SearchElementsRequest{Query: "*"},
			expectedMsg:  "Found 5 elements matching '*'",
		},
		{
			name:         "limited results",
			totalMatched: 10,
			returned:     5,
			request:      entities.SearchElementsRequest{Query: "*"},
			expectedMsg:  "Showing 5 of 10 matching elements (use limit parameter to adjust)",
		},
		{
			name:         "with type filter",
			totalMatched: 3,
			returned:     3,
			request:      entities.SearchElementsRequest{Query: "*", Type: "component"},
			expectedMsg:  "Found 3 elements matching '*' with filters: type=component",
		},
		{
			name:         "with technology filter",
			totalMatched: 2,
			returned:     2,
			request:      entities.SearchElementsRequest{Query: "*", Technology: "Go"},
			expectedMsg:  "Found 2 elements matching '*' with filters: technology=Go",
		},
		{
			name:         "with tag filter",
			totalMatched: 1,
			returned:     1,
			request:      entities.SearchElementsRequest{Query: "*", Tag: "core"},
			expectedMsg:  "Found 1 elements matching '*' with filters: tag=core",
		},
		{
			name:         "with multiple filters",
			totalMatched: 4,
			returned:     4,
			request:      entities.SearchElementsRequest{Query: "*", Type: "component", Technology: "Go", Tag: "core"},
			expectedMsg:  "Found 4 elements matching '*' with filters: type=component, technology=Go, tag=core",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := uc.buildMessage(tt.totalMatched, tt.returned, tt.request)
			if msg != tt.expectedMsg {
				t.Errorf("buildMessage() = %q, want %q", msg, tt.expectedMsg)
			}
		})
	}
}

// TestFormatMessage tests the formatMessage helper function.
func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple string format",
			format:   "Hello %s",
			args:     []interface{}{"World"},
			expected: "Hello World",
		},
		{
			name:     "integer format",
			format:   "Found %d items",
			args:     []interface{}{5},
			expected: "Found 5 items",
		},
		{
			name:     "mixed format",
			format:   "Found %d %s items",
			args:     []interface{}{3, "test"},
			expected: "Found 3 test items",
		},
		{
			name:     "no placeholders",
			format:   "No placeholders here",
			args:     []interface{}{},
			expected: "No placeholders here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMessage(tt.format, tt.args...)
			if result != tt.expected {
				t.Errorf("formatMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestIntToString tests the intToString helper function.
func TestIntToString(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{name: "zero", input: 0, expected: "0"},
		{name: "positive", input: 42, expected: "42"},
		{name: "negative", input: -15, expected: "-15"},
		{name: "large positive", input: 123456789, expected: "123456789"},
		{name: "large negative", input: -987654321, expected: "-987654321"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := intToString(tt.input)
			if result != tt.expected {
				t.Errorf("intToString(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
