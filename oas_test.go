package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

// TestAllExampleFiles loads and tests all files from the examples folder
func TestAllExampleFiles(t *testing.T) {
	examplesDir := "examples"

	// Get all files in the examples directory
	files, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only test YAML and JSON files
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

// testExampleFile tests a single example file
func testExampleFile(t *testing.T, filepath string) {
	// Read the file
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filepath, err)
	}

	// Test document creation (similar to main.go)
	document, err := libopenapi.NewDocument(content)
	if err != nil {
		t.Fatalf("Error creating document from %s: %v", filepath, err)
	}

	// Test v3 model building
	v3Model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		t.Fatalf("Error building v3 model from %s: %v", filepath, errs[0])
	}

	if v3Model == nil {
		t.Fatalf("V3 model is nil for %s", filepath)
	}

	// Test model creation and rendering
	model := NewModel(&v3Model.Model)

	// Test that the model was created successfully
	if model.doc == nil {
		t.Fatalf("Model document is nil for %s", filepath)
	}

	// Test endpoint extraction
	endpoints := model.endpoints
	t.Logf("File %s: Found %d endpoints", filepath, len(endpoints))

	// Test that endpoints can be formatted without errors
	for i, ep := range endpoints {
		details := formatEndpointDetails(ep)
		if details == "" {
			t.Errorf("Empty endpoint details for endpoint %d (%s %s) in %s",
				i, ep.method, ep.path, filepath)
		}
	}

	// Test component extraction
	components := model.components
	t.Logf("File %s: Found %d components", filepath, len(components))

	// Test that components have proper details (allowing some to be empty as this may be valid)
	emptyDetailsCount := 0
	for _, comp := range components {
		if comp.details == "" {
			emptyDetailsCount++
		}
	}
	if emptyDetailsCount > 0 {
		t.Logf("File %s: %d components have empty details (this may be normal for some schema types)",
			filepath, emptyDetailsCount)
	}

	// Test webhook extraction
	webhooks := model.webhooks
	t.Logf("File %s: Found %d webhooks", filepath, len(webhooks))

	// Test that webhooks can be formatted without errors (allowing some to be empty)
	emptyWebhookCount := 0
	for _, hook := range webhooks {
		details := formatWebhookDetails(hook)
		if details == "" {
			emptyWebhookCount++
		}
	}
	if emptyWebhookCount > 0 {
		t.Logf("File %s: %d webhooks have empty details (this may be normal)",
			filepath, emptyWebhookCount)
	}

	// Test model rendering in all view modes
	testModelRendering(t, &model, filepath)

	t.Logf("Successfully processed %s: %d endpoints, %d components, %d webhooks",
		filepath, len(endpoints), len(components), len(webhooks))
}

// testModelRendering tests that the model can render in all view modes without errors
func testModelRendering(t *testing.T, model *Model, filepath string) {
	// Set reasonable dimensions
	model.width = 120
	model.height = 40

	// Test endpoints view
	model.mode = viewEndpoints
	endpointsView := model.View()
	if endpointsView == "" {
		t.Errorf("Empty endpoints view for %s", filepath)
	}

	// Test components view
	model.mode = viewComponents
	componentsView := model.View()
	if componentsView == "" {
		t.Errorf("Empty components view for %s", filepath)
	}

	// Test webhooks view (if webhooks exist)
	if len(model.webhooks) > 0 {
		model.mode = viewWebhooks
		webhooksView := model.View()
		if webhooksView == "" {
			t.Errorf("Empty webhooks view for %s", filepath)
		}
	}

	// Test help view
	model.showHelp = true
	helpView := model.View()
	if helpView == "" {
		t.Errorf("Empty help view for %s", filepath)
	}
	model.showHelp = false

	// Test folding/unfolding functionality
	if len(model.endpoints) > 0 {
		model.mode = viewEndpoints
		model.cursor = 0

		// Test folded state
		model.endpoints[0].folded = true
		foldedView := model.View()
		if foldedView == "" {
			t.Errorf("Empty folded endpoints view for %s", filepath)
		}

		// Test unfolded state
		model.endpoints[0].folded = false
		unfoldedView := model.View()
		if unfoldedView == "" {
			t.Errorf("Empty unfolded endpoints view for %s", filepath)
		}
	}

	if len(model.components) > 0 {
		model.mode = viewComponents
		model.cursor = 0

		// Test folded state
		model.components[0].folded = true
		foldedView := model.View()
		if foldedView == "" {
			t.Errorf("Empty folded components view for %s", filepath)
		}

		// Test unfolded state
		model.components[0].folded = false
		unfoldedView := model.View()
		if unfoldedView == "" {
			t.Errorf("Empty unfolded components view for %s", filepath)
		}
	}
}

// TestSpecificExampleFiles tests specific known files for additional validation
func TestSpecificExampleFiles(t *testing.T) {
	specificTests := []struct {
		filename      string
		minEndpoints  int
		minComponents int
	}{
		{"petstore-3.0.yaml", 1, 1},
		{"petstore-3.1.yaml", 1, 1},
		// Note: train-travel files have external references that cause issues
		// We test them in the main test but they may be skipped
	}

	for _, test := range specificTests {
		t.Run(test.filename, func(t *testing.T) {
			filepath := filepath.Join("examples", test.filename)

			// Check if file exists
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

			v3Model, errs := document.BuildV3Model()
			if len(errs) > 0 {
				t.Fatalf("Error building v3 model from %s: %v", filepath, errs[0])
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

			// Verify that document info is accessible
			if model.doc.Info == nil {
				t.Errorf("Document info is nil for %s", test.filename)
			}
		})
	}
}

// TestModelNavigation tests navigation functionality
func TestModelNavigation(t *testing.T) {
	// Use petstore as a test file since it should have multiple endpoints
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

	v3Model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		t.Fatalf("Error building v3 model: %v", errs[0])
	}

	model := NewModel(&v3Model.Model)
	model.width = 120
	model.height = 40

	// Test cursor navigation doesn't crash
	initialCursor := model.cursor

	// Test moving cursor down
	if len(model.endpoints) > 1 {
		model.cursor = 1
		model.ensureCursorVisible()
		if model.cursor != 1 {
			t.Errorf("Cursor should be 1, got %d", model.cursor)
		}
	}

	// Test moving cursor back to start
	model.cursor = 0
	model.ensureCursorVisible()
	if model.cursor != 0 {
		t.Errorf("Cursor should be 0, got %d", model.cursor)
	}

	// Test view mode switching
	originalMode := model.mode

	// Switch to components view
	model.mode = viewComponents
	model.cursor = 0
	componentsView := model.View()
	if componentsView == "" {
		t.Error("Components view should not be empty")
	}

	// Switch back
	model.mode = originalMode
	model.cursor = initialCursor
	endpointsView := model.View()
	if endpointsView == "" {
		t.Error("Endpoints view should not be empty")
	}
}

// TestEmptyDocument tests behavior with minimal valid documents
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

	v3Model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		t.Fatalf("Error building v3 model for minimal spec: %v", errs[0])
	}

	model := NewModel(&v3Model.Model)
	model.width = 80
	model.height = 24

	// Should not crash even with no endpoints/components
	view := model.View()
	if view == "" {
		t.Error("View should not be empty even for minimal document")
	}

	// Should handle navigation gracefully
	model.ensureCursorVisible()
	if model.cursor != 0 {
		t.Errorf("Cursor should remain 0 for empty document, got %d", model.cursor)
	}
}
