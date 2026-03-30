package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/itchyny/gojq"
)

// PrintJSON writes data to w according to jsonOpts:
//   - If JQExpr is set, applies the jq filter and prints results.
//   - If Fields is set, selects only the requested fields.
//   - Otherwise pretty-prints the full JSON.
//
// data must be JSON-marshalable.
func PrintJSON(w io.Writer, data any, fields string, jqExpr string) error {
	// Marshal to JSON first
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling to JSON: %w", err)
	}

	// Select specific fields if requested
	if fields != "" && jqExpr == "" {
		raw, err = selectFields(raw, fields)
		if err != nil {
			return err
		}
	}

	// Apply jq expression if given
	if jqExpr != "" {
		return applyJQ(w, raw, jqExpr)
	}

	// Pretty-print
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, buf.String())
	return err
}

// applyJQ runs the jq expression against the JSON data and writes results line-by-line.
func applyJQ(w io.Writer, raw []byte, expr string) error {
	q, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid jq expression %q: %w", expr, err)
	}

	var input any
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("parsing JSON for jq: %w", err)
	}

	code, err := gojq.Compile(q)
	if err != nil {
		return fmt.Errorf("compiling jq expression: %w", err)
	}

	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if jqErr, ok := v.(error); ok {
			return fmt.Errorf("jq error: %w", jqErr)
		}
		out, err := gojq.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
	}
	return nil
}

// selectFields filters a JSON object (or array of objects) to only the requested fields.
// fields is a comma-separated list like "id,title,state".
func selectFields(raw []byte, fields string) ([]byte, error) {
	fieldList := splitFields(fields)
	if len(fieldList) == 0 {
		return raw, nil
	}

	// Build a jq expression: {id, title, state}
	jqExpr := "{" + strings.Join(fieldList, ", ") + "}"

	// For arrays, wrap: [.[] | {id, title}]
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		jqExpr = "[.[] | " + jqExpr + "]"
	}

	var buf bytes.Buffer
	if err := applyJQ(&buf, raw, jqExpr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func splitFields(fields string) []string {
	var result []string
	for _, f := range strings.Split(fields, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}
