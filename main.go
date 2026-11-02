package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	doc, err := loadDocument(os.Args[1:], os.Stdin)
	if err == nil {
		_, err = tea.NewProgram(NewModel(doc), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "oq: %v\n", err)
		os.Exit(1)
	}
}
