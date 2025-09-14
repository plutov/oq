package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getkin/kin-openapi/openapi3"
)

type viewMode int

const (
	viewEndpoints viewMode = iota
	viewComponents
	viewWebhooks
)

const keySequenceThreshold = 500 * time.Millisecond

type webhook struct {
	name   string
	method string
	op     *openapi3.Operation
	folded bool
}

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
	doc          *openapi3.T
	endpoints    []endpoint
	components   []component
	webhooks     []webhook
	cursor       int
	mode         viewMode
	width        int
	height       int
	showHelp     bool
	lastKey      string
	lastKeyAt    time.Time
	scrollOffset int
}

func (m *Model) ensureCursorVisible() {
	// header (~6 lines) + footer (~4 lines)
	contentHeight := max(1, m.height-10)

	// For simplicity, treat each item as taking 1 line for visibility calculations
	// The actual rendering will handle multi-line items
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+contentHeight {
		m.scrollOffset = m.cursor - contentHeight + 1
	}

	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func NewModel(doc *openapi3.T) Model {
	endpoints := extractEndpoints(doc)
	components := extractComponents(doc)
	webhooks := extractWebhooks(doc)

	return Model{
		doc:          doc,
		endpoints:    endpoints,
		components:   components,
		webhooks:     webhooks,
		cursor:       0,
		mode:         viewEndpoints,
		width:        80,
		height:       24,
		showHelp:     false,
		scrollOffset: 0,
	}
}

func NewModelWithWebhooks(doc *openapi3.T, webhooks []webhook) Model {
	endpoints := extractEndpoints(doc)
	components := extractComponents(doc)

	return Model{
		doc:          doc,
		endpoints:    endpoints,
		components:   components,
		webhooks:     webhooks,
		cursor:       0,
		mode:         viewEndpoints,
		width:        80,
		height:       24,
		showHelp:     false,
		scrollOffset: 0,
	}
}

func (m *Model) hasWebhooks() bool {
	return len(m.webhooks) > 0
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

		case "tab", "L":
			if !m.showHelp {
				// Cycle forward through available views
				switch m.mode {
				case viewEndpoints:
					if m.hasWebhooks() {
						m.mode = viewWebhooks
					} else {
						m.mode = viewComponents
					}
				case viewWebhooks:
					m.mode = viewComponents
				case viewComponents:
					m.mode = viewEndpoints
				}
				m.cursor = 0
				m.scrollOffset = 0
			}

		case "shift+tab", "H":
			if !m.showHelp {
				// Cycle backwards through available views
				switch m.mode {
				case viewEndpoints:
					m.mode = viewComponents
				case viewWebhooks:
					m.mode = viewEndpoints
				case viewComponents:
					if m.hasWebhooks() {
						m.mode = viewWebhooks
					} else {
						m.mode = viewEndpoints
					}
				}
				m.cursor = 0
				m.scrollOffset = 0
			}

		case "up", "k":
			if !m.showHelp && m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}

		case "down", "j":
			if !m.showHelp {
				var maxItems int
				switch m.mode {
				case viewEndpoints:
					maxItems = len(m.endpoints) - 1
				case viewComponents:
					maxItems = len(m.components) - 1
				case viewWebhooks:
					maxItems = len(m.webhooks) - 1
				}

				if m.cursor < maxItems {
					m.cursor++
					m.ensureCursorVisible()
				}
			}

		case "G":
			if !m.showHelp {
				var maxItems int
				switch m.mode {
				case viewEndpoints:
					maxItems = len(m.endpoints) - 1
				case viewComponents:
					maxItems = len(m.components) - 1
				case viewWebhooks:
					maxItems = len(m.webhooks) - 1
				}

				if maxItems >= 0 {
					m.cursor = maxItems
					m.ensureCursorVisible()
				}
			}

		case "g":
			now := time.Now()
			if m.lastKey == "g" && now.Sub(m.lastKeyAt) < keySequenceThreshold {
				if !m.showHelp {
					m.cursor = 0
					m.ensureCursorVisible()
				}

				// reset, so "ggg" wouldn't be triggered
				m.lastKey = ""
				m.lastKeyAt = time.Time{}

			} else {
				m.lastKey = "g"
				m.lastKeyAt = now
			}

		case "enter", " ":
			if !m.showHelp {
				if m.mode == viewEndpoints && m.cursor < len(m.endpoints) {
					m.endpoints[m.cursor].folded = !m.endpoints[m.cursor].folded
				} else if m.mode == viewComponents && m.cursor < len(m.components) {
					m.components[m.cursor].folded = !m.components[m.cursor].folded
				} else if m.mode == viewWebhooks && m.cursor < len(m.webhooks) {
					m.webhooks[m.cursor].folded = !m.webhooks[m.cursor].folded
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
	switch m.mode {
	case viewEndpoints:
		content = m.renderEndpoints()
	case viewComponents:
		content = m.renderComponents()
	case viewWebhooks:
		content = m.renderWebhooks()
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
