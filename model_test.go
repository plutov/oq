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
			if m.cursor < 0 || m.cursor >= len(m.activeItems()) || m.scrollOffset < 0 || m.scrollOffset > m.cursor {
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
