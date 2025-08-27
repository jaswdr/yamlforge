package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yamlforge/yamlforge/internal/parser"
	"github.com/yamlforge/yamlforge/internal/server"
)

const version = "0.1.0"

func main() {
	var (
		port        int
		host        string
		showHelp    bool
		showVersion bool
	)

	flag.IntVar(&port, "port", 8080, "Server port")
	flag.StringVar(&host, "host", "0.0.0.0", "Server host")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	if showVersion {
		fmt.Printf("yamlforge v%s\n", version)
		os.Exit(0)
	}

	if showHelp || flag.NArg() < 1 {
		printUsage()
		os.Exit(0)
	}

	command := flag.Arg(0)

	switch command {
	case "serve":
		if flag.NArg() < 2 {
			fmt.Println("Error: missing YAML configuration file")
			printUsage()
			os.Exit(1)
		}
		configFile := flag.Arg(1)
		handleServe(configFile, port, host)

	case "build":
		if flag.NArg() < 2 {
			fmt.Println("Error: missing YAML configuration file")
			printUsage()
			os.Exit(1)
		}
		configFile := flag.Arg(1)
		handleBuild(configFile)

	case "validate":
		if flag.NArg() < 2 {
			fmt.Println("Error: missing YAML configuration file")
			printUsage()
			os.Exit(1)
		}
		configFile := flag.Arg(1)
		handleValidate(configFile)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: yamlforge <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  serve <config.yaml>    Start development server")
	fmt.Println("  build <config.yaml>    Generate static files")
	fmt.Println("  validate <config.yaml> Validate configuration")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func handleServe(configFile string, port int, host string) {
	config, err := parser.ParseConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	if config.Server.Port != 0 {
		port = config.Server.Port
	}
	if config.Server.Host != "" {
		host = config.Server.Host
	}

	srv := server.New(config)

	fmt.Printf("Starting yamlforge server on %s:%d\n", host, port)
	fmt.Printf("Configuration: %s\n", configFile)

	if err := srv.Start(host, port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleBuild(configFile string) {
	fmt.Printf("Building static files from %s\n", configFile)
	fmt.Println("Build functionality not yet implemented")
}

func handleValidate(configFile string) {
	config, err := parser.ParseConfig(configFile)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration %s is valid\n", configFile)
	fmt.Printf("App: %s v%s\n", config.App.Name, config.App.Version)
	fmt.Printf("Models: %d\n", len(config.Models))
}
