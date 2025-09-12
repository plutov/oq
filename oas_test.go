package main

import (
	"context"
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

	details := formatSchemaDetails("User", userSchema.Value)

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

	details := formatSchemaDetails("User", userSchema.Value)

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
