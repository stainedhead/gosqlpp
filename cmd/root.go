package cmd

import (
	"fmt"
	"os"
	"time"

	"gosqlpp/internal/config"
	"gosqlpp/internal/database"
	"gosqlpp/internal/file"
	"gosqlpp/internal/output"
	"gosqlpp/internal/schema"

	"github.com/spf13/cobra"
)

// Version information - should be set at build time
var (
	Version   = "1.0.0"
	BuildDate = ""
	GitCommit = ""
)

var (
	// Global flags
	connectionName  string
	outputFormat    string
	inputFile       string
	inputDirectory  string
	newerThan       string
	forceExecution  bool
	useStdin        bool
	listConnections bool
	showVersion     bool

	// Global config
	cfg *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sqlpp [file]",
	Short: "SQL preprocessor and executor with multi-database support",
	Long: `sqlpp is a SQL preprocessor and executor that supports multiple database backends.
It provides #include, #define, and conditional preprocessing capabilities,
along with schema introspection and multiple output formats.

Schema Commands:
  @drivers                            # List all available database drivers
  @schema-tables [filter]             # List database tables
  @schema-views [filter]              # List database views  
  @schema-procedures [filter]         # List stored procedures
  @schema-functions [filter]          # List functions
  @schema-all [filter]                # Show all schema information

Examples:
  sqlpp script.sql                    # Execute a single SQL file
  sqlpp -c mydb script.sql            # Use specific connection
  sqlpp -d /path/to/scripts           # Process all .sql files in directory
  sqlpp -o json script.sql            # Output results as JSON
  sqlpp --newer "2023-01-01" -d .     # Process files newer than date
  sqlpp --stdin                       # Read SQL from standard input
  sqlpp -                             # Read SQL from standard input (alternative)
  echo "SELECT 1;  " | sqlpp --stdin  # Pipe SQL commands
  cat script.sql | sqlpp -            # Pipe file content`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSqlpp,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&connectionName, "connection", "c", "",
		"database connection name from config")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "",
		"output format (table, json, yaml, csv)")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "file", "f", "",
		"SQL file to process")
	rootCmd.PersistentFlags().StringVarP(&inputDirectory, "directory", "d", "",
		"directory containing SQL files to process")
	rootCmd.PersistentFlags().StringVarP(&newerThan, "newer", "n", "",
		"process only files newer than this date/time (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)")
	rootCmd.PersistentFlags().BoolVar(&forceExecution, "force", false,
		"continue execution even on errors (overrides end-on-error config)")
	rootCmd.PersistentFlags().BoolVar(&useStdin, "stdin", false,
		"read SQL commands from standard input")
	rootCmd.PersistentFlags().BoolVarP(&listConnections, "list-connections", "l", false,
		"list available database connections and exit")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false,
		"show version information and exit")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}
}

// runSqlpp is the main execution function
func runSqlpp(cmd *cobra.Command, args []string) error {
	// Show version and exit if requested
	if showVersion {
		fmt.Printf("sqlpp version %s\n", Version)
		if GitCommit != "" {
			fmt.Printf("  Build commit: %s\n", GitCommit)
		}
		if BuildDate != "" {
			fmt.Printf("  Build date:   %s\n", BuildDate)
		}
		return nil
	}

	// List connections and exit if requested
	if listConnections {
		connections := cfg.GetConnectionNames()
		if len(connections) == 0 {
			fmt.Println("No connections configured.")
			return nil
		}

		fmt.Println("Available connections:")
		for _, connName := range connections {
			prefix := "  - "
			if connName == cfg.DefaultConnection {
				prefix = "  * "
			}
			fmt.Printf("%s%s", prefix, connName)
			if connName == cfg.DefaultConnection {
				fmt.Printf(" (default)")
			}
			fmt.Println()
		}
		return nil
	}

	// Determine input source
	var inputSource string
	var isStdinInput bool

	if len(args) > 0 {
		inputSource = args[0]
		// Check if user specified "-" as file argument (stdin)
		if inputSource == "-" {
			isStdinInput = true
		}
	} else if inputFile != "" {
		inputSource = inputFile
		// Check if user specified "-" as file flag (stdin)
		if inputSource == "-" {
			isStdinInput = true
		}
	} else if useStdin {
		isStdinInput = true
	} else if inputDirectory == "" {
		return fmt.Errorf("no input specified: provide a file as argument, use --file, use --directory, or use --stdin")
	}

	// Validate mutually exclusive options
	if inputDirectory != "" && (inputSource != "" || inputFile != "" || isStdinInput || useStdin) {
		return fmt.Errorf("cannot specify --directory with file input or stdin")
	}

	if isStdinInput && inputDirectory != "" {
		return fmt.Errorf("cannot use stdin with --directory")
	}

	if newerThan != "" && inputDirectory == "" {
		return fmt.Errorf("--newer flag can only be used with --directory")
	}

	// Parse newer-than date if provided
	var newerThanTime time.Time
	if newerThan != "" {
		var err error
		// Try parsing as date first (YYYY-MM-DD)
		newerThanTime, err = time.Parse("2006-01-02", newerThan)
		if err != nil {
			// Try parsing as datetime (YYYY-MM-DD HH:MM:SS)
			newerThanTime, err = time.Parse("2006-01-02 15:04:05", newerThan)
			if err != nil {
				return fmt.Errorf("invalid date format for --newer: %s (use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", newerThan)
			}
		}
	}

	// Override config with command line flags
	effectiveConfig := *cfg
	if connectionName != "" {
		effectiveConfig.DefaultConnection = connectionName
	}
	if outputFormat != "" {
		effectiveConfig.Output = outputFormat
	}
	if forceExecution {
		effectiveConfig.EndOnError = false
	}

	// Validate effective configuration
	if err := effectiveConfig.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create database manager and establish connection
	dbManager := database.NewManager()
	defer dbManager.CloseAll()

	// Get connection configuration
	connConfig, err := effectiveConfig.GetConnection(effectiveConfig.DefaultConnection)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	// Connect to database
	if err := dbManager.Connect(effectiveConfig.DefaultConnection, connConfig); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get database connection
	conn, err := dbManager.GetConnection(effectiveConfig.DefaultConnection)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create executor and formatter
	executor := database.NewExecutor(conn)
	formatter := output.NewFormatter(effectiveConfig.Output, os.Stdout)
	introspector := schema.NewIntrospector(conn, formatter)

	// Create file processor
	processor := file.NewProcessor(executor, formatter, introspector, effectiveConfig.EndOnError)

	// Process files
	if inputDirectory != "" {
		fmt.Printf("Processing directory: %s\n", inputDirectory)
		if !newerThanTime.IsZero() {
			fmt.Printf("Files newer than: %s\n", newerThanTime.Format("2006-01-02 15:04:05"))
		}
		return processor.ProcessDirectory(inputDirectory, newerThanTime)
	} else if isStdinInput {
		fmt.Printf("Processing input from stdin\n")
		return processor.ProcessStdin()
	} else {
		fmt.Printf("Processing file: %s\n", inputSource)
		return processor.ProcessFile(inputSource)
	}
}

// GetConfig returns the loaded configuration (for use by other packages)
func GetConfig() *config.Config {
	return cfg
}
