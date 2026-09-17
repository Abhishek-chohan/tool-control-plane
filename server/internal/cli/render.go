package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format is the output selection every verb carries: human tables by
// default, JSON when scripts ask (--format json). Both shapes carry the
// same fields.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// ParseFormat validates a --format value.
func ParseFormat(value string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(FormatTable):
		return FormatTable, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("invalid --format %q: want table or json", value)
	}
}

// Renderer writes a verb's result in the selected format. Table mode is
// for humans (columns aligned, one row per record); JSON mode is the
// stable machine shape — pipeable into jq, never prose.
type Renderer struct {
	Out    io.Writer
	Format Format
}

// Table renders column headers plus one line per row, space-aligned to
// the widest cell. Rows are []string cells; a nil header skips the header
// line.
func (r Renderer) Table(header []string, rows [][]string) {
	if r.Format == FormatJSON || r.Out == nil {
		return
	}
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	line := func(cells []string) {
		var b strings.Builder
		for i, cell := range cells {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(cell)
			if i < len(widths)-1 {
				b.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
			}
		}
		fmt.Fprintln(r.Out, strings.TrimRight(b.String(), " "))
	}
	if header != nil {
		line(header)
		seps := make([]string, len(header))
		for i, w := range widths {
			seps[i] = strings.Repeat("-", w)
		}
		line(seps)
	}
	for _, row := range rows {
		line(row)
	}
}

// JSON renders one value as pretty-printed JSON with a trailing newline.
func (r Renderer) JSON(value interface{}) error {
	if r.Format != FormatJSON || r.Out == nil {
		return nil
	}
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(r.Out, string(encoded))
	return err
}

// Emit writes a verb result: JSON in JSON mode, the table in table mode.
func (r Renderer) Emit(header []string, rows [][]string, jsonValue interface{}) error {
	if r.Format == FormatJSON {
		return r.JSON(jsonValue)
	}
	r.Table(header, rows)
	return nil
}
