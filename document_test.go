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
