package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestAllExampleFiles(t *testing.T) {
	examplesDir := "examples"

	files, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(filename, ".yaml") &&
			!strings.HasSuffix(filename, ".yml") &&
			!strings.HasSuffix(filename, ".json") {
			continue
		}

		t.Run(filename, func(t *testing.T) {
			filepath := filepath.Join(examplesDir, filename)
			testExampleFile(t, filepath)
		})
	}
}

func testExampleFile(t *testing.T, filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filepath, err)
	}

	document, err := libopenapi.NewDocument(content)
	if err != nil {
		t.Fatalf("Error creating document from %s: %v", filepath, err)
	}

	v3Model, err := document.BuildV3Model()
	if err != nil {
		t.Fatalf("Error building v3 model from %s: %v", filepath, err)
	}

	if v3Model == nil {
		t.Fatalf("V3 model is nil for %s", filepath)
	}

	model := NewModel(&v3Model.Model)

	if model.doc == nil {
		t.Fatalf("Model document is nil for %s", filepath)
	}

	endpoints := model.endpoints

	for i, ep := range endpoints {
		details := formatEndpointDetails(ep)
		if details == "" {
			t.Errorf("Empty endpoint details for endpoint %d (%s %s) in %s",
				i, ep.method, ep.path, filepath)
		}
	}

	components := model.components

	emptyDetailsCount := 0
	for _, comp := range components {
		if comp.details == "" {
			emptyDetailsCount++
		}
	}

	webhooks := model.webhooks

	emptyWebhookCount := 0
	for _, hook := range webhooks {
		details := formatWebhookDetails(hook)
		if details == "" {
			emptyWebhookCount++
		}
	}

	testModelRendering(t, &model, filepath)
}

func testModelRendering(t *testing.T, model *Model, filepath string) {
	model.width = 120
	model.height = 40

	model.mode = viewEndpoints
	endpointsView := model.View()
	if endpointsView == "" {
		t.Errorf("Empty endpoints view for %s", filepath)
	}

	model.mode = viewComponents
	componentsView := model.View()
	if componentsView == "" {
		t.Errorf("Empty components view for %s", filepath)
	}

	if len(model.webhooks) > 0 {
		model.mode = viewWebhooks
		webhooksView := model.View()
		if webhooksView == "" {
			t.Errorf("Empty webhooks view for %s", filepath)
		}
	}

	model.showHelp = true
	helpView := model.View()
	if helpView == "" {
		t.Errorf("Empty help view for %s", filepath)
	}
	model.showHelp = false

	if len(model.endpoints) > 0 {
		model.mode = viewEndpoints
		model.cursor = 0

		model.endpoints[0].folded = true
		foldedView := model.View()
		if foldedView == "" {
			t.Errorf("Empty folded endpoints view for %s", filepath)
		}

		model.endpoints[0].folded = false
		unfoldedView := model.View()
		if unfoldedView == "" {
			t.Errorf("Empty unfolded endpoints view for %s", filepath)
		}
	}

	if len(model.components) > 0 {
		model.mode = viewComponents
		model.cursor = 0

		model.components[0].folded = true
		foldedView := model.View()
		if foldedView == "" {
			t.Errorf("Empty folded components view for %s", filepath)
		}

		model.components[0].folded = false
		unfoldedView := model.View()
		if unfoldedView == "" {
			t.Errorf("Empty unfolded components view for %s", filepath)
		}
	}
}

func TestSpecificExampleFiles(t *testing.T) {
	specificTests := []struct {
		filename      string
		minEndpoints  int
		minComponents int
	}{
		{"petstore-3.0.yaml", 1, 1},
		{"petstore-3.1.yaml", 1, 1},
	}

	for _, test := range specificTests {
		t.Run(test.filename, func(t *testing.T) {
			filepath := filepath.Join("examples", test.filename)

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Skipf("File %s does not exist, skipping", filepath)
				return
			}

			content, err := os.ReadFile(filepath)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", filepath, err)
			}

			document, err := libopenapi.NewDocument(content)
			if err != nil {
				t.Fatalf("Error creating document from %s: %v", filepath, err)
			}

			v3Model, err := document.BuildV3Model()
			if err != nil {
				t.Fatalf("Error building v3 model from %s: %v", filepath, err)
			}

			model := NewModel(&v3Model.Model)

			if len(model.endpoints) < test.minEndpoints {
				t.Errorf("Expected at least %d endpoints in %s, got %d",
					test.minEndpoints, test.filename, len(model.endpoints))
			}

			if len(model.components) < test.minComponents {
				t.Errorf("Expected at least %d components in %s, got %d",
					test.minComponents, test.filename, len(model.components))
			}

			if model.doc.Info == nil {
				t.Errorf("Document info is nil for %s", test.filename)
			}
		})
	}
}

func TestModelNavigation(t *testing.T) {
	filepath := "examples/petstore-3.0.yaml"
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Skip("petstore-3.0.yaml not found, skipping navigation test")
		return
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read petstore: %v", err)
	}

	document, err := libopenapi.NewDocument(content)
	if err != nil {
		t.Fatalf("Error creating document: %v", err)
	}

	v3Model, err := document.BuildV3Model()
	if err != nil {
		t.Fatalf("Error building v3 model: %v", err)
	}

	model := NewModel(&v3Model.Model)
	model.width = 120
	model.height = 40

	initialCursor := model.cursor

	if len(model.endpoints) > 1 {
		model.cursor = 1
		model.ensureCursorVisible()
		if model.cursor != 1 {
			t.Errorf("Cursor should be 1, got %d", model.cursor)
		}
	}

	model.cursor = 0
	model.ensureCursorVisible()
	if model.cursor != 0 {
		t.Errorf("Cursor should be 0, got %d", model.cursor)
	}

	originalMode := model.mode

	model.mode = viewComponents
	model.cursor = 0
	componentsView := model.View()
	if componentsView == "" {
		t.Error("Components view should not be empty")
	}

	model.mode = originalMode
	model.cursor = initialCursor
	endpointsView := model.View()
	if endpointsView == "" {
		t.Error("Endpoints view should not be empty")
	}
}

func TestPetstoreRequestBodySchemaDisplay(t *testing.T) {
	filepath := "examples/petstore-3.0.yaml"
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Skip("petstore-3.0.yaml not found, skipping test")
		return
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read petstore: %v", err)
	}

	document, err := libopenapi.NewDocument(content)
	if err != nil {
		t.Fatalf("Error creating document: %v", err)
	}

	v3Model, err := document.BuildV3Model()
	if err != nil {
		t.Fatalf("Error building v3 model: %v", err)
	}

	model := NewModel(&v3Model.Model)

	var addPetEndpoint *endpoint
	for i := range model.endpoints {
		if model.endpoints[i].path == "/pet" && model.endpoints[i].method == "POST" {
			addPetEndpoint = &model.endpoints[i]
			break
		}
	}

	if addPetEndpoint == nil {
		t.Fatal("Could not find POST /pet endpoint")
	}

	details := formatEndpointDetails(*addPetEndpoint)

	// Verify request body shows description
	if !strings.Contains(details, "Create a new pet in the store") {
		t.Error("Request body description not displayed")
	}

	// Verify required status is shown
	if !strings.Contains(details, "Required: true") {
		t.Error("Request body required status not displayed")
	}

	// Verify schema references are shown for all media types
	expectedSchemas := []string{
		"application/json (schema: Pet)",
		"application/xml (schema: Pet)",
		"application/x-www-form-urlencoded (schema: Pet)",
	}

	for _, expected := range expectedSchemas {
		if !strings.Contains(details, expected) {
			t.Errorf("Expected schema reference not found: %s", expected)
		}
	}
	// Verify media types are sorted alphabetically
	jsonIdx := strings.Index(details, "application/json")
	xmlIdx := strings.Index(details, "application/xml")
	formIdx := strings.Index(details, "application/x-www-form-urlencoded")

	if jsonIdx == -1 || xmlIdx == -1 || formIdx == -1 {
		t.Fatal("Not all media types found in output")
	}

	if !(jsonIdx < formIdx && formIdx < xmlIdx) {
		t.Error("Media types are not sorted alphabetically")
	}
}
func TestEmptyDocument(t *testing.T) {
	minimalSpec := `{
		"openapi": "3.0.3",
		"info": {
			"title": "Minimal API",
			"version": "1.0.0"
		},
		"paths": {}
	}`

	document, err := libopenapi.NewDocument([]byte(minimalSpec))
	if err != nil {
		t.Fatalf("Error creating minimal document: %v", err)
	}

	v3Model, err := document.BuildV3Model()
	if err != nil {
		t.Fatalf("Error building v3 model for minimal spec: %v", err)
	}

	model := NewModel(&v3Model.Model)
	model.width = 80
	model.height = 24

	view := model.View()
	if view == "" {
		t.Error("View should not be empty even for minimal document")
	}

	model.ensureCursorVisible()
	if model.cursor != 0 {
		t.Errorf("Cursor should remain 0 for empty document, got %d", model.cursor)
	}
}
