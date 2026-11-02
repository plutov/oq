package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func loadDocument(args []string, stdin io.Reader) (*v3.Document, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("usage: oq [openapi-file]")
	}
	config := datamodel.NewDocumentConfiguration()
	var content []byte
	var err error
	if len(args) == 0 {
		content, err = io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
	} else {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return nil, fmt.Errorf("resolve spec path: %w", err)
		}
		content, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read spec: %w", err)
		}
		config.BasePath = filepath.Dir(path)
		config.SpecFilePath = path
		config.AllowFileReferences = true
	}
	document, err := libopenapi.NewDocumentWithConfiguration(content, config)
	if err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("build OpenAPI model: %w", err)
	}
	return &model.Model, nil
}
