package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

// sitePagesReserved are flags that are not API filters.
var sitePagesReserved = map[string]bool{"format": true, "help": true, "no-browser": true}

// apiParamName turns a CLI flag into the API's query parameter:
// --page-kind -> pageKind, --max-own-words -> maxOwnWords. A flag already in
// camelCase (--pageKind) passes through as typed.
func apiParamName(flag string) string {
	parts := strings.Split(flag, "-")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// SitePages reads the crawled pages of a site as a crawler without JavaScript
// sees them (GPTBot, ClaudeBot and PerplexityBot do not run it).
//
// Every flag is forwarded to GET /websites/{id}/site-pages as a query
// parameter, so a filter the API gains works here the day it ships, with no
// new CLI release: --sells true, --page-kind service, --issue noindex,
// --google-indexed false, --max-own-words 300, --has-overlap true, --q "...",
// --url <page> [--markdown true], --page / --page-size.
//
// No filter: the site summary. A filter: the matching pages. --url: one page.
// The output is the API's answer as sent; --format text prints a short list.
func SitePages(websiteID string, flags map[string]string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}
	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	for k, v := range flags {
		if sitePagesReserved[k] || v == "" {
			continue
		}
		params[apiParamName(k)] = v
	}

	body, err := client.Get(fmt.Sprintf("/websites/%s/site-pages", websiteID), params)
	if err != nil {
		exitError(err.Error())
	}

	if OutputFormat == "text" {
		var list struct {
			Total int `json:"total"`
			Pages []struct {
				URL      string `json:"url"`
				Status   *int   `json:"status"`
				Title    string `json:"title"`
				PageKind string `json:"pageKind"`
			} `json:"pages"`
		}
		if json.Unmarshal(body, &list) == nil && list.Pages != nil {
			fmt.Printf("%d page(s)\n", list.Total)
			for _, p := range list.Pages {
				status := "-"
				if p.Status != nil {
					status = fmt.Sprintf("%d", *p.Status)
				}
				fmt.Printf("  %s  %-10s %s\n", status, p.PageKind, p.URL)
				if p.Title != "" {
					fmt.Printf("       %s\n", truncate(p.Title, 100))
				}
			}
			return
		}
	}

	// The API's answer as sent: every field, null kept as null (unmeasured).
	var out bytes.Buffer
	if json.Indent(&out, body, "", "  ") != nil {
		os.Stdout.Write(body)
		return
	}
	out.WriteByte('\n')
	os.Stdout.Write(out.Bytes())
}
