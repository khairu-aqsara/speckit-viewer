package ui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
)

func dashboardColumns(width int) []table.Column {
	nameWidth := max(20, width-6-18-22-8)
	return []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: nameWidth},
		{Title: "Phase", Width: 18},
		{Title: "Progress", Width: 20},
	}
}

func dashboardRows(features []featureVM) []table.Row {
	rows := make([]table.Row, 0, len(features))
	for _, f := range features {
		progress := "—"
		if f.Tasks != nil && f.PhaseResult.Total > 0 {
			progress = progressCell(f.PhaseResult.Checked, f.PhaseResult.Total)
		}
		rows = append(rows, table.Row{f.ID, f.Name, f.PhaseResult.Label(), progress})
	}
	return rows
}

func progressCell(checked, total int) string {
	const width = 10
	filled := checked * width / total
	bar := ""
	for i := range width {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return fmt.Sprintf("%s %d/%d", bar, checked, total)
}
