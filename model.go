package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getkin/kin-openapi/openapi3"
)

type viewMode int

const (
	viewEndpoints viewMode = iota
	viewComponents
)

type endpoint struct {
	path   string
	method string
	op     *openapi3.Operation
	folded bool
}

type component struct {
	name        string
	compType    string
	description string
	details     string
	folded      bool
}

type Model struct {
	doc        *openapi3.T
	endpoints  []endpoint
	components []component
	cursor     int
	mode       viewMode
	width      int
	height     int
}

func NewModel(doc *openapi3.T) Model {
	endpoints := extractEndpoints(doc)
	components := extractComponents(doc)

	return Model{
		doc:        doc,
		endpoints:  endpoints,
		components: components,
		cursor:     0,
		mode:       viewEndpoints,
		width:      80,
		height:     24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			if m.mode == viewEndpoints {
				m.mode = viewComponents
			} else {
				m.mode = viewEndpoints
			}
			m.cursor = 0

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxItems := 0
			if m.mode == viewEndpoints {
				maxItems = len(m.endpoints) - 1
			} else {
				maxItems = len(m.components) - 1
			}
			if m.cursor < maxItems {
				m.cursor++
			}

		case "enter", " ":
			if m.mode == viewEndpoints && m.cursor < len(m.endpoints) {
				m.endpoints[m.cursor].folded = !m.endpoints[m.cursor].folded
			} else if m.mode == viewComponents && m.cursor < len(m.components) {
				m.components[m.cursor].folded = !m.components[m.cursor].folded
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	s.WriteString(m.renderHeader())

	if m.mode == viewEndpoints {
		s.WriteString(m.renderEndpoints())
	} else {
		s.WriteString(m.renderComponents())
	}

	s.WriteString(m.renderFooter())

	return s.String()
}
