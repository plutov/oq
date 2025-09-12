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
	showHelp   bool
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
		showHelp:   false,
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
			if m.showHelp {
				m.showHelp = false
			} else {
				return m, tea.Quit
			}

		case "?":
			m.showHelp = !m.showHelp

		case "esc":
			if m.showHelp {
				m.showHelp = false
			}

		case "tab":
			if !m.showHelp {
				if m.mode == viewEndpoints {
					m.mode = viewComponents
				} else {
					m.mode = viewEndpoints
				}
				m.cursor = 0
			}

		case "up", "k":
			if !m.showHelp && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.showHelp {
				maxItems := 0
				if m.mode == viewEndpoints {
					maxItems = len(m.endpoints) - 1
				} else {
					maxItems = len(m.components) - 1
				}
				if m.cursor < maxItems {
					m.cursor++
				}
			}

		case "enter", " ":
			if !m.showHelp {
				if m.mode == viewEndpoints && m.cursor < len(m.endpoints) {
					m.endpoints[m.cursor].folded = !m.endpoints[m.cursor].folded
				} else if m.mode == viewComponents && m.cursor < len(m.components) {
					m.components[m.cursor].folded = !m.components[m.cursor].folded
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	header := m.renderHeader()
	s.WriteString(header)

	var content string
	if m.mode == viewEndpoints {
		content = m.renderEndpoints()
	} else {
		content = m.renderComponents()
	}
	s.WriteString(content)

	footer := m.renderFooter()

	headerLines := strings.Count(header, "\n")
	contentLines := strings.Count(content, "\n")
	footerLines := strings.Count(footer, "\n")

	usedLines := headerLines + contentLines + footerLines
	remainingLines := m.height - usedLines - 1

	if remainingLines > 0 {
		s.WriteString(strings.Repeat("\n", remainingLines))
	}

	s.WriteString(footer)

	baseView := s.String()

	if m.showHelp {
		return m.renderHelpModal()
	}

	return baseView
}
