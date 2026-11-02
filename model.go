package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type viewMode int

const (
	viewEndpoints viewMode = iota
	viewComponents
	viewWebhooks
	viewCount
)

const keySequenceThreshold = 500 * time.Millisecond

type item struct {
	name        string
	kind        string
	description string
	details     string
	folded      bool
}

type Model struct {
	title        string
	items        [viewCount][]item
	cursor       int
	mode         viewMode
	width        int
	height       int
	showHelp     bool
	lastKeyAt    time.Time
	scrollOffset int
}

func NewModel(doc *v3.Document) Model {
	m := Model{width: 80, height: 24}
	if doc != nil {
		if doc.Info != nil {
			m.title = doc.Info.Title + " v" + doc.Info.Version
		}
		if doc.Paths != nil {
			m.items[viewEndpoints] = extractOperations(doc.Paths.PathItems)
		}
		m.items[viewWebhooks] = extractOperations(doc.Webhooks)
		m.items[viewComponents] = extractComponents(doc)
	}
	return m
}

func (m Model) activeItems() []item {
	return m.items[m.mode]
}

func (m Model) availableViews() []viewMode {
	views := []viewMode{viewEndpoints}
	if len(m.items[viewWebhooks]) > 0 {
		views = append(views, viewWebhooks)
	}
	return append(views, viewComponents)
}

func (m *Model) cycleView(direction int) {
	views := m.availableViews()
	for i, mode := range views {
		if mode == m.mode {
			m.mode = views[(i+direction+len(views))%len(views)]
			break
		}
	}
	m.cursor, m.scrollOffset = 0, 0
}

func (m Model) contentHeight() int {
	return max(1, m.height-4)
}

func (m Model) getItemHeight(index int) int {
	entry := m.activeItems()[index]
	if entry.folded || entry.details == "" {
		return 1
	}
	return 2 + strings.Count(strings.TrimRight(entry.details, "\n"), "\n")
}

func (m *Model) ensureCursorVisible() {
	if len(m.activeItems()) == 0 {
		m.cursor, m.scrollOffset = 0, 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.activeItems())-1))
	m.scrollOffset = max(0, min(m.scrollOffset, m.cursor))
	if m.cursor == 0 {
		m.scrollOffset = 0
		return
	}

	height := m.contentHeight()
	lines := min(m.getItemHeight(m.cursor), max(1, height-1))
	for i := m.scrollOffset; i < m.cursor; i++ {
		lines += m.getItemHeight(i)
	}
	for m.scrollOffset < m.cursor {
		indicatorHeight := 0
		if m.scrollOffset > 0 {
			indicatorHeight = 1
		}
		if lines+indicatorHeight <= height {
			break
		}
		lines -= m.getItemHeight(m.scrollOffset)
		m.scrollOffset++
	}
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	m.ensureCursorVisible()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		m.ensureCursorVisible()
	case tea.KeyMsg:
		key := msg.String()
		previousG := m.lastKeyAt
		m.lastKeyAt = time.Time{}

		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			if key == "q" || key == "esc" {
				m.showHelp = false
			}
			return m, nil
		}

		switch key {
		case "q":
			return m, tea.Quit
		case "tab", "L":
			m.cycleView(1)
		case "shift+tab", "H":
			m.cycleView(-1)
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "ctrl+d":
			m.moveCursor(max(1, m.contentHeight()/2))
		case "ctrl+u":
			m.moveCursor(-max(1, m.contentHeight()/2))
		case "G":
			m.cursor = len(m.activeItems()) - 1
			m.ensureCursorVisible()
		case "g":
			now := time.Now()
			if !previousG.IsZero() && now.Sub(previousG) < keySequenceThreshold {
				m.cursor = 0
				m.ensureCursorVisible()
			} else {
				m.lastKeyAt = now
			}
		case "enter", " ":
			entries := m.activeItems()
			if len(entries) > 0 {
				entries[m.cursor].folded = !entries[m.cursor].folded
				m.ensureCursorVisible()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.showHelp {
		return fitView(m.renderHelpModal(), m.width, m.height)
	}
	return fitView(m.renderHeader()+m.renderItems()+m.renderFooter(), m.width, m.height)
}
