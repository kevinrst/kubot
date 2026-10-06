package render

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type EventRow struct {
	Namespace string
	Object    string
	Reason    string
	Count     int32
	Message   string
}

// Column budgets. Fixed columns are cut to fit (object in the middle, so the
// workload prefix and unique suffix survive)
const (
	evNS     = 16
	evObject = 34
	evReason = 20
	evCount  = 6
	evChrome = 12 // borders + cell padding
)

func PrintEventsTable(w io.Writer, rows []EventRow, color bool, maxRows, width int) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Object < rows[j].Object
	})
	shown := rows
	if maxRows > 0 && len(rows) > maxRows {
		shown = rows[:maxRows]
	}
	msgWidth := 100
	if width > 0 {
		if m := width - 2 - (evNS + evObject + evReason + evCount + evChrome); m > 20 {
			msgWidth = m
		} else {
			msgWidth = 20
		}
	}
	t, p := baseTable(color)
	t.Headers("NS", "OBJECT", "REASON", "COUNT", "MESSAGE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return p.head()
			}
			if col == 3 {
				return p.warn()
			}
			return p.plain()
		})
	for _, r := range shown {
		t.Row(
			cutEnd(r.Namespace, evNS),
			cutMiddle(r.Object, evObject),
			cutEnd(r.Reason, evReason),
			fmt.Sprintf("x%d", r.Count),
			cutEnd(r.Message, msgWidth),
		)
	}
	fmt.Fprintln(w, t.Render())
}

func cutEnd(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// cutMiddle keeps the head and tail: pod names carry the workload up front
// and the unique id at the end, the hash in the middle is expendable.
func cutMiddle(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n < 10 {
		return cutEnd(s, n)
	}
	keep := n - 1
	head := keep * 2 / 3
	return string(r[:head]) + "…" + string(r[len(r)-(keep-head):])
}
