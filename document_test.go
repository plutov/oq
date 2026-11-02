package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const minimalSpec = `{"openapi":"3.0.3","info":{"title":"Minimal API","version":"1.0.0"},"paths":{}}`

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestLoadDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "openapi.json")
	if err := os.WriteFile(path, []byte(minimalSpec), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{nil, {path}} {
		doc, err := loadDocument(args, strings.NewReader(minimalSpec))
		if err != nil {
			t.Fatal(err)
		}
		if doc.Info.Title != "Minimal API" || doc.Info.Version != "1.0.0" {
			t.Errorf("unexpected info: %+v", doc.Info)
		}
	}
}

func TestLoadDocumentErrors(t *testing.T) {
	readError := errors.New("reader failed")
	for _, test := range []struct {
		name  string
		args  []string
		input io.Reader
		want  string
	}{
		{"arguments", []string{"a", "b"}, nil, "usage:"},
		{"file", []string{filepath.Join(t.TempDir(), "missing.yaml")}, nil, "read spec:"},
		{"stdin", nil, errorReader{readError}, "read stdin:"},
		{"empty", nil, strings.NewReader(""), "parse spec:"},
		{"malformed", nil, strings.NewReader("not: [valid"), "parse spec:"},
		{"unsupported", nil, strings.NewReader(`{"swagger":"2.0","info":{"title":"Old","version":"1"},"paths":{}}`), "build OpenAPI model:"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadDocument(test.args, test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			if test.name == "stdin" && !errors.Is(err, readError) {
				t.Error("reader error was not preserved")
			}
		})
	}
}

func TestLocalReferences(t *testing.T) {
	path, err := filepath.Abs("examples/local-refs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	doc, err := loadDocument([]string{path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(doc)
	if len(m.items[viewEndpoints]) == 0 || len(m.items[viewComponents]) == 0 {
		t.Fatal("local references did not resolve outside working directory")
	}
	for _, entry := range m.items[viewComponents] {
		if entry.kind == "Schema" && !strings.Contains(entry.details, "Properties:") {
			t.Errorf("schema reference was not expanded: %+v", entry)
		}
	}
}

const catalog32JSON = `{
  "openapi": "3.2.0",
  "info": {"title": "Catalog API", "version": "1.0.0"},
  "paths": {
    "/products/{id}": {
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}],
      "query": {
        "operationId": "queryProduct",
        "responses": {"200": {"description": "Matching product", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Product"}}}, "headers": {"X-Request-ID": {"schema": {"type": "string", "format": "uuid"}}}}}
      },
      "additionalOperations": {
        "COPY": {"operationId": "copyProduct", "responses": {"201": {"description": "Product copied"}}}
      }
    }
  },
  "webhooks": {
    "productChanged": {"post": {"operationId": "productChanged", "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Product"}}}}, "responses": {"204": {"description": "Notification accepted"}}}}
  },
  "components": {
    "schemas": {"Product": {"description": "A catalog product", "type": "object", "required": ["id", "name"], "properties": {"id": {"type": "string", "format": "uuid"}, "name": {"type": "string"}, "discontinuedAt": {"type": ["string", "null"], "format": "date-time"}}}}
  }
}`

func TestOpenAPI32YAMLAndJSON(t *testing.T) {
	yaml, err := os.ReadFile("examples/catalog-3.2.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for format, spec := range map[string]string{"yaml": string(yaml), "json": catalog32JSON} {
		t.Run(format, func(t *testing.T) {
			doc := parseSpec(t, spec)
			if doc.Version != "3.2.0" || doc.Info.Title != "Catalog API" {
				t.Fatalf("lost 3.2 document metadata: %s, %s", doc.Version, doc.Info.Title)
			}
			path := doc.Paths.PathItems.GetOrZero("/products/{id}")
			if path == nil || path.Query == nil || path.AdditionalOperations.GetOrZero("COPY") == nil {
				t.Fatal("parser lost QUERY or additionalOperations")
			}
			if doc.Webhooks.GetOrZero("productChanged") == nil || doc.Components.Schemas.GetOrZero("Product").Schema().Properties.Len() != 3 {
				t.Fatal("parser lost webhook or schema properties")
			}
			m := NewModel(doc)
			endpoints := m.items[viewEndpoints]
			if len(endpoints) != 2 || endpoints[0].kind != "COPY" || endpoints[1].kind != "QUERY" || endpoints[1].name != "/products/{id}" {
				t.Fatalf("incorrect 3.2 endpoints: %+v", endpoints)
			}
			for _, entry := range endpoints {
				if !entry.folded || !strings.Contains(entry.details, "id (path):") || !strings.Contains(entry.details, "Required: true\n    Type: string\n    Format: uuid") {
					t.Fatalf("3.2 operation lost inherited parameter or collapsed state: %+v", entry)
				}
			}
			for _, text := range []string{"Operation ID: queryProduct", "200: Matching product", "application/json (schema: Product)", "X-Request-ID", "Format: uuid"} {
				if !strings.Contains(endpoints[1].details, text) {
					t.Errorf("QUERY details missing %q: %s", text, endpoints[1].details)
				}
			}
			if !strings.Contains(endpoints[0].details, "Operation ID: copyProduct") || !strings.Contains(endpoints[0].details, "201: Product copied") {
				t.Fatal("COPY operation details lost")
			}
			if len(m.items[viewWebhooks]) != 1 || m.items[viewWebhooks][0].name != "productChanged" || !strings.Contains(m.items[viewWebhooks][0].details, "application/json (schema: Product)") {
				t.Fatal("webhook view lost payload")
			}
			if len(m.items[viewComponents]) != 1 || m.items[viewComponents][0].name != "Product" || !strings.Contains(m.items[viewComponents][0].details, "discontinuedAt: [string null]") {
				t.Fatal("component view lost 3.2 nullable property")
			}
			for _, mode := range m.availableViews() {
				m.mode = mode
				for _, entry := range m.activeItems() {
					if !entry.folded || strings.Contains(m.renderItems(), "Operation ID:") {
						t.Fatal("3.2 entities should start collapsed")
					}
				}
				m, _ = press(m, "enter")
				assertViewport(t, m)
			}
		})
	}
}
