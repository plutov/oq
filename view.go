package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderEndpoints() string {
	var s strings.Builder

	methodColors := map[string]lipgloss.Color{
		"GET":     "#10B981",
		"POST":    "#3B82F6",
		"PUT":     "#F59E0B",
		"DELETE":  "#EF4444",
		"PATCH":   "#8B5CF6",
		"HEAD":    "#6B7280",
		"OPTIONS": "#6B7280",
		"TRACE":   "#6B7280",
	}

	for i, ep := range m.endpoints {
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.Background(lipgloss.Color("#374151"))
		}

		methodColor := methodColors[ep.method]
		if methodColor == "" {
			methodColor = "#6B7280"
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
				Foreground(lipgloss.Color("#9CA3AF"))
			s.WriteString(detailStyle.Render(details))
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m Model) renderComponents() string {
	var s strings.Builder

	componentColors := map[string]lipgloss.Color{
		"Schema":         "#10B981",
		"RequestBody":    "#3B82F6",
		"Response":       "#F59E0B",
		"Parameter":      "#8B5CF6",
		"Header":         "#EF4444",
		"SecurityScheme": "#6B7280",
	}

	for i, comp := range m.components {
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.Background(lipgloss.Color("#374151"))
		}

		componentColor := componentColors[comp.compType]
		if componentColor == "" {
			componentColor = "#6B7280"
		}

		typeStyle := lipgloss.NewStyle().
			Foreground(componentColor).
			Bold(true).
			Width(12)

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
				Foreground(lipgloss.Color("#9CA3AF"))
			s.WriteString(detailStyle.Render(comp.details))
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		PaddingBottom(1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666"))

	activeTabStyle := tabStyle.
		BorderForeground(lipgloss.Color("#7C3AED")).
		Foreground(lipgloss.Color("#7C3AED"))

	title := fmt.Sprintf("OpenAPI Schema Viewer - %s %s",
		m.doc.Info.Title, m.doc.Info.Version)

	var header strings.Builder
	header.WriteString(titleStyle.Render(title))
	header.WriteString("\n")

	endpointsTab := "Endpoints"
	componentsTab := "Components"

	if m.mode == viewEndpoints {
		endpointsTab = activeTabStyle.Render(endpointsTab)
		componentsTab = tabStyle.Render(componentsTab)
	} else {
		endpointsTab = tabStyle.Render(endpointsTab)
		componentsTab = activeTabStyle.Render(componentsTab)
	}

	header.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, endpointsTab, " ", componentsTab))
	header.WriteString("\n\n")

	return header.String()
}

func (m Model) renderFooter() string {
	return "\n\nPress 'tab' to switch views, 'enter' to toggle details, 'q' to quit"
}
