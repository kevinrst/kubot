package render

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func baseTable(color bool) (*table.Table, palette) {
	p := palette{on: color}
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(p.border()).
		BorderColumn(false)
	return t, p
}

type palette struct{ on bool }

func (p palette) style(code string, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Padding(0, 1)
	if !p.on {
		return s
	}
	s = s.Foreground(lipgloss.Color(code))
	if bold {
		s = s.Bold(true)
	}
	return s
}

func (p palette) dim() lipgloss.Style    { return p.style("243", false) }
func (p palette) head() lipgloss.Style   { return p.style("243", true) }
func (p palette) warn() lipgloss.Style   { return p.style("216", true) }
func (p palette) plain() lipgloss.Style  { return lipgloss.NewStyle().Padding(0, 1) }
func (p palette) border() lipgloss.Style { return p.style("243", false).Padding(0, 0) }
