package common

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	reset = "\033[0m"
	bold  = "\033[1m"
	cyan  = "\033[36m"
	green = "\033[32m"
	gray  = "\033[90m"
)

type BannerConfig struct {
	AppName     string
	Description string
	Version     string
	Environment string
	Port        string
	URL         string
}

func PrintBanner(config BannerConfig) {
	fmt.Println(cyan + bold)
	fmt.Println("███╗   ██╗███████╗██╗  ██╗ ██████╗  ██████╗ ██╗   ██╗")
	fmt.Println("████╗  ██║██╔════╝╚██╗██╔╝██╔════╝ ██╔═══██╗██║   ██║")
	fmt.Println("██╔██╗ ██║█████╗   ╚███╔╝ ██║  ███╗██║   ██║██║   ██║")
	fmt.Println("██║╚██╗██║██╔══╝   ██╔██╗ ██║   ██║██║   ██║██║   ██║")
	fmt.Println("██║ ╚████║███████╗██╔╝ ██╗╚██████╔╝╚██████╔╝╚██████╔╝")
	fmt.Println("╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝  ╚═════╝ ")
	fmt.Println(reset)

	fmt.Printf("%s%s%s\n", bold, config.Description, reset)
	fmt.Println(gray + "────────────────────────────────" + reset)

	printRow("Version", config.Version)
	printRow("Env", config.Environment)
	printRow("Port", config.Port)
	printRow("URL", config.URL)
	printRow("Go", runtime.Version())
	printRow("OS", runtime.GOOS+"/"+runtime.GOARCH)
	printRow("PID", fmt.Sprintf("%d", os.Getpid()))
	printRow("Started", time.Now().Format("15:04:05"))

	fmt.Println(gray + "────────────────────────────────" + reset)
	fmt.Println(green + "Ready." + reset)
	fmt.Println()
}

func printRow(label string, value string) {
	fmt.Printf("%s%-9s%s %s\n", green, label+":", reset, value)
}