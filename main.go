package main

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pb33f/libopenapi"
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

	document, err := libopenapi.NewDocument(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating document: %v\n", err)
		os.Exit(1)
	}

	v3Model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Error building v3 model: %v\n", errs[0])
		os.Exit(1)
	}

	m := NewModel(&v3Model.Model)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
