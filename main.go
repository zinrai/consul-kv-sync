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
	DefaultConfigFile = "./environments.yaml"
)

func main() {
	// Define command line flags
	var (
		environment = flag.String("env", "", "Environment name (required)")
		configFile  = flag.String("config", DefaultConfigFile, "Path to environments configuration file")
		dryRun      = flag.Bool("dry-run", false, "Perform a dry run without making actual changes")
		export      = flag.Bool("export", false, "Export KV pairs in Consul JSON format to stdout")
		consulAddr  = flag.String("consul-addr", DefaultConsulAddr, "Consul HTTP API address")
		datacenter  = flag.String("datacenter", DefaultDatacenter, "Consul datacenter")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -env <environment> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "consul-kv-syncer synchronizes YAML files to Consul KV store.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -env production\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -env staging -dry-run\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -env production -export > production-kv.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -env production -consul-addr http://consul:8500 -verbose\n", os.Args[0])
	}

	flag.Parse()

	// Validate required flags
	if *environment == "" {
		fmt.Fprintf(os.Stderr, "Error: -env flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Set up logging
	if !*verbose {
		log.SetOutput(io.Discard)
	}

	// Execute main logic
	if err := run(*environment, *configFile, *dryRun, *export, *consulAddr, *datacenter, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(environment, configFile string, dryRun, export bool, consulAddr, datacenter string, verbose bool) error {
	// Load configuration and files
	kvMaps, filenames, err := loadConfigurationAndFiles(environment, configFile, verbose)
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

func loadConfigurationAndFiles(environment, configFile string, verbose bool) ([]map[string]interface{}, []string, error) {
	// Step 1: Load environment configuration
	if verbose {
		fmt.Printf("Loading configuration from %s...\n", configFile)
	}

	config, err := loadEnvironments(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Step 2: Get files for the specified environment
	files, err := getEnvironmentFiles(config, environment)
	if err != nil {
		return nil, nil, err
	}

	if verbose {
		fmt.Printf("Found %d files for environment '%s'\n", len(files), environment)
	}

	// Step 3: Resolve file paths
	resolvedPaths := resolveFilePaths(configFile, files)

	// Step 4: Load all YAML files
	if verbose {
		fmt.Println("Loading YAML files...")
	}

	kvMaps, filenames, err := loadAllYAMLFiles(resolvedPaths)
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
