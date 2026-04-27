package commands

import (
	"encoding/json"
	"fmt"
	"os"
)

// OutputFormat controls how results are displayed
var OutputFormat string

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func exitError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

func isJSON() bool {
	return OutputFormat == "json"
}

func isMarkdown() bool {
	return OutputFormat == "markdown" || OutputFormat == "md"
}
