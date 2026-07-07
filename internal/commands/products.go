package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type CatalogProduct struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Description string            `json:"description"`
	Region      string            `json:"region"`
	Category    string            `json:"category"`
	Features    []string          `json:"features"`
	ImageURL    string            `json:"imageUrl"`
	Attributes  map[string]string `json:"attributes"`
}

type ProductsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Products []CatalogProduct `json:"products"`
		Count    int              `json:"count"`
	} `json:"data"`
}

// productFilterReserved are flags that are NOT product-field filters.
var productFilterReserved = map[string]bool{"format": true, "help": true, "no-browser": true}

// ListProducts fetches the client's own product catalog, filtered by any field
// the client uploaded (region, category, or a custom attribute like capacity /
// make / year). Every `--field value` flag becomes a filter, so B2C generation
// recommends real products (with their own URLs) instead of competitors.
func ListProducts(websiteID string, flags map[string]string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}
	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	for k, v := range flags {
		if productFilterReserved[k] || v == "" {
			continue
		}
		if k == "query" {
			k = "q"
		}
		params[k] = v // region/category/limit/q or any custom field → query param
	}

	var result ProductsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/products", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result.Data)
		return
	}

	products := result.Data.Products
	if len(products) == 0 {
		fmt.Println("No products found. Upload a catalog in Brand voice → Products (or adjust your --field filters).")
		return
	}

	// meta assembles the filterable fields for a product (region, category, and
	// any custom attributes) for display.
	meta := func(p CatalogProduct, sep string) string {
		parts := []string{}
		if p.Region != "" {
			parts = append(parts, "region: "+p.Region)
		}
		if p.Category != "" {
			parts = append(parts, "category: "+p.Category)
		}
		for k, v := range p.Attributes {
			parts = append(parts, k+": "+v)
		}
		return strings.Join(parts, sep)
	}

	if isMarkdown() {
		var b strings.Builder
		fmt.Fprintf(&b, "## Products (%d)\n\n", len(products))
		for _, p := range products {
			fmt.Fprintf(&b, "- [%s](%s)", p.Name, p.URL)
			if m := meta(p, ", "); m != "" {
				fmt.Fprintf(&b, " — %s", m)
			}
			b.WriteString("\n")
		}
		fmt.Print(b.String())
		return
	}

	fmt.Printf("Products: %d\n", len(products))
	fmt.Println(strings.Repeat("─", 60))
	for _, p := range products {
		fmt.Printf("  • %s\n", p.Name)
		fmt.Printf("    %s\n", p.URL)
		if m := meta(p, " | "); m != "" {
			fmt.Printf("    %s\n", m)
		}
	}
}
