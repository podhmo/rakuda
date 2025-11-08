package rakuda

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// PrintRoutes prints a formatted table of all registered routes to the provided writer.
func PrintRoutes(w io.Writer, b *Builder) {
	// Format:
	// METHOD <2 spaces> PATTERN
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	b.Walk(func(method, pattern string) {
		fmt.Fprintf(tw, "%s\t%s\n", strings.ToUpper(method), pattern)
	})
}
