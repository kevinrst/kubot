package render

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
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

func PrintResourcesTable(w io.Writer, rows []ResourceRow, color bool, maxRows int) {
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
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, st.dim("POD\tCONTAINER\tCPU-REQ\tMEM-REQ\tMEM-LIMIT\tMEM-USE\tUSE%"))
	for _, r := range shown {
		use := r.MemUse
		pct := "-"
		if r.UseRatio >= 0 {
			pct = fmt.Sprintf("%.0f%%", r.UseRatio*100)
			if r.Hot {
				pct = st.warn(pct)
			}
		}
		limit := r.MemLimit
		if limit == "-" {
			limit = st.dim("none")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Pod, r.Container, r.CPUReq, r.MemReq, limit, use, pct)
	}
	tw.Flush()
	if hidden > 0 {
		fmt.Fprintf(w, "%s\n", st.dim(fmt.Sprintf("… and %d more containers (use -n to narrow)", hidden)))
	}
}
