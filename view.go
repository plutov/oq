package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	colorGreen       = "#10B981"
	colorBlue        = "#3B82F6"
	colorYellow      = "#F59E0B"
	colorRed         = "#EF4444"
	colorPurple      = "#8B5CF6"
	colorGray        = "#6B7280"
	colorThemePurple = "#7C3AED"
	colorBackground  = "#374151"
	colorDetailGray  = "#9CA3AF"
	colorFooterText  = "#000000"
	colorWhite       = "#FFFFFF"
)

var kindColors = map[string]lipgloss.Color{
	"GET": colorGreen, "POST": colorBlue, "PUT": colorYellow,
	"DELETE": colorRed, "PATCH": colorPurple,
	"Schema": colorGreen, "RequestBody": colorBlue, "Response": colorYellow,
	"Parameter": colorPurple, "Header": colorRed,
}

func singleLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func fitView(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	lines = lines[:min(len(lines), max(1, height))]
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, max(1, width), "")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderItem(entry item, selected bool) []string {
	style := lipgloss.NewStyle()
	color := kindColors[entry.kind]
	if color == "" {
		color = colorGray
	}
	label, labelWidth := singleLine(entry.kind), 7
	if m.mode == viewComponents {
		label, labelWidth = label+":", 16
	}
	labelStyle := lipgloss.NewStyle().Foreground(color).Bold(true).Width(labelWidth)
	if selected {
		style = style.Background(lipgloss.Color(colorBackground))
		labelStyle = labelStyle.Background(lipgloss.Color(colorBackground))
	}
	icon := "▶"
	if !entry.folded {
		icon = "▼"
	}
	name := singleLine(entry.name)
	if entry.description != "" {
		name += " - " + singleLine(entry.description)
	}
	line := style.Render(icon+" ") + labelStyle.Render(ansi.Truncate(label, labelWidth, "")) + style.Render(" "+name)
	line = ansi.Truncate(line, max(1, m.width), "")
	line += style.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(line))))
	lines := []string{line}
	if !entry.folded && entry.details != "" {
		detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorDetailGray))
		for _, detail := range strings.Split(strings.TrimRight(entry.details, "\n"), "\n") {
			lines = append(lines, ansi.Truncate(detailStyle.Render("  "+detail), max(1, m.width), ""))
		}
	}
	return lines
}

func (m Model) scrollbar() (int, int) {
	height, total := m.contentHeight(), m.totalLines()
	if total <= height {
		return 0, 0
	}
	size := max(1, height*height/total)
	offset := max(0, min(m.scrollOffset, total-height))
	return offset * (height - size) / (total - height), size
}

func (m Model) renderItems() string {
	height := m.contentHeight()
	lines := make([]string, 0, height)
	thumb, size := m.scrollbar()
	if size > 0 {
		m.width = max(0, m.width-1)
	}
	top := 0
	for i, entry := range m.activeItems() {
		itemHeight := m.getItemHeight(i)
		if top+itemHeight > m.scrollOffset {
			itemLines := m.renderItem(entry, i == m.cursor)
			start := max(0, m.scrollOffset-top)
			lines = append(lines, itemLines[start:min(len(itemLines), start+height-len(lines))]...)
		}
		top += itemHeight
		if len(lines) == height {
			break
		}
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	if size > 0 {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorGray))
		for i, line := range lines {
			line = ansi.Truncate(line, m.width, "")
			bar := "│"
			if i >= thumb && i < thumb+size {
				bar = "█"
			}
			lines[i] = line + strings.Repeat(" ", max(0, m.width-lipgloss.Width(line))) + style.Render(bar)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) renderHeader() string {
	buttonStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color(colorGray))
	activeStyle := buttonStyle.Background(lipgloss.Color(colorThemePurple)).
		Foreground(lipgloss.Color(colorWhite)).Bold(true)
	titles := [...]string{"Requests", "Components", "Webhooks"}
	var buttons []string
	for _, mode := range m.availableViews() {
		style := buttonStyle
		if mode == m.mode {
			style = activeStyle
		}
		buttons = append(buttons, style.Render(titles[mode]))
	}
	header := strings.Join(buttons, " │ ")
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorThemePurple)).
		Render("oq - OpenAPI Spec Viewer")
	if space := m.width - lipgloss.Width(header) - lipgloss.Width(title); space > 4 {
		header += strings.Repeat(" ", space) + title
	}
	return header + "\n\n"
}

func (m Model) renderFooter() string {
	width := max(1, m.width)
	if query := m.queries[m.mode]; m.searching || query != "" {
		status := ""
		if query != "" {
			count, current := 0, 0
			for i := range m.activeItems() {
				if m.matches(i) {
					count++
					if i == m.cursor {
						current = count
					}
				}
			}
			switch {
			case count == 0:
				status = "  No matches"
			case current > 0:
				status = fmt.Sprintf("  Match %d/%d", current, count)
			default:
				status = fmt.Sprintf("  %d matches", count)
			}
		}
		return "\n" + ansi.Truncate("/"+singleLine(query)+status, width, "")
	}
	schemaInfo := ansi.Truncate(singleLine(m.title), max(0, width-2), "")
	help := "Press '?' for help"
	if m.showHelp || lipgloss.Width(help)+lipgloss.Width(schemaInfo)+4 > width {
		help = ""
	}
	padding := max(0, width-2-lipgloss.Width(help)-lipgloss.Width(schemaInfo))
	style := lipgloss.NewStyle().Background(lipgloss.Color(colorGray)).
		Foreground(lipgloss.Color(colorFooterText))
	return "\n" + ansi.Truncate(style.Render(" "+help+strings.Repeat(" ", padding)+schemaInfo+" "), width, "")
}

func (m Model) renderHelpModal() string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorBlue)).Bold(true)
	rows := [][2]string{
		{"↑/k, ↓/j", "Previous/next item"},
		{"gg / G", "First/last item"},
		{"Ctrl-U/D", "Move half a screen of items"},
		{"PgUp/PgDn", "Scroll a screen of lines"},
		{"Mouse wheel", "Scroll lines"},
		{"Tab/L", "Cycle forward through views"},
		{"Shift+Tab/H", "Cycle backward through views"},
		{"Enter/Space", "Toggle details"},
		{"/", "Incremental search (literal, ignore case)"},
		{"Enter / Esc", "Accept/cancel search"},
		{"n / N", "Next/previous match (wrap)"},
		{"?", "Toggle help"},
		{"Esc/q", "Close help"},
		{"q / Ctrl+C", "Quit (Ctrl+C also during search)"},
	}
	var lines []string
	for _, row := range rows {
		lines = append(lines, keyStyle.Width(12).Render(row[0])+" "+row[1])
	}
	modal := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorThemePurple)).Padding(1, 2).
		Render("Help\n\n" + strings.Join(lines, "\n"))
	return lipgloss.Place(max(1, m.width), max(1, m.height), lipgloss.Center, lipgloss.Center, modal)
}
