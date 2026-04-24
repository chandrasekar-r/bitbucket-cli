package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Table renders aligned columns to a writer using tabwriter.
// Use AddRow to accumulate rows, then Render to flush.
//
// Example:
//
//	t := output.NewTable(os.Stdout)
//	t.AddHeader("NAME", "STATE", "AUTHOR")
//	t.AddRow("fix-login", "OPEN", "alice")
//	t.Render()
type Table struct {
	w       *tabwriter.Writer
	headers []string
	rows    [][]string
	color   bool
}

// NewTable creates a table that writes to w.
// Set color=true to enable ANSI header formatting when stdout is a TTY.
func NewTable(w io.Writer, color bool) *Table {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	return &Table{w: tw, color: color}
}

// AddHeader sets the column headers (printed once before rows).
func (t *Table) AddHeader(cols ...string) {
	t.headers = cols
}

// AddRow appends a data row. Column count should match headers.
func (t *Table) AddRow(cols ...string) {
	t.rows = append(t.rows, cols)
}

// Render writes all headers and rows to the underlying writer.
func (t *Table) Render() error {
	if len(t.headers) > 0 {
		headers := t.headers
		if t.color {
			for i, h := range headers {
				headers[i] = bold(h)
			}
		}
		fmt.Fprintln(t.w, strings.Join(headers, "\t"))
	}
	for _, row := range t.rows {
		fmt.Fprintln(t.w, strings.Join(row, "\t"))
	}
	return t.w.Flush()
}

// bold wraps a string in ANSI bold escape codes.
func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}
