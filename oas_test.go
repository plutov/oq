package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Test data for OpenAPI 3.0
const openapi30Spec = `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
  description: A test API for OpenAPI 3.0
  contact:
    name: Test Support
    email: test@example.com
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT
paths:
  /users:
    get:
      summary: List users
      description: Get a list of users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUser'
      responses:
        '201':
          description: Created
  /users/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      summary: Get user by ID
      responses:
        '200':
          description: Success
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
          format: email
    CreateUser:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        email:
          type: string
          format: email
  parameters:
    UserIdParam:
      name: userId
      in: path
      required: true
      schema:
        type: string
  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                type: string
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
`

// Test data for OpenAPI 3.1
const openapi31Spec = `
openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
  description: A test API for OpenAPI 3.1
  contact:
    name: Test Support
    email: test@example.com
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT
paths:
  /users:
    get:
      summary: List users
      description: Get a list of users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
          format: email
    CreateUser:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        email:
          type: string
          format: email
`

func TestLoadOpenAPI30(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi30Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.0 spec: %v", err)
	}

	if doc.OpenAPI != "3.0.3" {
		t.Errorf("Expected OpenAPI version 3.0.3, got %s", doc.OpenAPI)
	}

	if doc.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", doc.Info.Title)
	}

	// Test validation
	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("OpenAPI 3.0 spec validation failed: %v", err)
	}
}

func TestLoadOpenAPI31(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi31Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.1 spec: %v", err)
	}

	if doc.OpenAPI != "3.1.0" {
		t.Errorf("Expected OpenAPI version 3.1.0, got %s", doc.OpenAPI)
	}

	if doc.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", doc.Info.Title)
	}

	// Test validation
	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("OpenAPI 3.1 spec validation failed: %v", err)
	}
}

func TestExtractEndpoints30(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi30Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.0 spec: %v", err)
	}

	endpoints := extractEndpoints(doc)

	expectedEndpoints := 3 // GET /users, POST /users, GET /users/{id}
	if len(endpoints) != expectedEndpoints {
		t.Errorf("Expected %d endpoints, got %d", expectedEndpoints, len(endpoints))
	}

	// Check for specific endpoints
	foundGetUsers := false
	foundPostUsers := false
	foundGetUserById := false

	for _, ep := range endpoints {
		if ep.path == "/users" && ep.method == "GET" {
			foundGetUsers = true
		}
		if ep.path == "/users" && ep.method == "POST" {
			foundPostUsers = true
		}
		if ep.path == "/users/{id}" && ep.method == "GET" {
			foundGetUserById = true
		}
	}

	if !foundGetUsers {
		t.Error("Expected GET /users endpoint not found")
	}
	if !foundPostUsers {
		t.Error("Expected POST /users endpoint not found")
	}
	if !foundGetUserById {
		t.Error("Expected GET /users/{id} endpoint not found")
	}
}

func TestExtractEndpoints31(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi31Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.1 spec: %v", err)
	}

	endpoints := extractEndpoints(doc)

	expectedEndpoints := 1 // Only GET /users in 3.1 spec
	if len(endpoints) != expectedEndpoints {
		t.Errorf("Expected %d endpoints, got %d", expectedEndpoints, len(endpoints))
	}

	// Check for specific endpoint
	foundGetUsers := false

	for _, ep := range endpoints {
		if ep.path == "/users" && ep.method == "GET" {
			foundGetUsers = true
		}
	}

	if !foundGetUsers {
		t.Error("Expected GET /users endpoint not found")
	}
}

func TestExtractComponents30(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi30Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.0 spec: %v", err)
	}

	components := extractComponents(doc)

	// Should have schemas, parameters, responses, and security schemes
	expectedMinComponents := 5 // 2 schemas + 1 parameter + 1 response + 1 security scheme
	if len(components) < expectedMinComponents {
		t.Errorf("Expected at least %d components, got %d", expectedMinComponents, len(components))
	}

	// Check for specific components
	foundUserSchema := false
	foundSecurityScheme := false

	for _, comp := range components {
		if comp.name == "User" && comp.compType == "Schema" {
			foundUserSchema = true
		}
		if comp.name == "BearerAuth" && comp.compType == "SecurityScheme" {
			foundSecurityScheme = true
		}
	}

	if !foundUserSchema {
		t.Error("Expected User schema component not found")
	}
	if !foundSecurityScheme {
		t.Error("Expected BearerAuth security scheme component not found")
	}
}

func TestFormatSchemaDetails30(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi30Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.0 spec: %v", err)
	}

	userSchema := doc.Components.Schemas["User"]
	if userSchema == nil {
		t.Fatal("User schema not found")
	}

	details := formatSchemaDetails(userSchema.Value)

	if !strings.Contains(details, "Type: object") {
		t.Error("Expected 'Type: object' in schema details")
	}
	if !strings.Contains(details, "Required: [id name]") {
		t.Error("Expected required fields in schema details")
	}
	if !strings.Contains(details, "Properties:") {
		t.Error("Expected properties section in schema details")
	}
}

func TestFormatSchemaDetails31(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi31Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.1 spec: %v", err)
	}

	userSchema := doc.Components.Schemas["User"]
	if userSchema == nil {
		t.Fatal("User schema not found")
	}

	details := formatSchemaDetails(userSchema.Value)

	if !strings.Contains(details, "Type: object") {
		t.Error("Expected 'Type: object' in schema details")
	}
	if !strings.Contains(details, "id:") {
		t.Error("Expected 'id' property in schema details")
	}
}

func TestFormatEndpointDetails(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi30Spec))
	if err != nil {
		t.Fatalf("Failed to load OpenAPI 3.0 spec: %v", err)
	}

	endpoints := extractEndpoints(doc)
	var getUsersEndpoint endpoint

	for _, ep := range endpoints {
		if ep.path == "/users" && ep.method == "GET" {
			getUsersEndpoint = ep
			break
		}
	}

	if getUsersEndpoint.op == nil {
		t.Fatal("GET /users endpoint not found")
	}

	details := formatEndpointDetails(getUsersEndpoint)

	if !strings.Contains(details, "Summary: List users") {
		t.Error("Expected summary in endpoint details")
	}
	if !strings.Contains(details, "Description: Get a list of users") {
		t.Error("Expected description in endpoint details")
	}
	if !strings.Contains(details, "Responses:") {
		t.Error("Expected responses section in endpoint details")
	}
}

// JSON test data for OpenAPI 3.0
const openapi30JSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Test API JSON",
    "version": "1.0.0",
    "description": "A test API for JSON format"
  },
  "paths": {
    "/users": {
      "get": {
        "summary": "List users",
        "responses": {
          "200": {
            "description": "Success",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/User"
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "User": {
        "type": "object",
        "required": ["id", "name"],
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"}
        }
      }
    }
  }
}`

// JSON test data for OpenAPI 3.1
const openapi31JSON = `{
  "openapi": "3.1.0",
  "info": {
    "title": "Test API JSON 3.1",
    "version": "1.0.0",
    "description": "A test API for JSON format with OpenAPI 3.1"
  },
  "paths": {
    "/products": {
      "get": {
        "summary": "List products",
        "responses": {
          "200": {
            "description": "Success",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Product"
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Product": {
        "type": "object",
        "required": ["id", "name"],
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"},
          "price": {
            "type": "number",
            "minimum": 0,
            "exclusiveMinimum": true
          }
        }
      }
    }
  }
}`

func TestLoadJSONFormat(t *testing.T) {
	tests := []struct {
		name     string
		jsonSpec string
		version  string
		title    string
	}{
		{
			name:     "OpenAPI 3.0 JSON",
			jsonSpec: openapi30JSON,
			version:  "3.0.3",
			title:    "Test API JSON",
		},
		{
			name:     "OpenAPI 3.1 JSON",
			jsonSpec: openapi31JSON,
			version:  "3.1.0",
			title:    "Test API JSON 3.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(tt.jsonSpec))
			if err != nil {
				t.Fatalf("Failed to load JSON spec: %v", err)
			}

			if doc.OpenAPI != tt.version {
				t.Errorf("Expected OpenAPI version %s, got %s", tt.version, doc.OpenAPI)
			}

			if doc.Info.Title != tt.title {
				t.Errorf("Expected title '%s', got %s", tt.title, doc.Info.Title)
			}

			// Test validation
			err = doc.Validate(context.Background())
			if err != nil {
				t.Fatalf("JSON spec validation failed: %v", err)
			}

			// Test that we can extract endpoints
			endpoints := extractEndpoints(doc)
			if len(endpoints) == 0 {
				t.Error("No endpoints extracted from JSON spec")
			}

			// Test that we can extract components
			components := extractComponents(doc)
			if len(components) == 0 {
				t.Error("No components extracted from JSON spec")
			}
		})
	}
}

func TestBothVersionsWithSameLogic(t *testing.T) {
	specs := map[string]string{
		"3.0.3": openapi30Spec,
		"3.1.0": openapi31Spec,
	}

	for version, spec := range specs {
		t.Run("version_"+version, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(spec))
			if err != nil {
				t.Fatalf("Failed to load OpenAPI %s spec: %v", version, err)
			}

			// Both versions should be parseable
			if doc.OpenAPI != version {
				t.Errorf("Expected OpenAPI version %s, got %s", version, doc.OpenAPI)
			}

			// Both should validate
			err = doc.Validate(context.Background())
			if err != nil {
				t.Fatalf("OpenAPI %s spec validation failed: %v", version, err)
			}

			// Both should extract endpoints
			endpoints := extractEndpoints(doc)
			if len(endpoints) == 0 {
				t.Errorf("No endpoints extracted for version %s", version)
			}

			// Both should extract components
			components := extractComponents(doc)
			if len(components) == 0 {
				t.Errorf("No components extracted for version %s", version)
			}
		})
	}
}

func TestFormatCompatibility(t *testing.T) {
	formats := map[string]string{
		"YAML": openapi30Spec,
		"JSON": openapi30JSON,
	}

	for format, spec := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(spec))
			if err != nil {
				t.Fatalf("Failed to load %s spec: %v", format, err)
			}

			// Both formats should work with same logic
			endpoints := extractEndpoints(doc)
			components := extractComponents(doc)

			if len(endpoints) == 0 {
				t.Errorf("No endpoints extracted for %s format", format)
			}

			if len(components) == 0 {
				t.Errorf("No components extracted for %s format", format)
			}

			// Validation should work for both
			err = doc.Validate(context.Background())
			if err != nil {
				t.Fatalf("%s spec validation failed: %v", format, err)
			}
		})
	}
}

func TestSortResponseCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "mixed numeric and text codes",
			input:    []string{"500", "200", "default", "404", "201"},
			expected: []string{"200", "201", "404", "500", "default"},
		},
		{
			name:     "all numeric codes",
			input:    []string{"500", "200", "404", "201", "400"},
			expected: []string{"200", "201", "400", "404", "500"},
		},
		{
			name:     "all text codes",
			input:    []string{"default", "error", "abc"},
			expected: []string{"abc", "default", "error"},
		},
		{
			name:     "single code",
			input:    []string{"200"},
			expected: []string{"200"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "from example.json order",
			input:    []string{"400", "500", "200"},
			expected: []string{"200", "400", "500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes := make([]string, len(tt.input))
			copy(codes, tt.input)

			sortResponseCodes(codes)

			if len(codes) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(codes))
				return
			}

			for i, expected := range tt.expected {
				if codes[i] != expected {
					t.Errorf("At index %d: expected %s, got %s", i, expected, codes[i])
				}
			}
		})
	}
}

func TestResponseOrderingStability(t *testing.T) {
	// Create a test spec with responses in different order
	testSpec := `{
		"openapi": "3.0.3",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"500": {"description": "Internal server error"},
						"200": {"description": "Success"},
						"400": {"description": "Bad request"},
						"404": {"description": "Not found"},
						"default": {"description": "Default response"}
					}
				}
			}
		}
	}`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(testSpec))
	if err != nil {
		t.Fatalf("Failed to load test spec: %v", err)
	}

	endpoints := extractEndpoints(doc)
	if len(endpoints) != 1 {
		t.Fatalf("Expected 1 endpoint, got %d", len(endpoints))
	}

	// Format details multiple times and ensure order is consistent
	details1 := formatEndpointDetails(endpoints[0])
	details2 := formatEndpointDetails(endpoints[0])
	details3 := formatEndpointDetails(endpoints[0])

	if details1 != details2 || details2 != details3 {
		t.Error("Response ordering is not stable across multiple calls")
		t.Logf("Details1:\n%s", details1)
		t.Logf("Details2:\n%s", details2)
		t.Logf("Details3:\n%s", details3)
	}

	// Check that the expected order appears in the output
	expectedOrder := []string{"200", "400", "404", "500", "default"}
	for i := 0; i < len(expectedOrder)-1; i++ {
		current := expectedOrder[i]
		next := expectedOrder[i+1]

		currentIndex := strings.Index(details1, current+":")
		nextIndex := strings.Index(details1, next+":")

		if currentIndex == -1 {
			t.Errorf("Response code %s not found in details", current)
			continue
		}
		if nextIndex == -1 {
			t.Errorf("Response code %s not found in details", next)
			continue
		}

		if currentIndex > nextIndex {
			t.Errorf("Response code %s should appear before %s, but found at positions %d and %d",
				current, next, currentIndex, nextIndex)
		}
	}
}

func TestExampleJSONResponseOrdering(t *testing.T) {
	content, err := os.ReadFile("examples/petstore-3.0.yaml")
	if err != nil {
		t.Errorf("Failed to read petstore-3.0.yaml: %v", err)
		return
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(content)
	if err != nil {
		t.Fatalf("Failed to load petstore example: %v", err)
	}

	endpoints := extractEndpoints(doc)

	// Find the PUT /pet endpoint which has multiple response codes
	var putPetEndpoint *endpoint
	for i := range endpoints {
		if endpoints[i].path == "/pet" && endpoints[i].method == "PUT" {
			putPetEndpoint = &endpoints[i]
			break
		}
	}

	if putPetEndpoint == nil {
		t.Fatal("PUT /pet endpoint not found in petstore-3.0.yaml")
	}

	// Format the endpoint details multiple times
	details1 := formatEndpointDetails(*putPetEndpoint)
	details2 := formatEndpointDetails(*putPetEndpoint)

	// Verify they are identical (stable ordering)
	if details1 != details2 {
		t.Error("Response ordering is not stable for petstore example")
	}

	// Verify numeric codes appear before text codes
	// PUT /pet has: 200, 400, 404, 422, default
	expectedCodes := []string{"200", "400", "404", "422", "default"}

	// Find positions of each response code in the formatted output
	positions := make(map[string]int)
	for _, code := range expectedCodes {
		pos := strings.Index(details1, "- "+code+":")
		if pos == -1 {
			t.Errorf("Response code %s not found in formatted details", code)
			continue
		}
		positions[code] = pos
	}

	// Verify that codes appear in the correct order
	for i := 0; i < len(expectedCodes)-1; i++ {
		current := expectedCodes[i]
		next := expectedCodes[i+1]
		if positions[current] > positions[next] {
			t.Errorf("Response %s should appear before %s", current, next)
		}
	}

	t.Logf("Response order verified for PUT /pet endpoint")
}

// Test OpenAPI 3.1 to 3.0 conversion
func TestOpenAPI31To30Conversion(t *testing.T) {
	// Test with OpenAPI 3.1 spec that needs conversion
	openapi31WithExamples := `{
		"openapi": "3.1.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"200": {"description": "Success"}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Product": {
					"type": "object",
					"properties": {
						"id": {
							"type": "string",
							"examples": ["123", "456"]
						},
						"price": {
							"type": "number",
							"exclusiveMinimum": 0
						}
					}
				}
			}
		}
	}`

	// Test conversion
	converted, err := convertOpenAPI31To30([]byte(openapi31WithExamples))
	if err != nil {
		t.Fatalf("Failed to convert OpenAPI 3.1 to 3.0: %v", err)
	}

	// Parse the converted spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(converted)
	if err != nil {
		t.Fatalf("Failed to load converted spec: %v", err)
	}

	// Verify it validates
	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("Converted spec validation failed: %v", err)
	}
}

// Test with OpenAPI 3.0 spec (should pass through unchanged)
func TestOpenAPI30Passthrough(t *testing.T) {
	openapi30Spec := `{
		"openapi": "3.0.3",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"200": {"description": "Success"}
					}
				}
			}
		}
	}`

	converted, err := convertOpenAPI31To30([]byte(openapi30Spec))
	if err != nil {
		t.Fatalf("Failed to process OpenAPI 3.0 spec: %v", err)
	}

	// Should be identical to input
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(converted)
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}

	if doc.OpenAPI != "3.0.3" {
		t.Errorf("Expected version to remain 3.0.3, got %s", doc.OpenAPI)
	}
}

// Test all example files
func TestAllExampleFiles(t *testing.T) {
	exampleFiles := []string{
		"examples/petstore-3.0.yaml",
		"examples/petstore-3.1.yaml",
		"examples/train-travel-3.1.json",
		"examples/train-travel-3.1.yaml",
	}

	for _, filename := range exampleFiles {
		t.Run(filename, func(t *testing.T) {
			content, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read %s: %v", filename, err)
				return
			}

			// Try to convert (will pass through if already 3.0)
			converted, err := convertOpenAPI31To30(content)
			if err != nil {
				// If conversion fails, try with original content
				converted = content
			}

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true

			doc, err := loader.LoadFromData(converted)
			if err != nil {
				t.Fatalf("Failed to load %s: %v", filename, err)
			}

			// Validate the spec
			err = doc.Validate(loader.Context)
			if err != nil {
				t.Fatalf("Validation failed for %s: %v", filename, err)
			}

			// Test endpoint extraction
			endpoints := extractEndpoints(doc)
			if len(endpoints) == 0 {
				t.Errorf("No endpoints found in %s", filename)
			}

			// Test component extraction
			components := extractComponents(doc)
			// Note: some specs might not have components, so we just check that extraction doesn't crash

			// Test endpoint formatting
			for _, ep := range endpoints {
				details := formatEndpointDetails(ep)
				if details == "" {
					t.Errorf("Empty details for endpoint %s %s in %s", ep.method, ep.path, filename)
				}
			}

			// Test component formatting
			for _, comp := range components {
				// Just verify that formatting doesn't crash
				switch comp.compType {
				case "Schema":
					if len(comp.details) == 0 {
						t.Errorf("Empty details for schema %s in %s", comp.name, filename)
					}
				}
			}

			t.Logf("Successfully processed %s with %d endpoints and %d components",
				filename, len(endpoints), len(components))
		})
	}
}

// Test converter with YAML input
func TestConverterWithYAML(t *testing.T) {
	yamlSpec := `
openapi: 3.1.0
info:
  title: YAML Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: Success
components:
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
          examples: 
            - "test-id"
        status:
          const: "active"
`

	converted, err := convertOpenAPI31To30([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("Failed to convert YAML spec: %v", err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(converted)
	if err != nil {
		t.Fatalf("Failed to load converted YAML spec: %v", err)
	}

	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("Converted YAML spec validation failed: %v", err)
	}
}
