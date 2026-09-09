package main

import (
	"fmt"
	"os"

	"github.com/italiatroller-1990/garterscopic/internal/build"
	"github.com/italiatroller-1990/garterscopic/internal/config"
)

// runCheck validates the project without producing the final site.
//
//	garterscopic check [--verbose]
//
// Exit status is non-zero when errors are found (warnings alone do not fail
// unless links.fail_on_broken promotes them, which validation handles).
func runCheck(args []string) {
	verbose := false
	for _, a := range args {
		if a == "--verbose" {
			verbose = true
		}
	}

	fmt.Println("Garterscopic check")
	fmt.Println()

	cfg, err := config.Load("site.yaml")
	if err != nil {
		fmt.Println("✗ Configuration")
		printError("failed to load configuration", err)
		os.Exit(1)
	}

	builder := build.New(cfg)
	if err := builder.Load(); err != nil {
		fmt.Println("✗ Site structure")
		printError("failed to load site", err)
		os.Exit(1)
	}

	buildErr := builder.Validate()

	fmt.Println("✓ Configuration")
	fmt.Println("✓ Pages")
	fmt.Println("✓ Components")
	fmt.Println("✓ Styles")
	fmt.Println("✓ Assets")

	if buildErr == nil {
		fmt.Println("\nNo problems found.")
		return
	}

	for _, e := range buildErr.Warnings {
		fmt.Printf("warning: %s\n", e.Message)
		if verbose {
			fmt.Printf("  File: %s\n  Fix:  %s\n", e.Path, e.Hint)
		}
	}

	if buildErr.HasErrors() {
		fmt.Printf("\n✗ %d error(s)\n", len(buildErr.Errors))
		for _, e := range buildErr.Errors {
			fmt.Printf("\nerror: %s\n", e.Message)
			if verbose {
				fmt.Printf("  File: %s\n  Fix:  %s\n", e.Path, e.Hint)
			}
		}
		os.Exit(1)
	}

	fmt.Printf("\n⚠ %d warning(s), no errors\n", len(buildErr.Warnings))
}

// runGraph prints the resolved page -> layout/component/style structure.
//
//	garterscopic graph [--json]
func runGraph(args []string) {
	jsonOut := false
	for _, a := range args {
		if a == "--json" {
			jsonOut = true
		}
	}

	cfg, err := config.Load("site.yaml")
	if err != nil {
		printError("failed to load configuration", err)
		os.Exit(1)
	}

	builder := build.New(cfg)
	if err := builder.Load(); err != nil {
		printError("failed to load site", err)
		os.Exit(1)
	}

	siteGraph := builder.SiteGraph()

	if jsonOut {
		out, err := siteGraph.RenderJSON()
		if err != nil {
			printError("failed to render graph", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	fmt.Println(siteGraph.RenderText())
}
