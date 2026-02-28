package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikburt/csvmap/internal/config"
	"github.com/erikburt/csvmap/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "1.0.0"

func main() {
	// Define flags
	hasHeader := flag.Bool("has-header", false, "CSV file has a header row")
	noHeader := flag.Bool("no-header", false, "CSV file does not have a header row")
	configDir := flag.String("config-dir", "", "Override default config directory")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "csvmap - Interactive CSV column mapping tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: csvmap <input.csv> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  input.csv    Path to the input CSV file\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  csvmap statement.csv\n")
		fmt.Fprintf(os.Stderr, "  csvmap statement.csv --has-header\n")
		fmt.Fprintf(os.Stderr, "  csvmap data.csv --no-header --config-dir /path/to/config\n")
	}

	flag.Parse()

	// Handle version and help
	if *showVersion {
		fmt.Printf("csvmap version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Validate arguments
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: input CSV file required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *hasHeader && *noHeader {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both --has-header and --no-header\n")
		os.Exit(1)
	}

	csvPath := args[0]

	// Check if file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", csvPath)
		os.Exit(1)
	}

	// Determine header status
	var headerStatus *bool
	if *hasHeader {
		h := true
		headerStatus = &h
	} else if *noHeader {
		h := false
		headerStatus = &h
	}

	// Initialize config
	cfg, err := config.New(*configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %s\n", err)
		os.Exit(1)
	}

	// Create and run TUI
	model := tui.NewModel(csvPath, headerStatus, cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %s\n", err)
		os.Exit(1)
	}
}
