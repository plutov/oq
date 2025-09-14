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

// Layout constants (shared with view.go)
const (
	headerApproxLines = 2 // Single line header + one empty line
	footerApproxLines = 4
	layoutBuffer      = 2 // Extra buffer to ensure header visibility
)

// calculateContentHeight returns the available height for content given the total viewport height
func calculateContentHeight(totalHeight int) int {
	return max(1, totalHeight-headerApproxLines-footerApproxLines-layoutBuffer)
}

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

func (m *Model) getItemHeight(index int) int {
	switch m.mode {
	case viewEndpoints:
		if index >= len(m.endpoints) {
			return 1
		}
		ep := m.endpoints[index]
		if ep.folded {
			return 1 // Just the main line when folded
		}
		// When unfolded, count main line + detail lines
		details := formatEndpointDetails(ep)
		return 1 + strings.Count(details, "\n") + 1 // +1 for main line, +1 for the detail section
	case viewComponents:
		if index >= len(m.components) {
			return 1
		}
		comp := m.components[index]
		if comp.folded {
			return 1 // Just the main line when folded
		}
		// When unfolded, count main line + detail lines
		return 1 + strings.Count(comp.details, "\n") + 1 // +1 for main line, +1 for the detail section
	case viewWebhooks:
		if index >= len(m.webhooks) {
			return 1
		}
		hook := m.webhooks[index]
		if hook.folded {
			return 1 // Just the main line when folded
		}
		// When unfolded, count main line + detail lines
		details := formatWebhookDetails(hook)
		return 1 + strings.Count(details, "\n") + 1 // +1 for main line, +1 for the detail section
	}
	return 1
}

func (m *Model) ensureCursorVisible() {
	// Calculate available content height using shared function
	contentHeight := calculateContentHeight(m.height)

	// Special case: if cursor is at 0, ensure we scroll to the very top
	if m.cursor == 0 {
		m.scrollOffset = 0
		return
	}

	// Calculate the actual rendered height of items to properly handle viewport
	var items []interface{}
	switch m.mode {
	case viewEndpoints:
		for i := range m.endpoints {
			items = append(items, m.endpoints[i])
		}
	case viewComponents:
		for i := range m.components {
			items = append(items, m.components[i])
		}
	case viewWebhooks:
		for i := range m.webhooks {
			items = append(items, m.webhooks[i])
		}
	}

	if len(items) == 0 {
		return
	}

	// Calculate lines used by items from scrollOffset to cursor
	linesUsed := 0

	// If cursor is above current scroll position, scroll up to show it
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
		return
	}

	// Calculate how many lines are used from scrollOffset to cursor (inclusive)
	for i := m.scrollOffset; i <= m.cursor && i < len(items); i++ {
		linesUsed += m.getItemHeight(i)
	}

	// Account for scroll indicators
	if m.scrollOffset > 0 {
		linesUsed++ // "More items above" indicator
	}

	// If the cursor item extends beyond available content height, scroll down
	if linesUsed > contentHeight {
		// Find the minimum scroll offset that keeps cursor visible
		for newScrollOffset := m.scrollOffset + 1; newScrollOffset <= m.cursor; newScrollOffset++ {
			testLinesUsed := 0

			// Account for "More items above" indicator
			if newScrollOffset > 0 {
				testLinesUsed++
			}

			// Calculate lines from new scroll offset to cursor
			for i := newScrollOffset; i <= m.cursor && i < len(items); i++ {
				testLinesUsed += m.getItemHeight(i)
			}

			if testLinesUsed <= contentHeight {
				m.scrollOffset = newScrollOffset
				break
			}
		}
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

// truncateContent ensures content doesn't exceed the available lines
func (m Model) truncateContent(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}

	// Truncate to fit and add an indicator
	truncatedLines := lines[:maxLines-1]
	truncatedLines = append(truncatedLines, "⬇ Content truncated to fit viewport...")

	return strings.Join(truncatedLines, "\n")
}

func (m Model) View() string {
	var s strings.Builder

	header := m.renderHeader()
	footer := m.renderFooter()

	headerLines := strings.Count(header, "\n")
	footerLines := strings.Count(footer, "\n")

	// Calculate how many lines are available for content
	availableContentLines := m.height - headerLines - footerLines - 1
	if availableContentLines < 1 {
		availableContentLines = 1
	}

	// Render content
	var content string
	switch m.mode {
	case viewEndpoints:
		content = m.renderEndpoints()
	case viewComponents:
		content = m.renderComponents()
	case viewWebhooks:
		content = m.renderWebhooks()
	}

	// Truncate content if it's too long
	content = m.truncateContent(content, availableContentLines)

	s.WriteString(header)
	s.WriteString(content)

	contentLines := strings.Count(content, "\n")
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
