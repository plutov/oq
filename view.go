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

	// Calculate available content height using shared function
	contentHeight := calculateContentHeight(m.height)

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

	// Calculate available content height using shared function
	contentHeight := calculateContentHeight(m.height)

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

	// Calculate available content height using shared function
	contentHeight := calculateContentHeight(m.height)

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
	// Button styles for navigation
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(colorGray))

	activeButtonStyle := buttonStyle.
		Background(lipgloss.Color(colorThemePurple)).
		Foreground(lipgloss.Color(colorWhite)).
		Bold(true)

	// App title style for right side
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorThemePurple))

	// Build navigation buttons
	var buttons []string

	// Endpoints button
	if m.mode == viewEndpoints {
		buttons = append(buttons, activeButtonStyle.Render("Requests"))
	} else {
		buttons = append(buttons, buttonStyle.Render("Requests"))
	}

	// Webhooks button (only if available)
	if m.hasWebhooks() {
		if m.mode == viewWebhooks {
			buttons = append(buttons, activeButtonStyle.Render("Webhooks"))
		} else {
			buttons = append(buttons, buttonStyle.Render("Webhooks"))
		}
	}

	// Components button
	if m.mode == viewComponents {
		buttons = append(buttons, activeButtonStyle.Render("Components"))
	} else {
		buttons = append(buttons, buttonStyle.Render("Components"))
	}

	// Join buttons with separators
	navSection := strings.Join(buttons, " │ ")

	// App title for right side
	appTitle := titleStyle.Render("oq - OpenAPI Spec Viewer")

	// Calculate total width for proper spacing
	navWidth := lipgloss.Width(navSection)
	titleWidth := lipgloss.Width(appTitle)
	totalContentWidth := navWidth + titleWidth

	// Create the header line with proper spacing
	var headerLine string
	if m.width > totalContentWidth+4 { // 4 for some padding
		spacingWidth := m.width - totalContentWidth
		spacing := strings.Repeat(" ", spacingWidth)
		headerLine = navSection + spacing + appTitle
	} else {
		// If not enough space, just show navigation and truncate title if needed
		headerLine = navSection
	}

	// Return header with one empty line below
	return headerLine + "\n\n"
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
		{"Shift+Tab/H", "Cycle backward through views"},
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
