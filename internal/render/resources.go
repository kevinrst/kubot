package render

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type ResourceRow struct {
	Pod       string
	Container string
	CPUReq    string
	MemReq    string
	MemLimit  string
	MemUse    string
	UseRatio  float64
	Hot       bool
}

const useCol = 6

func PrintResourcesTable(w io.Writer, rows []ResourceRow, color bool, maxRows, width int) {
	st := styler{on: color}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Pod != rows[j].Pod {
			return rows[i].Pod < rows[j].Pod
		}
		return rows[i].Container < rows[j].Container
	})
	shown := rows
	hidden := 0
	if maxRows > 0 && len(rows) > maxRows {
		shown, hidden = rows[:maxRows], len(rows)-maxRows
	}
	t, p := baseTable(color)
	t.Headers("POD", "CONTAINER", "CPU-REQ", "MEM-REQ", "MEM-LIMIT", "MEM-USE", "USE%").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return p.head()
			}
			if col == useCol && row < len(shown) && shown[row].Hot {
				return p.warn()
			}
			if shown[row].MemLimit == "-" && col == 4 {
				return p.dim()
			}
			return p.plain()
		})
	for _, r := range shown {
		pct := "-"
		if r.UseRatio >= 0 {
			pct = fmt.Sprintf("%.0f%%", r.UseRatio*100)
		}
		limit := r.MemLimit
		if limit == "-" {
			limit = "none"
		}
		t.Row(r.Pod, r.Container, r.CPUReq, r.MemReq, limit, r.MemUse, pct)
	}
	out := t.Render()
	if width > 0 {
		if w := lipgloss.Width(out); w > width-2 {
			t.Width(width - 2).Wrap(true)
			out = t.Render()
		}
	}
	fmt.Fprintln(w, out)
	if hidden > 0 {
		fmt.Fprintf(w, "%s\n", st.dim(fmt.Sprintf("… and %d more containers (use -n to narrow)", hidden)))
	}
}
