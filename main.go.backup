package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func main() {
	var content []byte
	var err error

	if len(os.Args) > 1 {
		content, err = os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Parse webhooks from original content before conversion
	originalWebhooks := parseWebhooksFromRawContent(content)

	// Try to convert OpenAPI 3.1 to 3.0 if needed
	convertedContent, err := convertOpenAPI31To30(content)
	if err != nil {
		convertedContent = content
	}

	doc, err := loader.LoadFromData(convertedContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing OpenAPI: %v\n", err)
		os.Exit(1)
	}

	err = doc.Validate(loader.Context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI validation error: %v\n", err)
		os.Exit(1)
	}

	m := NewModelWithWebhooks(doc, originalWebhooks)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func parseWebhooksFromRawContent(content []byte) []webhook {
	var webhooks []webhook

	// Parse as YAML first, then JSON if that fails
	var spec map[string]interface{}
	err := yaml.Unmarshal(content, &spec)
	if err != nil {
		// Try JSON
		err = json.Unmarshal(content, &spec)
		if err != nil {
			return webhooks
		}
	}

	if version, ok := spec["openapi"].(string); !ok || !strings.HasPrefix(version, "3.1") {
		return webhooks
	}

	if webhookData, ok := spec["webhooks"].(map[string]interface{}); ok {
		webhooks = parseWebhooksFromData(webhookData)
	}

	return webhooks
}
