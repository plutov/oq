package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func (m Model) renderEndpoints() string {
	var s strings.Builder

	methodColors := map[string]lipgloss.Color{
		"GET":     colorGreen,
		"POST":    colorBlue,
		"PUT":     colorYellow,
		"DELETE":  colorRed,
		"PATCH":   colorPurple,
		"HEAD":    colorGray,
		"OPTIONS": colorGray,
		"TRACE":   colorGray,
	}

	// header (~6 lines) + footer (~4 lines)
	contentHeight := max(1, m.height-10)

	startIdx := m.scrollOffset
	endIdx := min(m.scrollOffset+contentHeight, len(m.endpoints))

	// Add scroll indicator for items above
	if m.scrollOffset > 0 {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGray)).
			Render("⬆ More items above...")
		s.WriteString(indicator)
		s.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		ep := m.endpoints[i]
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.Background(lipgloss.Color(colorBackground))
		}

		methodColor := methodColors[ep.method]
		if methodColor == "" {
			methodColor = colorGray
		}

		methodStyle := lipgloss.NewStyle().
			Foreground(methodColor).
			Bold(true).
			Width(7)

		foldIcon := "▶"
		if !ep.folded {
			foldIcon = "▼"
		}

		line := fmt.Sprintf("%s %s %s",
			foldIcon,
			methodStyle.Render(ep.method),
			ep.path)

		s.WriteString(style.Render(line))
		s.WriteString("\n")

		if !ep.folded {
			details := formatEndpointDetails(ep)
			detailStyle := lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color(colorDetailGray))
			s.WriteString(detailStyle.Render(details))
			s.WriteString("\n")
		}
	}

	// Add scroll indicator for items below
	if endIdx < len(m.endpoints) {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGray)).
			Render("⬇ More items below...")
		s.WriteString(indicator)
		s.WriteString("\n")
	}

	return s.String()
}

func (m Model) renderComponents() string {
	var s strings.Builder

	componentColors := map[string]lipgloss.Color{
		"Schema":         colorGreen,
		"RequestBody":    colorBlue,
		"Response":       colorYellow,
		"Parameter":      colorPurple,
		"Header":         colorRed,
		"SecurityScheme": colorGray,
	}

	contentHeight := max(1, m.height-10)

	startIdx := m.scrollOffset
	endIdx := min(m.scrollOffset+contentHeight, len(m.components))

	// Add scroll indicator for items above
	if m.scrollOffset > 0 {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGray)).
			Render("⬆ More items above...")
		s.WriteString(indicator)
		s.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		comp := m.components[i]
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.Background(lipgloss.Color(colorBackground))
		}

		componentColor := componentColors[comp.compType]
		if componentColor == "" {
			componentColor = colorGray
		}

		typeStyle := lipgloss.NewStyle().
			Foreground(componentColor).
			Bold(true).
			Width(16)

		foldIcon := "▶"
		if !comp.folded {
			foldIcon = "▼"
		}

		line := fmt.Sprintf("%s %s %s",
			foldIcon,
			typeStyle.Render(comp.compType+":"),
			comp.name)

		if comp.description != "" {
			line += fmt.Sprintf(" - %s", comp.description)
		}

		s.WriteString(style.Render(line))
		s.WriteString("\n")

		if !comp.folded {
			detailStyle := lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color(colorDetailGray))
			s.WriteString(detailStyle.Render(comp.details))
			s.WriteString("\n")
		}
	}

	// Add scroll indicator for items below
	if endIdx < len(m.components) {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGray)).
			Render("⬇ More items below...")
		s.WriteString(indicator)
		s.WriteString("\n")
	}

	return s.String()
}

func (m Model) renderWebhooks() string {
	var s strings.Builder

	methodColors := map[string]lipgloss.Color{
		"GET":     colorGreen,
		"POST":    colorBlue,
		"PUT":     colorYellow,
		"DELETE":  colorRed,
		"PATCH":   colorPurple,
		"HEAD":    colorGray,
		"OPTIONS": colorGray,
		"TRACE":   colorGray,
	}

	// header (~6 lines) + footer (~4 lines)
	contentHeight := max(1, m.height-10)

	startIdx := m.scrollOffset
	endIdx := min(m.scrollOffset+contentHeight, len(m.webhooks))

	// Add scroll indicator for items above
	if m.scrollOffset > 0 {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGray)).
			Render("⬆ More items above...")
		s.WriteString(indicator)
		s.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		hook := m.webhooks[i]
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.Background(lipgloss.Color(colorBackground))
		}

		methodColor := methodColors[hook.method]
		if methodColor == "" {
			methodColor = colorGray
		}

		methodStyle := lipgloss.NewStyle().
			Foreground(methodColor).
			Bold(true).
			Width(7)

		foldIcon := "▶"
		if !hook.folded {
			foldIcon = "▼"
		}

		line := fmt.Sprintf("%s %s %s",
			foldIcon,
			methodStyle.Render(hook.method),
			hook.name)

		s.WriteString(style.Render(line))
		s.WriteString("\n")

		if !hook.folded {
			details := formatWebhookDetails(hook)
			detailStyle := lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color(colorDetailGray))
			s.WriteString(detailStyle.Render(details))
			s.WriteString("\n")
		}
	}

	// Add scroll indicator for items below
	if endIdx < len(m.webhooks) {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGray)).
			Render("⬇ More items below...")
		s.WriteString(indicator)
		s.WriteString("\n")
	}

	return s.String()
}

func (m Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorThemePurple)).
		PaddingBottom(1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorGray))

	activeTabStyle := tabStyle.
		BorderForeground(lipgloss.Color(colorThemePurple)).
		Foreground(lipgloss.Color(colorThemePurple))

	var header strings.Builder
	header.WriteString(titleStyle.Render("oq - OpenAPI Spec Viewer"))
	header.WriteString("\n")

	endpointsTab := "Endpoints"
	componentsTab := "Components"
	webhooksTab := "Webhooks"

	// Build tab array based on what's available
	var tabs []string

	if m.mode == viewEndpoints {
		tabs = append(tabs, activeTabStyle.Render(endpointsTab))
	} else {
		tabs = append(tabs, tabStyle.Render(endpointsTab))
	}

	if m.hasWebhooks() {
		if m.mode == viewWebhooks {
			tabs = append(tabs, activeTabStyle.Render(webhooksTab))
		} else {
			tabs = append(tabs, tabStyle.Render(webhooksTab))
		}
	}

	if m.mode == viewComponents {
		tabs = append(tabs, activeTabStyle.Render(componentsTab))
	} else {
		tabs = append(tabs, tabStyle.Render(componentsTab))
	}

	header.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	header.WriteString("\n\n")

	return header.String()
}

func (m Model) renderFooter() string {
	schemaInfo := fmt.Sprintf("%s v%s", m.doc.Info.Title, m.doc.Info.Version)

	helpText := "Press '?' for help"
	if m.showHelp {
		helpText = ""
	}

	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorGray)).
		Foreground(lipgloss.Color(colorFooterText)).
		Padding(0, 1).
		Width(m.width).
		Align(lipgloss.Left)

	availableWidth := m.width - len(schemaInfo) - 4
	if len(helpText) > availableWidth {
		helpText = ""
	}

	footerContent := fmt.Sprintf("%s%s%s",
		helpText,
		strings.Repeat(" ", m.width-len(helpText)-len(schemaInfo)-2),
		schemaInfo)

	return "\n" + footerStyle.Render(footerContent)
}

func (m Model) renderHelpModal() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBlue)).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorWhite))

	helpData := [][]string{
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"gg", "Move to the top"},
		{"G", "Move to the bottom"},
		{"Tab/L", "Cycle forward through views"},
		{"Shift+Tab/H", "Cycle backwards through views"},
		{"Enter/Space", "Toggle details"},
		{"?", "Toggle help"},
		{"Esc/q", "Close help"},
		{"Ctrl+C", "Quit"},
	}

	// Find max width for first column
	maxKeyWidth := 0
	for _, row := range helpData {
		if len(row[0]) > maxKeyWidth {
			maxKeyWidth = len(row[0])
		}
	}

	var helpItems []string
	for _, row := range helpData {
		key := keyStyle.Render(fmt.Sprintf("%-*s", maxKeyWidth, row[0]))
		desc := textStyle.Render(" " + row[1])
		helpItems = append(helpItems, key+desc)
	}

	helpContent := strings.Join(helpItems, "\n")

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorThemePurple)).
		Padding(1, 2).
		Width(45)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorThemePurple)).
		Align(lipgloss.Center).
		Width(28)

	title := titleStyle.Render("Help")
	modal := modalStyle.Render(title + "\n\n" + helpContent)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}
