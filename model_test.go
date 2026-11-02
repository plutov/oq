package main

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func press(m Model, key string) (Model, tea.Cmd) {
	keys := map[string]tea.KeyType{
		"ctrl+c": tea.KeyCtrlC, "ctrl+d": tea.KeyCtrlD, "ctrl+u": tea.KeyCtrlU,
		"tab": tea.KeyTab, "shift+tab": tea.KeyShiftTab, "enter": tea.KeyEnter,
		"esc": tea.KeyEsc, "up": tea.KeyUp, "down": tea.KeyDown,
		"backspace": tea.KeyBackspace, "pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown,
		" ": tea.KeySpace,
	}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if kind, ok := keys[key]; ok {
		msg = tea.KeyMsg{Type: kind}
	}
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

func listModel(count int) Model {
	m := NewModel(nil)
	for _, mode := range []viewMode{viewEndpoints, viewComponents, viewWebhooks} {
		for i := range count {
			m.items[mode] = append(m.items[mode], item{
				name: fmt.Sprintf("/item-%03d", i), kind: "GET", details: "Summary: test\n", folded: true,
			})
		}
	}
	return m
}

func assertViewport(t *testing.T, m Model) {
	t.Helper()
	view := m.View()
	if view == "" {
		t.Fatal("empty viewport")
	}
	lines := strings.Split(view, "\n")
	if len(lines) > max(1, m.height) {
		t.Fatalf("view has %d lines, height %d", len(lines), m.height)
	}
	for i, line := range lines {
		if lipgloss.Width(line) > max(1, m.width) {
			t.Fatalf("line %d has width %d, viewport %d: %q", i, lipgloss.Width(line), m.width, line)
		}
	}
}

func TestEmptyListNavigation(t *testing.T) {
	m := NewModel(parseSpec(t, minimalSpec))
	for _, key := range []string{"ctrl+d", "enter", " ", "j", "G", "ctrl+u", "k", "g", "g", "tab", "ctrl+d", "enter", "shift+tab"} {
		m, _ = press(m, key)
		if m.cursor != 0 || m.scrollOffset != 0 {
			t.Fatalf("after %q: cursor=%d offset=%d", key, m.cursor, m.scrollOffset)
		}
		assertViewport(t, m)
	}
}

func TestViewCycling(t *testing.T) {
	for _, webhooks := range []bool{false, true} {
		m := listModel(3)
		want := []viewMode{viewEndpoints, viewComponents}
		if webhooks {
			want = []viewMode{viewEndpoints, viewWebhooks, viewComponents}
		} else {
			m.items[viewWebhooks] = nil
		}
		if !reflect.DeepEqual(m.availableViews(), want) {
			t.Fatalf("views = %v, want %v", m.availableViews(), want)
		}
		for i := range len(want) * 2 {
			m.cursor, m.scrollOffset = 2, 1
			m, _ = press(m, "tab")
			if m.mode != want[(i+1)%len(want)] || m.cursor != 0 || m.scrollOffset != 0 {
				t.Fatalf("forward cycle %d: %+v", i, m)
			}
		}
		for i := range len(want) {
			m, _ = press(m, "shift+tab")
			if m.mode != want[(len(want)-1-i)%len(want)] {
				t.Fatalf("backward cycle %d: mode=%d", i, m.mode)
			}
		}
	}
}

func TestCursorMovement(t *testing.T) {
	for _, mode := range []viewMode{viewEndpoints, viewComponents, viewWebhooks} {
		m := listModel(100)
		m.mode = mode
		for _, test := range []struct {
			key  string
			want int
		}{
			{"k", 0}, {"j", 1}, {"down", 2}, {"up", 1},
			{"ctrl+d", 11}, {"ctrl+u", 1}, {"G", 99}, {"j", 99}, {"g", 99}, {"g", 0},
		} {
			m, _ = press(m, test.key)
			if m.cursor != test.want {
				t.Fatalf("mode %d, key %q: cursor=%d, want %d", mode, test.key, m.cursor, test.want)
			}
			if !strings.Contains(m.View(), m.activeItems()[m.cursor].name) {
				t.Fatalf("selected item %s is not visible", m.activeItems()[m.cursor].name)
			}
		}
		m, _ = press(m, "enter")
		if m.activeItems()[0].folded || !strings.Contains(m.View(), "Summary: test") {
			t.Fatal("enter did not expand details")
		}
		m, _ = press(m, " ")
		if !m.activeItems()[0].folded || strings.Contains(m.View(), "Summary: test") {
			t.Fatal("space did not fold details")
		}
	}
}

func TestGGRequiresConsecutiveKeys(t *testing.T) {
	m := listModel(100)
	m.cursor = 50
	for _, key := range []string{"g", "j", "g"} {
		m, _ = press(m, key)
	}
	if m.cursor != 51 {
		t.Fatal("non-consecutive g keys reset cursor")
	}
	m.lastKeyAt = time.Now().Add(-keySequenceThreshold - time.Second)
	m, _ = press(m, "g")
	if m.cursor != 51 {
		t.Fatal("expired gg reset cursor")
	}
	m, _ = press(m, "g")
	if m.cursor != 0 {
		t.Fatal("consecutive gg did not reset cursor")
	}
}

func TestHelpAndQuit(t *testing.T) {
	m := listModel(10)
	m, _ = press(m, "?")
	if !m.showHelp || !strings.Contains(m.View(), "Help") {
		t.Fatal("help did not open")
	}
	for _, key := range []string{"j", "tab", "enter", "G", "ctrl+d"} {
		m, _ = press(m, key)
		if m.cursor != 0 || m.mode != viewEndpoints || !m.activeItems()[0].folded {
			t.Fatalf("%q changed model while help is open", key)
		}
	}
	for _, key := range []string{"q", "esc", "?"} {
		m.showHelp = true
		var cmd tea.Cmd
		m, cmd = press(m, key)
		if m.showHelp || cmd != nil {
			t.Fatalf("%q did not close help without quitting", key)
		}
	}
	for _, key := range []string{"q", "ctrl+c"} {
		m.showHelp = key == "ctrl+c"
		_, cmd := press(m, key)
		if cmd == nil {
			t.Fatalf("%q did not quit", key)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%q did not return QuitMsg", key)
		}
	}
}

func TestExpandedItemScrollingAndResize(t *testing.T) {
	m := listModel(50)
	for i := range m.activeItems() {
		m.activeItems()[i].details = strings.Repeat("A detail line\n", 40)
		m.activeItems()[i].folded = false
	}
	for _, size := range [][2]int{{80, 24}, {40, 10}, {20, 5}, {10, 1}, {0, 0}} {
		updated, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		m = updated.(Model)
		for _, key := range []string{"j", "G", "k", "ctrl+u", "ctrl+d", "g", "g"} {
			m, _ = press(m, key)
			assertViewport(t, m)
			if m.height >= 5 && m.width >= 20 && !strings.Contains(m.View(), m.activeItems()[m.cursor].name) {
				t.Fatalf("selected item hidden after %q at size %v, offset=%d", key, size, m.scrollOffset)
			}
		}
	}
}

func TestResizeKeepsVisibleSelection(t *testing.T) {
	m := listModel(100)
	m, _ = press(m, "G")
	if m.cursor != 99 || m.scrollOffset != 80 {
		t.Fatalf("initial cursor=%d offset=%d", m.cursor, m.scrollOffset)
	}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m = updated.(Model)
	if m.cursor != 99 || m.scrollOffset != 94 || !strings.Contains(m.renderItems(), "/item-099") {
		t.Fatalf("selection hidden after shrink: cursor=%d offset=%d", m.cursor, m.scrollOffset)
	}
	assertViewport(t, m)
}

func TestResizePreservesManualScroll(t *testing.T) {
	t.Run("selection offscreen", func(t *testing.T) {
		m := listModel(100)
		m, _ = press(m, "pgdown")
		for _, height := range []int{10, 24} {
			updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: height})
			m = updated.(Model)
			if m.cursor != 0 || m.scrollOffset != 20 {
				t.Fatalf("height %d: cursor=%d offset=%d", height, m.cursor, m.scrollOffset)
			}
			assertViewport(t, m)
		}
	})
	t.Run("inside selected details", func(t *testing.T) {
		m := listModel(1)
		m.activeItems()[0].details = strings.Repeat("detail\n", 60)
		m, _ = press(m, "enter")
		m, _ = press(m, "pgdown")
		for _, height := range []int{10, 24, 50, 100} {
			updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: height})
			m = updated.(Model)
			want := min(20, max(0, m.totalLines()-m.contentHeight()))
			if m.cursor != 0 || m.scrollOffset != want {
				t.Fatalf("height %d: cursor=%d offset=%d, want offset=%d", height, m.cursor, m.scrollOffset, want)
			}
			if !strings.Contains(m.renderItems(), "detail") {
				t.Fatalf("height %d: details not visible", height)
			}
			assertViewport(t, m)
		}
	})
}

func TestViewportDimensions(t *testing.T) {
	m := NewModel(parseSpec(t, operationSpec))
	for _, size := range [][2]int{{120, 40}, {80, 24}, {20, 10}, {1, 1}, {0, 0}} {
		m.width, m.height = size[0], size[1]
		for _, mode := range m.availableViews() {
			m.mode = mode
			for i := range m.activeItems() {
				m.activeItems()[i].folded = false
			}
			assertViewport(t, m)
		}
		m.showHelp = true
		assertViewport(t, m)
		m.showHelp = false
	}
}

func TestLongAndMultilineLabels(t *testing.T) {
	m := listModel(1)
	m.width = 40
	m.activeItems()[0].name = strings.Repeat("猫", 80) + "\nanother line"
	m.activeItems()[0].description = "first\nsecond"
	m.activeItems()[0].kind = "PROPFIND"
	m.title = strings.Repeat("猫", 100) + "\nmultiline title"
	for _, width := range []int{0, 1, 10, 40, 80} {
		m.width = width
		assertViewport(t, m)
		if len(m.renderItem(m.activeItems()[0], true)) != m.getItemHeight(0) {
			t.Error("custom method changed item height")
		}
		if strings.Count(m.renderItem(m.activeItems()[0], true)[0], "\n") != 0 {
			t.Error("custom method wrapped to multiple lines")
		}
	}
}

func FuzzModelNavigation(f *testing.F) {
	f.Add([]byte{4, 7, 8, 9, 10, 11, 12, 13})
	f.Add([]byte{0, 0, 0, 2, 3, 5, 6, 8})
	f.Fuzz(func(t *testing.T, input []byte) {
		m := listModel(12)
		for mode := range m.items {
			for i := range m.items[mode] {
				m.items[mode][i].details = strings.Repeat("Details\n", i*3)
			}
		}
		keys := []string{"j", "k", "ctrl+d", "ctrl+u", "enter", " ", "tab", "shift+tab", "g", "G", "?", "esc", "q", "ctrl+c"}
		for _, value := range input[:min(len(input), 128)] {
			if value%5 == 0 {
				updated, _ := m.Update(tea.WindowSizeMsg{Width: int(value) % 100, Height: int(value) % 40})
				m = updated.(Model)
			} else {
				m, _ = press(m, keys[int(value)%len(keys)])
			}
			if m.cursor < 0 || m.cursor >= len(m.activeItems()) || m.scrollOffset < 0 || m.scrollOffset > max(0, m.totalLines()-m.contentHeight()) {
				t.Fatalf("invalid cursor=%d offset=%d", m.cursor, m.scrollOffset)
			}
			assertViewport(t, m)
			if !m.showHelp && m.width >= 40 && m.height >= 5 && !strings.Contains(m.View(), m.activeItems()[m.cursor].name) {
				t.Fatalf("selected item %s hidden at %dx%d", m.activeItems()[m.cursor].name, m.width, m.height)
			}
		}
	})
}

func TestProgramLifecycle(t *testing.T) {
	program := tea.NewProgram(NewModel(parseSpec(t, minimalSpec)),
		tea.WithInput(strings.NewReader("q")), tea.WithOutput(io.Discard),
		tea.WithoutRenderer(), tea.WithoutSignalHandler(),
	)
	if _, err := program.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestSearchAcrossViews(t *testing.T) {
	for _, mode := range []viewMode{viewEndpoints, viewComponents, viewWebhooks} {
		for _, field := range []string{"name", "kind", "description", "details"} {
			t.Run(fmt.Sprintf("%d/%s", mode, field), func(t *testing.T) {
				m := listModel(6)
				m.mode, m.height = mode, 7
				for _, i := range []int{2, 5} {
					entry := &m.items[mode][i]
					switch field {
					case "name":
						entry.name = "Needle"
					case "kind":
						entry.kind = "Needle"
					case "description":
						entry.description = "Needle"
					case "details":
						entry.details = "Needle"
					}
				}
				m, _ = press(m, "/")
				m, _ = press(m, "nEeDlE")
				if m.cursor != 2 || !m.searching || !strings.Contains(m.renderFooter(), "Match 1/2") {
					t.Fatalf("incremental match: cursor=%d footer=%s", m.cursor, m.renderFooter())
				}
				m, _ = press(m, "enter")
				if m.searching || !m.activeItems()[2].folded {
					t.Fatal("accepting search toggled details or left search open")
				}
				for _, step := range []struct {
					key  string
					want int
				}{{"n", 5}, {"n", 2}, {"N", 5}, {"N", 2}} {
					m, _ = press(m, step.key)
					if m.cursor != step.want || !strings.Contains(m.renderItems(), "Needle") && field != "details" {
						t.Fatalf("%s selected %d, want visible %d", step.key, m.cursor, step.want)
					}
					assertViewport(t, m)
				}
				m, _ = press(m, "tab")
				if m.queries[m.mode] != "" || m.cursor != 0 {
					t.Fatal("view switch leaked query or cursor")
				}
				m, _ = press(m, "shift+tab")
				m, _ = press(m, "n")
				if m.mode != mode || m.cursor != 2 || m.queries[mode] != "nEeDlE" {
					t.Fatal("view query was not preserved")
				}
			})
		}
	}
}

func TestSearchInputAndCancellation(t *testing.T) {
	for _, mode := range []viewMode{viewEndpoints, viewComponents, viewWebhooks} {
		m := listModel(30)
		m.mode = mode
		m.activeItems()[20].details = "q? 猫.*"
		m.activeItems()[25].details = "q? 猫.*"
		m.cursor = 10
		m.ensureCursorVisible()
		cursor, offset := m.cursor, m.scrollOffset
		m.queries[mode] = "old"
		m, _ = press(m, "/")
		for _, key := range []string{"q", "?", " ", "猫.*"} {
			var cmd tea.Cmd
			m, cmd = press(m, key)
			if cmd != nil || m.showHelp || !m.searching {
				t.Fatalf("search input %q invoked a command", key)
			}
		}
		if m.queries[mode] != "q? 猫.*" || m.cursor != 20 {
			t.Fatalf("literal search query=%q cursor=%d", m.queries[mode], m.cursor)
		}
		m, _ = press(m, "backspace")
		if m.queries[mode] != "q? 猫." {
			t.Fatal("backspace did not remove one rune")
		}
		m, _ = press(m, "backspace")
		m, _ = press(m, "backspace")
		if m.queries[mode] != "q? " {
			t.Fatal("backspace corrupted Unicode")
		}
		for _, key := range []string{"tab", "up", "pgdown"} {
			m, _ = press(m, key)
			if m.mode != mode || m.cursor != 20 {
				t.Fatal("navigation escaped search input")
			}
		}
		m, _ = press(m, "esc")
		if m.searching || m.cursor != cursor || m.scrollOffset != offset || m.queries[mode] != "old" {
			t.Fatalf("cancellation did not restore state: %+v", m)
		}
		m, _ = press(m, "/")
		_, cmd := press(m, "ctrl+c")
		if cmd == nil {
			t.Fatal("Ctrl+C did not quit during search")
		}
	}
}

func TestSearchBackspaceEncodings(t *testing.T) {
	for _, key := range []struct {
		name string
		kind tea.KeyType
	}{
		{"BS", tea.KeyCtrlH},
		{"DEL", tea.KeyBackspace},
		{"Delete", tea.KeyDelete},
	} {
		t.Run(key.name, func(t *testing.T) {
			m := listModel(4)
			for i, name := range []string{"/origin", "a", "a猫", "a猫🙂"} {
				m.activeItems()[i].name = name
				m.activeItems()[i].details = ""
			}
			m, _ = press(m, "/")
			m, _ = press(m, "a猫🙂")
			if m.cursor != 3 {
				t.Fatalf("initial search selected %d, want 3", m.cursor)
			}
			for _, step := range []struct {
				query  string
				cursor int
			}{{"a猫", 2}, {"a", 1}, {"", 0}, {"", 0}} {
				updated, cmd := m.Update(tea.KeyMsg{Type: key.kind})
				m = updated.(Model)
				if m.queries[m.mode] != step.query || m.cursor != step.cursor || !m.searching || cmd != nil {
					t.Fatalf("query=%q cursor=%d searching=%v, want query=%q cursor=%d", m.queries[m.mode], m.cursor, m.searching, step.query, step.cursor)
				}
			}
		})
	}
}

func TestSearchEmptyNoMatchAndResize(t *testing.T) {
	for _, count := range []int{0, 8} {
		m := listModel(count)
		m, _ = press(m, "/")
		m, _ = press(m, "backspace")
		m, _ = press(m, "enter")
		for _, key := range []string{"n", "N", "/", "no such entity"} {
			m, _ = press(m, key)
		}
		if !strings.Contains(m.renderFooter(), "No matches") || m.cursor != 0 {
			t.Fatal("no-match search moved cursor or lacked status")
		}
		for _, size := range [][2]int{{1, 1}, {0, 0}, {40, 10}} {
			updated, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			m = updated.(Model)
			assertViewport(t, m)
		}
		for range len([]rune("no such entity")) {
			m, _ = press(m, "backspace")
		}
		if m.queries[m.mode] != "" || m.cursor != 0 {
			t.Fatal("empty incremental search changed selection")
		}
		m, _ = press(m, "esc")
		m, _ = press(m, "tab")
		assertViewport(t, m)
	}
}

func TestLongItemLineScrolling(t *testing.T) {
	for _, mode := range []viewMode{viewEndpoints, viewComponents, viewWebhooks} {
		m := listModel(1)
		m.mode, m.height = mode, 10
		m.activeItems()[0].details = strings.Repeat("detail\n", 50) + "FINAL DETAIL\n"
		m, _ = press(m, "enter")
		if strings.Contains(m.View(), "FINAL DETAIL") {
			t.Fatal("fixture must overflow")
		}
		for range 20 {
			m, _ = press(m, "pgdown")
			assertViewport(t, m)
		}
		if !strings.Contains(m.View(), "FINAL DETAIL") || m.scrollOffset != m.totalLines()-m.contentHeight() {
			t.Fatalf("last detail unreachable: offset=%d", m.scrollOffset)
		}
		bottom := m.scrollOffset
		updated, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
		m = updated.(Model)
		if m.scrollOffset != bottom-3 {
			t.Fatal("wheel up did not scroll lines")
		}
		updated, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
		m = updated.(Model)
		if m.scrollOffset != bottom {
			t.Fatal("wheel down did not scroll lines")
		}
		for range 20 {
			m, _ = press(m, "pgup")
		}
		if m.scrollOffset != 0 || !strings.Contains(m.View(), m.activeItems()[0].name) {
			t.Fatal("page up did not reach header")
		}
		m, _ = press(m, "pgdown")
		m, _ = press(m, "enter")
		if m.scrollOffset != 0 || !m.activeItems()[0].folded {
			t.Fatal("folding did not clamp scroll")
		}
	}
}

func TestProportionalScrollbar(t *testing.T) {
	m := listModel(100)
	m.width, m.height = 30, 14
	for _, offset := range []int{0, 30, 90} {
		m.scrollOffset = offset
		top, size := m.scrollbar()
		if size != 1 || top != offset*9/90 {
			t.Fatalf("offset %d: thumb=%d size=%d", offset, top, size)
		}
		lines := strings.Split(strings.TrimSuffix(m.renderItems(), "\n"), "\n")
		for i, line := range lines {
			if lipgloss.Width(line) != m.width {
				t.Fatalf("scrollbar not at right edge: %q", line)
			}
			if strings.HasSuffix(line, "█") != (i >= top && i < top+size) {
				t.Fatalf("incorrect thumb on line %d: %q", i, line)
			}
		}
		assertViewport(t, m)
	}
	m = listModel(20)
	m.height = 14
	if top, size := m.scrollbar(); top != 0 || size != 5 {
		t.Fatalf("half-visible list: thumb=%d size=%d", top, size)
	}
	m.scrollOffset = 10
	for _, size := range [][2]int{{1, 1}, {2, 5}, {40, 8}, {40, 40}, {0, 0}} {
		updated, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		m = updated.(Model)
		top, thumb := m.scrollbar()
		if top < 0 || top+thumb > m.contentHeight() || m.scrollOffset > max(0, m.totalLines()-m.contentHeight()) {
			t.Fatalf("resize %v: invalid scrollbar or offset", size)
		}
		assertViewport(t, m)
	}
	m = listModel(2)
	if _, size := m.scrollbar(); size != 0 || strings.ContainsAny(m.renderItems(), "█│") {
		t.Fatal("nonoverflowing list has a scrollbar")
	}
}
