package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	DefaultConsulAddr = "http://127.0.0.1:8500"
	DefaultDatacenter = "dc1"
)

// Injected at build time by goreleaser ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Define command line flags
	var (
		kvDir       = flag.String("dir", "", "Path to the directory of YAML files to sync (required)")
		dryRun      = flag.Bool("dry-run", false, "Perform a dry run without making actual changes")
		export      = flag.Bool("export", false, "Export KV pairs in Consul JSON format to stdout")
		consulAddr  = flag.String("consul-addr", DefaultConsulAddr, "Consul HTTP API address")
		datacenter  = flag.String("datacenter", DefaultDatacenter, "Consul datacenter")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		showVersion = flag.Bool("version", false, "Show version information")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -dir <directory> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "consul-kv-sync synchronizes YAML files to Consul KV store.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("consul-kv-sync version %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	// Validate required flags
	if *kvDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -dir flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Set up logging
	if !*verbose {
		log.SetOutput(io.Discard)
	}

	// Execute main logic
	if err := run(*kvDir, *dryRun, *export, *consulAddr, *datacenter, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(kvDir string, dryRun, export bool, consulAddr, datacenter string, verbose bool) error {
	// Load configuration and files
	kvMaps, filenames, err := loadConfigurationAndFiles(kvDir, verbose)
	if err != nil {
		return err
	}

	// Check for duplicates
	if err := checkDuplicates(kvMaps, filenames, verbose); err != nil {
		return err
	}

	// Collect and process KV pairs
	allPairs := collectAllKVPairs(kvMaps)
	if verbose {
		fmt.Printf("Collected %d key-value pairs\n", len(allPairs))
	}

	// Handle export mode
	if export {
		return exportToJSON(allPairs)
	}

	// Handle dry run
	if dryRun {
		fmt.Println("\n[DRY RUN MODE] No changes will be made to Consul")
		fmt.Println(formatKVPairsForDisplay(allPairs))
		return nil
	}

	// Sync to Consul
	return syncToConsul(allPairs, consulAddr, datacenter, verbose)
}

func loadConfigurationAndFiles(kvDir string, verbose bool) ([]map[string]interface{}, []string, error) {
	// Step 1: Collect the YAML files in the directory
	paths, err := yamlFiles(kvDir)
	if err != nil {
		return nil, nil, err
	}

	if verbose {
		fmt.Printf("Found %d files in %s\n", len(paths), kvDir)
	}

	// Step 2: Load all YAML files
	if verbose {
		fmt.Println("Loading YAML files...")
	}

	kvMaps, filenames, err := loadAllYAMLFiles(paths)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load YAML files: %w", err)
	}

	return kvMaps, filenames, nil
}

func checkDuplicates(kvMaps []map[string]interface{}, filenames []string, verbose bool) error {
	if verbose {
		fmt.Println("Checking for duplicate keys...")
	}

	duplicates, err := detectDuplicates(kvMaps, filenames)
	if err != nil {
		return fmt.Errorf("failed to detect duplicates: %w", err)
	}

	if len(duplicates) > 0 {
		fmt.Fprintln(os.Stderr, formatDuplicateError(duplicates))
		return fmt.Errorf("duplicate keys detected")
	}

	return nil
}

func syncToConsul(allPairs []KVPair, consulAddr, datacenter string, verbose bool) error {
	fmt.Printf("Syncing %d key-value pairs to Consul KV store...\n", len(allPairs))

	client := NewConsulClient(consulAddr, datacenter)
	summary, err := client.syncKVPairs(allPairs, verbose)

	// Always display summary if available
	if summary != nil {
		fmt.Println(formatExecutionSummary(summary))
	}

	if err != nil && summary == nil {
		return fmt.Errorf("failed to sync KV pairs: %w", err)
	}

	if summary != nil && summary.FailedBatches > 0 {
		return fmt.Errorf("%d out of %d batches failed", summary.FailedBatches, summary.TotalBatches)
	}

	return nil
}
