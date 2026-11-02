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
	queries      [viewCount]string
	searching    bool
	searchCursor int
	searchOffset int
	searchQuery  string
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

func (m Model) totalLines() int {
	lines := 0
	for i := range m.activeItems() {
		lines += m.getItemHeight(i)
	}
	return lines
}

func (m *Model) clampScroll() {
	m.scrollOffset = max(0, min(m.scrollOffset, m.totalLines()-m.contentHeight()))
}

func (m *Model) ensureCursorVisible() {
	if len(m.activeItems()) == 0 {
		m.cursor, m.scrollOffset = 0, 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.activeItems())-1))
	top := 0
	for i := 0; i < m.cursor; i++ {
		top += m.getItemHeight(i)
	}
	height := m.contentHeight()
	bottom := top + m.getItemHeight(m.cursor)
	if top < m.scrollOffset || m.getItemHeight(m.cursor) > height {
		m.scrollOffset = top
	} else if bottom > m.scrollOffset+height {
		m.scrollOffset = bottom - height
	}
	m.clampScroll()
}

func (m *Model) scrollLines(delta int) {
	m.scrollOffset += delta
	m.clampScroll()
}

func (m Model) matches(index int) bool {
	entry := m.activeItems()[index]
	text := entry.name + "\n" + entry.kind + "\n" + entry.description + "\n" + entry.details
	return strings.Contains(strings.ToLower(text), strings.ToLower(m.queries[m.mode]))
}

func (m *Model) findMatch(start, direction int) {
	count := len(m.activeItems())
	if m.queries[m.mode] == "" || count == 0 {
		return
	}
	for step := 1; step <= count; step++ {
		index := ((start+step*direction)%count + count) % count
		if m.matches(index) {
			m.cursor = index
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) updateSearch(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
		m.cursor, m.scrollOffset = m.searchCursor, m.searchOffset
		m.queries[m.mode] = m.searchQuery
		m.clampScroll()
		return
	case tea.KeyEnter:
		m.searching = false
		return
	case tea.KeyBackspace, tea.KeyCtrlH, tea.KeyDelete:
		runes := []rune(m.queries[m.mode])
		if len(runes) > 0 {
			m.queries[m.mode] = string(runes[:len(runes)-1])
		}
	case tea.KeyRunes:
		m.queries[m.mode] += string(msg.Runes)
	case tea.KeySpace:
		m.queries[m.mode] += " "
	default:
		return
	}
	m.cursor, m.scrollOffset = m.searchCursor, m.searchOffset
	m.clampScroll()
	m.findMatch(m.searchCursor, 1)
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
		top := 0
		for i := 0; i < m.cursor; i++ {
			top += m.getItemHeight(i)
		}
		selectedVisible := len(m.activeItems()) > 0 && top >= m.scrollOffset && top < m.scrollOffset+m.contentHeight()
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		m.clampScroll()
		if selectedVisible && (top < m.scrollOffset || top >= m.scrollOffset+m.contentHeight()) {
			m.ensureCursorVisible()
		}
	case tea.MouseMsg:
		if !m.showHelp && !m.searching {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.scrollLines(-3)
			case tea.MouseButtonWheelDown:
				m.scrollLines(3)
			}
		}
	case tea.KeyMsg:
		key := msg.String()
		previousG := m.lastKeyAt
		m.lastKeyAt = time.Time{}

		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.searching {
			m.updateSearch(msg)
			return m, nil
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
		case "/":
			m.searching = true
			m.searchCursor, m.searchOffset = m.cursor, m.scrollOffset
			m.searchQuery, m.queries[m.mode] = m.queries[m.mode], ""
		case "n":
			m.findMatch(m.cursor, 1)
		case "N":
			m.findMatch(m.cursor, -1)
		case "pgup":
			m.scrollLines(-m.contentHeight())
		case "pgdown":
			m.scrollLines(m.contentHeight())
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
