package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
)

func main() {
	var content []byte
	var err error

	var specFile string

	if len(os.Args) > 1 {
		specFile = os.Args[1]
		content, err = os.ReadFile(specFile)
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

	config := datamodel.NewDocumentConfiguration()
	if specFile != "" {
		absPath, _ := filepath.Abs(specFile)
		config.BasePath = filepath.Dir(absPath)
		config.SpecFilePath = absPath
		config.AllowFileReferences = true
	}

	document, err := libopenapi.NewDocumentWithConfiguration(content, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating document: %v\n", err)
		os.Exit(1)
	}

	v3Model, err := document.BuildV3Model()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building V3 model: %v\n", err)
		os.Exit(1)
	}

	m := NewModel(&v3Model.Model)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
