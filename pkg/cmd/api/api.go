package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	bbapi "github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdAPI creates the `bb api` command for raw authenticated API requests.
func NewCmdAPI(f *cmdutil.Factory) *cobra.Command {
	var (
		method       string
		fields       []string
		headers      []string
		paginate     bool
		jqExpr       string
		input        string
	)

	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make an authenticated API request",
		Long: `Make an authenticated HTTP request to the Bitbucket API.

The endpoint argument is appended to "https://api.bitbucket.org/2.0".
Placeholders {workspace} and {repo} are replaced from git context.

The default HTTP method is GET, or POST if --field or --input is provided.`,
		Example: `  # Get current user
  bb api /user

  # List repositories (paginated)
  bb api /repositories/{workspace} --paginate

  # Create a pull request
  bb api /repositories/{workspace}/{repo}/pullrequests \
    -f title="My PR" -f source.branch.name=feature -f destination.branch.name=main

  # Use jq to filter output
  bb api /repositories/{workspace}/{repo}/pullrequests --jq '.values[].title'

  # Send a request body from a file
  bb api /repositories/{workspace}/{repo}/pullrequests --input body.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := replacePlaceholders(args[0])

			// Auto-detect method: POST if fields or input provided, else GET
			if method == "" {
				if len(fields) > 0 || input != "" {
					method = "POST"
				} else {
					method = "GET"
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			// Handle paginate mode
			if paginate {
				if method != "GET" {
					return fmt.Errorf("--paginate is only supported for GET requests")
				}
				return doPaginated(httpClient, f.BaseURL, endpoint, jqExpr, f.IOStreams.Out, headers)
			}

			// Build request body
			var body io.Reader
			if input != "" {
				inputFile, err := os.Open(input)
				if err != nil {
					return fmt.Errorf("opening input file: %w", err)
				}
				defer inputFile.Close()
				body = inputFile
			} else if len(fields) > 0 {
				nested := buildNestedFields(fields)
				data, err := json.Marshal(nested)
				if err != nil {
					return fmt.Errorf("marshaling fields: %w", err)
				}
				body = bytes.NewReader(data)
			}

			// Build request
			url := f.BaseURL + endpoint
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				return fmt.Errorf("creating request: %w", err)
			}
			req.Header.Set("Accept", "application/json")
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			// Error responses: status line to stderr, body to stdout, exit 1
			if resp.StatusCode >= 400 {
				fmt.Fprintf(f.IOStreams.ErrOut, "HTTP %d\n", resp.StatusCode)
				f.IOStreams.Out.Write(respBody)
				fmt.Fprintln(f.IOStreams.Out)
				return &cmdutil.FlagError{Err: fmt.Errorf("HTTP %d", resp.StatusCode)}
			}

			// Apply jq filter if provided
			if jqExpr != "" {
				var data any
				if err := json.Unmarshal(respBody, &data); err != nil {
					// Not JSON — write raw
					f.IOStreams.Out.Write(respBody)
					fmt.Fprintln(f.IOStreams.Out)
					return nil
				}
				return output.PrintJSON(f.IOStreams.Out, data, "", jqExpr)
			}

			// Pretty-print JSON by default
			var prettyBuf bytes.Buffer
			if json.Indent(&prettyBuf, respBody, "", "  ") == nil {
				fmt.Fprintln(f.IOStreams.Out, prettyBuf.String())
			} else {
				f.IOStreams.Out.Write(respBody)
				fmt.Fprintln(f.IOStreams.Out)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "", "HTTP method (default: GET, or POST if --field/--input is set)")
	cmd.Flags().StringArrayVarP(&fields, "field", "f", nil, "Request body field in key=value format (supports nested keys via dots)")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Additional HTTP header in key:value format")
	cmd.Flags().BoolVar(&paginate, "paginate", false, "Fetch all pages and concatenate values arrays (GET only)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter output with a jq expression")
	cmd.Flags().StringVar(&input, "input", "", "File path to use as request body (- for stdin)")

	return cmd
}

// replacePlaceholders substitutes {workspace} and {repo} from git context.
func replacePlaceholders(endpoint string) string {
	if !strings.Contains(endpoint, "{workspace}") && !strings.Contains(endpoint, "{repo}") {
		return endpoint
	}
	ctx := gitcontext.FromRemote()
	if ctx == nil {
		return endpoint
	}
	endpoint = strings.ReplaceAll(endpoint, "{workspace}", ctx.Workspace)
	endpoint = strings.ReplaceAll(endpoint, "{repo}", ctx.RepoSlug)
	return endpoint
}

// buildNestedFields turns ["title=My PR", "source.branch.name=feature"] into
// {"title":"My PR","source":{"branch":{"name":"feature"}}}.
func buildNestedFields(fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		keys := strings.Split(key, ".")
		setNested(result, keys, value)
	}
	return result
}

// setNested sets a value at the given key path in the object.
func setNested(obj map[string]interface{}, keys []string, value string) {
	for i, key := range keys {
		if i == len(keys)-1 {
			obj[key] = value
			return
		}
		next, ok := obj[key]
		if !ok {
			next = make(map[string]interface{})
			obj[key] = next
		}
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			nextMap = make(map[string]interface{})
			obj[key] = nextMap
		}
		obj = nextMap
	}
}

// doPaginated fetches all pages and outputs the concatenated values array.
func doPaginated(httpClient *http.Client, baseURL, path, jqExpr string, w io.Writer, extraHeaders []string) error {
	client := bbapi.New(httpClient, baseURL)
	items, err := bbapi.PaginateAll(client, path, 0)
	if err != nil {
		return err
	}

	// Convert []json.RawMessage to []interface{} for PrintJSON
	var all []interface{}
	for _, raw := range items {
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("parsing paginated item: %w", err)
		}
		all = append(all, v)
	}

	if jqExpr != "" {
		return output.PrintJSON(w, all, "", jqExpr)
	}

	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}
