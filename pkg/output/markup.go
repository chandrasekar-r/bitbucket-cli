package output

// BitbucketMarkdownGuide returns a formatting reference for Bitbucket Cloud
// comment bodies. The content is passed verbatim as content.raw to the API;
// Bitbucket renders Markdown server-side.
func BitbucketMarkdownGuide() string {
	return `Bitbucket Markdown Formatting Reference
========================================

Text style
  **bold**           __bold__
  *italic*           _italic_
  ~~strikethrough~~
  `+"`"+`inline code`+"`"+`

Headings
  # H1
  ## H2
  ### H3

Lists
  - unordered item       * unordered item
  1. ordered item

Code blocks
  ` + "```" + `go
  fmt.Println("hello")
  ` + "```" + `

  Supported languages: go, python, java, javascript, ts, bash, sql, yaml, json, …

Tables
  | column | column |
  |--------|--------|
  | cell   | cell   |

Links & images
  [link text](https://example.com)
  ![alt text](https://example.com/image.png)

Horizontal rule
  ---

Bitbucket-specific
  @username              — mention a user
  #123                   — link to issue #123
  pull request #7        — link to PR #7
  9cc27f2                — link to commit SHA
  :emoji_name:           — emoji shortcode (e.g. :tada: :white_check_mark:)

Note: only content.raw is sent to the API; content.markup and content.html
are output-only fields and must not be included in requests.
`
}
