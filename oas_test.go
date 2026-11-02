package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func parseSpec(t *testing.T, spec string) *v3.Document {
	t.Helper()
	doc, err := loadDocument(nil, strings.NewReader(spec))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestAllExampleFiles(t *testing.T) {
	files, err := os.ReadDir("examples")
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"examples/local-refs/openapi.yaml"}
	for _, file := range files {
		if !file.IsDir() {
			paths = append(paths, filepath.Join("examples", file.Name()))
		}
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			doc, err := loadDocument([]string{path}, nil)
			if err != nil {
				t.Fatal(err)
			}
			m := NewModel(doc)
			if len(m.items[viewEndpoints]) == 0 {
				t.Fatal("expected endpoints")
			}
			for _, mode := range m.availableViews() {
				m.mode = mode
				for i := range m.activeItems() {
					if !m.activeItems()[i].folded {
						t.Fatal("items must start folded")
					}
				}
				assertViewport(t, m)
				if len(m.activeItems()) > 0 {
					m.activeItems()[0].folded = false
					assertViewport(t, m)
				}
			}
			m.showHelp = true
			assertViewport(t, m)
		})
	}
}

func TestPetstoreRequestBodySchemaDisplay(t *testing.T) {
	doc, err := loadDocument([]string{"examples/petstore-3.0.yaml"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	entries := NewModel(doc).items[viewEndpoints]
	index := slices.IndexFunc(entries, func(entry item) bool {
		return entry.name == "/pet" && entry.kind == "POST"
	})
	if index < 0 {
		t.Fatal("missing POST /pet")
	}
	details := entries[index].details
	for _, expected := range []string{
		"Create a new pet in the store", "Required: true",
		"application/json (schema: Pet)", "application/xml (schema: Pet)",
		"application/x-www-form-urlencoded (schema: Pet)",
	} {
		if !strings.Contains(details, expected) {
			t.Errorf("missing %q in %s", expected, details)
		}
	}
	if !(strings.Index(details, "application/json") < strings.Index(details, "application/x-www-form-urlencoded") &&
		strings.Index(details, "application/x-www-form-urlencoded") < strings.Index(details, "application/xml")) {
		t.Error("media types are not sorted")
	}
}

const operationSpec = `openapi: 3.2.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets/{id}:
    parameters:
      - name: id
        in: path
        required: true
        description: inherited id
        schema:
          type: string
      - name: limit
        in: query
        description: inherited limit
        schema:
          type: integer
    post:
      summary: Create pet
      operationId: createPet
      description: Creates a pet
      parameters:
        - name: limit
          in: query
          description: operation limit
          schema:
            type: integer
      requestBody:
        description: Pet payload
        required: true
        content:
          text/plain:
            schema:
              type: [string, 'null']
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
      responses:
        '500':
          description: Failed
        '200':
          description: Created
        2XX:
          description: Success
        default:
          description: Fallback
    get:
      responses:
        '200':
          description: Retrieved
    query:
      responses:
        '200':
          description: Queried
    additionalOperations:
      COPY:
        responses:
          '200':
            description: Copied
  /a:
    delete:
      responses:
        '204':
          description: Deleted
webhooks:
  petAdded:
    post:
      operationId: petAdded
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
      responses:
        '200':
          description: Accepted
components:
  schemas:
    Pet:
      description: A pet
      type: object
      required: [name]
      properties:
        tags:
          type: array
          items:
            type: string
        name:
          type: [string, 'null']
  requestBodies:
    PetBody:
      description: Body description
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Pet'
  responses:
    PetResponse:
      description: Response description
      content:
        application/json:
          schema:
            type: object
      headers:
        Z-Rate:
          schema:
            type: integer
        A-Rate:
          schema:
            type: integer
  parameters:
    Limit:
      name: limit
      in: query
      description: Parameter description
      required: true
      schema:
        type: integer
        format: int32
      example: 10
  headers:
    Rate:
      description: Header description
      required: true
      schema:
        type: integer
        format: int64
  securitySchemes:
    Bearer:
      description: Security description
      type: http
      scheme: bearer
      bearerFormat: JWT
    ApiKey:
      type: apiKey
      in: header
      name: X-API-Key
`

func TestExtractOperations(t *testing.T) {
	doc := parseSpec(t, operationSpec)
	m := NewModel(doc)
	var names []string
	for _, entry := range m.items[viewEndpoints] {
		names = append(names, entry.kind+" "+entry.name)
	}
	want := []string{"DELETE /a", "COPY /pets/{id}", "GET /pets/{id}", "POST /pets/{id}", "QUERY /pets/{id}"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("operations = %v, want %v", names, want)
	}
	details := m.items[viewEndpoints][3].details
	for _, expected := range []string{
		"Summary: Create pet", "Description: Creates a pet", "Operation ID: createPet",
		"id (path): inherited id", "limit (query): operation limit", "Required: true",
		"application/json (schema: Pet)", "text/plain (types: [string null])",
		"  - 200: Created\n  - 500: Failed\n  - 2XX: Success\n  - default: Fallback\n",
	} {
		if !strings.Contains(details, expected) {
			t.Errorf("missing %q in %s", expected, details)
		}
	}
	if strings.Contains(details, "inherited limit") || strings.Count(details, "limit (query)") != 1 {
		t.Error("operation parameters must override path parameters")
	}
	if len(m.items[viewWebhooks]) != 1 {
		t.Fatalf("webhooks = %v", m.items[viewWebhooks])
	}
	hook := m.items[viewWebhooks][0]
	if hook.kind != "POST" || hook.name != "petAdded" ||
		!strings.Contains(hook.details, "Operation ID: petAdded") ||
		!strings.Contains(hook.details, "application/json (schema: Pet)") ||
		!strings.Contains(hook.details, "200: Accepted") {
		t.Errorf("incomplete webhook: %+v", hook)
	}
}

func TestExtractComponents(t *testing.T) {
	entries := extractComponents(parseSpec(t, operationSpec))
	wantKinds := []string{"Header", "Parameter", "RequestBody", "Response", "Schema", "SecurityScheme", "SecurityScheme"}
	var kinds []string
	for _, entry := range entries {
		kinds = append(kinds, entry.kind)
	}
	if !reflect.DeepEqual(kinds, wantKinds) {
		t.Fatalf("kinds = %v", kinds)
	}
	wantDetails := []string{
		"Required: true\nType: integer\nFormat: int64\n",
		"In: query\nRequired: true\nType: integer\nFormat: int32\nExample:",
		"Required: true\nContent Types:\n  - application/json (schema: Pet)\n",
		"Content Types:\n  - application/json (type: object)\nHeaders:\n  - A-Rate\n  - Z-Rate\n",
		"Type: object\nRequired: [name]\nProperties:\n  - name: [string null]\n  - tags: array\n",
		"Type: apiKey\nIn: header\nName: X-API-Key\n",
		"Type: http\nScheme: bearer\nBearer Format: JWT\n",
	}
	for i, want := range wantDetails {
		if !strings.HasPrefix(entries[i].details, want) {
			t.Errorf("%s details = %q, want %q", entries[i].kind, entries[i].details, want)
		}
		if !entries[i].folded {
			t.Error("components must start folded")
		}
	}
	if entries[4].description != "A pet" || entries[0].description != "Header description" {
		t.Error("component descriptions lost")
	}
}

func TestSortResponseCodes(t *testing.T) {
	codes := []string{"default", "500", "2XX", "20", "200", "100", "4XX"}
	sortResponseCodes(codes)
	want := []string{"20", "100", "200", "500", "2XX", "4XX", "default"}
	if !reflect.DeepEqual(codes, want) {
		t.Errorf("codes = %v, want %v", codes, want)
	}
}

func TestMissingDetails(t *testing.T) {
	for _, details := range []string{
		formatSchemaDetails(nil), formatRequestBodyDetails(nil), formatResponseDetails(nil),
		formatParameterDetails(nil), formatHeaderDetails(nil), formatSecuritySchemeDetails(nil),
	} {
		if !strings.HasPrefix(details, "No ") {
			t.Errorf("unexpected fallback: %q", details)
		}
	}
	paths := orderedmap.New[string, *v3.PathItem]()
	paths.Set("/nil", nil)
	paths.Set("/empty", &v3.PathItem{})
	if len(extractOperations(paths)) != 0 || len(extractOperations(nil)) != 0 {
		t.Error("empty paths must produce no operations")
	}
}
