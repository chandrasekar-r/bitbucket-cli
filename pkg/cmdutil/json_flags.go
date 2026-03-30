package cmdutil

import (
	"github.com/spf13/cobra"
)

// JSONOptions holds the parsed --json and --jq flag values.
type JSONOptions struct {
	// Fields is the comma-separated list of fields requested via --json.
	// Empty means JSON output was not requested.
	Fields string
	// JQExpr is the jq filter expression from --jq.
	JQExpr string
}

// Enabled reports whether JSON output was requested (either --json or --jq was set).
func (o *JSONOptions) Enabled() bool {
	return o.Fields != "" || o.JQExpr != ""
}

// AddJSONFlags registers --json and --jq flags on the command and returns a pointer
// to the JSONOptions that will be populated when the command runs.
//
// Usage pattern:
//
//	jsonOpts := cmdutil.AddJSONFlags(cmd)
//	// In RunE: use jsonOpts.Enabled(), jsonOpts.Fields, jsonOpts.JQExpr
func AddJSONFlags(cmd *cobra.Command) *JSONOptions {
	opts := &JSONOptions{}
	cmd.Flags().StringVar(
		&opts.Fields,
		"json",
		"",
		"Output JSON for the specified fields (comma-separated). Use --json --help to see available fields.",
	)
	cmd.Flags().StringVar(
		&opts.JQExpr,
		"jq",
		"",
		"Filter JSON output with a jq expression. Implies --json.",
	)
	return opts
}
