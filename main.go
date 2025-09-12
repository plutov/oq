package main

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getkin/kin-openapi/openapi3"
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
	doc, err := loader.LoadFromData(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing OpenAPI: %v\n", err)
		os.Exit(1)
	}

	err = doc.Validate(loader.Context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI validation error: %v\n", err)
		os.Exit(1)
	}

	m := NewModel(doc)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
