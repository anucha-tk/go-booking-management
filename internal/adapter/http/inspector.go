package http

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gin-gonic/gin"
)

// InspectRoutes prints all registered routes in a formatted table.
// It can also export the routes to a JSON file if needed.
func InspectRoutes(routes gin.RoutesInfo, filter string) {
	fmt.Println("\n🔍 API ROUTE INSPECTOR")
	fmt.Println(strings.Repeat("=", 60))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "METHOD\tPATH\tHANDLER")

	count := 0
	for _, route := range routes {
		if filter != "" && !strings.Contains(route.Path, filter) && !strings.Contains(route.Method, filter) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", route.Method, route.Path, route.Handler)
		count++
	}
	w.Flush()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Routes: %d\n\n", count)
}

// ExportRoutes exports the routes to a JSON file.
func ExportRoutes(routes gin.RoutesInfo, filePath string) error {
	data, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
