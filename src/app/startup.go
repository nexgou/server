package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/router"
)

// PrintBanner writes the NexGou startup banner to stdout.
func (app *App) PrintBanner(config common.BannerConfig) {
	common.PrintBanner(config)
}

// PrintRoutes writes the registered HTTP routes to stdout.
func (app *App) PrintRoutes() {
	app.WriteRoutes(os.Stdout)
}

// WriteRoutes writes the registered HTTP routes to the given writer.
func (app *App) WriteRoutes(writer io.Writer) {
	writeRoutes(writer, app.router.Routes())
}

func writeRoutes(writer io.Writer, routes []router.RouteInfo) {
	reset := "\033[0m"
	dim := "\033[2m"
	gray := "\033[90m"
	methodColor := map[string]string{
		"GET":     "\033[32m",
		"POST":    "\033[34m",
		"PUT":     "\033[33m",
		"PATCH":   "\033[35m",
		"DELETE":  "\033[31m",
		"OPTIONS": "\033[36m",
	}

	maxMethod := 6
	maxPath := 15
	for _, route := range routes {
		if len(route.Method) > maxMethod {
			maxMethod = len(route.Method)
		}
		if len(route.Path) > maxPath {
			maxPath = len(route.Path)
		}
	}

	_, _ = fmt.Fprintf(writer, "Mapped routes:\n")
	_, _ = fmt.Fprintf(writer, "%s--------------------------------%s\n", gray, reset)
	if len(routes) == 0 {
		_, _ = fmt.Fprintf(writer, "%s(no routes registered)%s\n\n", gray, reset)
		return
	}

	for _, route := range routes {
		color := methodColor[route.Method]
		if color == "" {
			color = "\033[37m"
		}
		method := fmt.Sprintf("%-*s", maxMethod, route.Method)
		path := fmt.Sprintf("%-*s", maxPath, route.Path)
		_, _ = fmt.Fprintf(writer, "%s%s%s   %s%s%s   %s\n", color, method, reset, dim, path, reset, routeBadges(route))
	}
	_, _ = fmt.Fprintf(writer, "\n")
}

func routeBadges(route router.RouteInfo) string {
	badges := make([]string, 0, 3)
	if len(route.Guards) == 0 {
		badges = append(badges, "public")
	} else {
		badges = append(badges, "guards: "+strings.Join(route.Guards, ", "))
	}
	if len(route.Pipes) > 0 {
		badges = append(badges, "pipes: "+strings.Join(route.Pipes, ", "))
	}
	if len(route.Interceptors) > 0 {
		badges = append(badges, "interceptors: "+strings.Join(route.Interceptors, ", "))
	}
	return strings.Join(badges, " | ")
}
