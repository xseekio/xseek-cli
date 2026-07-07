package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type CatalogProduct struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Region      string   `json:"region"`
	Category    string   `json:"category"`
	Features    []string `json:"features"`
	ImageURL    string   `json:"imageUrl"`
}

type ProductsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Products []CatalogProduct `json:"products"`
		Count    int              `json:"count"`
	} `json:"data"`
}

// ListProducts fetches the client's own product catalog, filtered by
// region/category/query, so B2C generation recommends real products (with
// their own URLs) instead of competitors.
func ListProducts(websiteID, region, category, query, limit string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}
	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if region != "" {
		params["region"] = region
	}
	if category != "" {
		params["category"] = category
	}
	if query != "" {
		params["q"] = query
	}
	if limit != "" {
		params["limit"] = limit
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
		fmt.Println("No products found. Upload a catalog in Brand voice → Products (or adjust --region/--category).")
		return
	}

	if isMarkdown() {
		var b strings.Builder
		fmt.Fprintf(&b, "## Products (%d)\n\n", len(products))
		for _, p := range products {
			fmt.Fprintf(&b, "- [%s](%s)", p.Name, p.URL)
			meta := []string{}
			if p.Region != "" {
				meta = append(meta, p.Region)
			}
			if p.Category != "" {
				meta = append(meta, p.Category)
			}
			if len(meta) > 0 {
				fmt.Fprintf(&b, " — %s", strings.Join(meta, ", "))
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
		meta := []string{}
		if p.Region != "" {
			meta = append(meta, "region: "+p.Region)
		}
		if p.Category != "" {
			meta = append(meta, "category: "+p.Category)
		}
		if len(meta) > 0 {
			fmt.Printf("    %s\n", strings.Join(meta, " | "))
		}
	}
}
