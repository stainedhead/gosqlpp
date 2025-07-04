# sqlpp - SQL Preprocessor and Executor

A modern, feature-rich SQL preprocessor and executor written in Go that supports multiple database backends, advanced preprocessing capabilities, and schema introspection.

## Features

### 🗄️ Multi-Database Support
- **PostgreSQL** - Full support with advanced features
- **MySQL/MariaDB** - Complete compatibility
- **SQLite** - Embedded database support
- **SQL Server** - Microsoft SQL Server integration
- **CockroachDB** - Distributed SQL database support

### 🔧 Advanced Preprocessing
- **Variable Definitions** - `#define` directives for constants and strings
- **File Inclusion** - `#include` support with proper line tracking
- **Conditional Compilation** - `#ifdef`, `#ifndef`, `#end` blocks
- **Variable Substitution** - Intelligent whole-word replacement
- **Comment Handling** - C-style comments in preprocessor directives

### 📊 Schema Introspection
- **@schema-tables** - List database tables with column information
- **@schema-views** - Display database views and metadata
- **@schema-procedures** - Show stored procedures (database-dependent)
- **@schema-functions** - List database functions (database-dependent)
- **@schema-all** - Comprehensive schema overview
- **@drivers** - List all available database drivers

### 🎯 Flexible Execution
- **Multiple Output Formats** - Table, JSON, YAML, CSV
- **Batch Processing** - Process entire directories of SQL files
- **Date Filtering** - Process only files newer than specified date
- **Progress Tracking** - Real-time progress bars for long operations
- **Error Handling** - Configurable stop-on-error behavior
- **Standard Input** - Support for piped input and interactive use

## Installation

### From Source
```bash
git clone <repository-url>
cd sqlpp
go build -o sqlpp
```

### Binary Installation
```bash
# Move the built binary to your PATH
sudo mv sqlpp /usr/local/bin/
```

## Quick Start

### 1. Configuration
Create a `.sqlppconfig` file in your project directory:

```yaml
default-connection: "main"
end-on-error: false
output: "table"

connections:
  main:
    driver: "sqlite3"
    connection-string: "test.db"
  postgres:
    driver: "postgres"
    connection-string: "postgres://user:password@localhost/dbname?sslmode=disable"
  mysql:
    driver: "mysql"
    connection-string: "username:password@tcp(127.0.0.1:3306)/database_name"
  sqlserver:
    driver: "sqlserver"
    connection-string: "server=localhost;user id=userdb;password=userpwd;port=1433"
```

### 2. Basic Usage
```bash
# Execute a single SQL file
sqlpp script.sql

# Use specific database connection
sqlpp -c postgres script.sql

# Process all SQL files in a directory
sqlpp -d /path/to/sql/scripts

# Output results as JSON
sqlpp -o json script.sql

# Read from standard input
echo "SELECT 1; go" | sqlpp --stdin
cat script.sql | sqlpp -
```

## Command Line Options

```
Usage: sqlpp [file]

Flags:
  -c, --connection string    Database connection name from config
  -d, --directory string     Directory containing SQL files to process
  -f, --file string         SQL file to process
      --force               Continue execution even on errors
  -h, --help                Help for sqlpp
  -n, --newer string        Process only files newer than date (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)
  -o, --output string       Output format (table, json, yaml, csv)
      --stdin               Read SQL commands from standard input
```

## Preprocessing Directives

### Variable Definitions
```sql
-- Define constants and strings
#define APP_NAME "My Application"
#define VERSION "1.0.0"
#define MAX_RECORDS 100

SELECT APP_NAME as name, VERSION as version;
SELECT * FROM users LIMIT MAX_RECORDS;
go
```

### File Inclusion
```sql
-- Include common definitions
#include "common_defines.sql"

-- Include reusable queries
#include "user_queries.sqi"

SELECT * FROM users WHERE status = ACTIVE_STATUS;
go
```

### Conditional Compilation
```sql
#define DEBUG_MODE 1

#ifdef DEBUG_MODE
SELECT 'Debug mode active' as debug_info;
go

#ifndef PRODUCTION
SELECT 'Development environment' as env_info;
go
#end

#end
```

## Schema Introspection

### List Database Objects
```sql
-- Show all available drivers
@drivers
go

-- List all tables
@schema-tables
go

-- List tables matching pattern
@schema-tables user%
go

-- Show all views
@schema-views
go

-- Display stored procedures
@schema-procedures
go

-- Show all schema information
@schema-all
go
```

## Output Formats

### Table Format (Default)
```
+----+----------+-------------------+
| id | username | email            |
+----+----------+-------------------+
|  1 | john     | john@example.com |
|  2 | jane     | jane@example.com |
+----+----------+-------------------+
```

### JSON Format
```bash
sqlpp -o json script.sql
```
```json
[
  {"id": 1, "username": "john", "email": "john@example.com"},
  {"id": 2, "username": "jane", "email": "jane@example.com"}
]
```

### YAML Format
```bash
sqlpp -o yaml script.sql
```
```yaml
- id: 1
  username: john
  email: john@example.com
- id: 2
  username: jane
  email: jane@example.com
```

### CSV Format
```bash
sqlpp -o csv script.sql
```
```csv
id,username,email
1,john,john@example.com
2,jane,jane@example.com
```

## Advanced Usage

### Batch Processing with Date Filtering
```bash
# Process only files modified after January 1, 2024
sqlpp -d ./sql-scripts --newer "2024-01-01"

# Process files with specific datetime
sqlpp -d ./migrations --newer "2024-01-01 10:30:00"
```

### Pipeline Integration
```bash
# Generate SQL dynamically and execute
echo "SELECT COUNT(*) FROM users; go" | sqlpp --stdin

# Process template files
envsubst < template.sql | sqlpp -

# Chain with other tools
cat migration_*.sql | sqlpp - | tee results.json
```

### Error Handling
```bash
# Continue on errors (override config)
sqlpp --force problematic_script.sql

# Stop on first error (default behavior can be configured)
sqlpp script.sql
```

## Project Structure

```
sqlpp/
├── cmd/                    # Command-line interface
│   └── root.go            # Main CLI command and flags
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and execution
│   ├── file/            # File processing and batch operations
│   ├── output/          # Output formatting (table, JSON, YAML, CSV)
│   ├── preprocessor/    # SQL preprocessing engine
│   └── schema/          # Database schema introspection
├── testdata/             # Test files and examples
│   ├── config/          # Test configurations
│   ├── includes/        # Include file examples
│   └── sql/            # SQL test files
├── prompts/             # Development prompts and documentation
├── .sqlppconfig         # Default configuration file
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── main.go             # Application entry point
```

## Configuration Reference

### Connection Types

#### SQLite
```yaml
sqlite_conn:
  driver: "sqlite3"
  connection-string: "path/to/database.db"
```

#### PostgreSQL
```yaml
postgres_conn:
  driver: "postgres"
  connection-string: "postgres://username:password@hostname:port/database?sslmode=disable"
```

#### MySQL/MariaDB
```yaml
mysql_conn:
  driver: "mysql"
  connection-string: "username:password@tcp(hostname:port)/database"
```

#### SQL Server
```yaml
sqlserver_conn:
  driver: "sqlserver"
  connection-string: "server=hostname;user id=username;password=password;port=1433;database=dbname"
```

### Global Settings
```yaml
default-connection: "main"    # Default connection to use
end-on-error: false          # Stop processing on first error
output: "table"              # Default output format
```

## Examples

### Complete Preprocessing Example
```sql
-- File: demo.sql
#define APP_NAME "User Management System"
#define VERSION "2.1.0"
#define DEBUG_MODE 1

-- Application info
SELECT APP_NAME as application, VERSION as version;
go

#ifdef DEBUG_MODE
-- Debug information
SELECT 'Debug mode enabled' as debug_status;
go
#end

-- Include common queries
#include "user_queries.sqi"

-- Schema information
@schema-tables user%
go

-- Final query with variable substitution
SELECT * FROM users WHERE active = 1 LIMIT 10;
go
```

### Batch Processing Example
```bash
# Process all SQL files in migrations directory
sqlpp -d ./migrations

# Process only recent migration files
sqlpp -d ./migrations --newer "2024-06-01"

# Process with specific output format
sqlpp -d ./reports -o json > results.json
```

## Error Handling and Debugging

### Error Messages
The tool provides detailed error messages with file and line information:
```
Error in file 'script.sql' at line 15: syntax error near 'SELCT'
Error in file 'included.sqi' at line 3 (included from 'main.sql' line 8): undefined variable 'MISSING_VAR'
```

### Debugging Tips
1. Use `#ifdef DEBUG_MODE` blocks for conditional debug output
2. Enable table output format to see data structure clearly
3. Use `@schema-tables` to verify table structure before queries
4. Test preprocessing with simple files before complex scripts

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development

### Building from Source
```bash
go mod download
go build -o sqlpp
```

### Running Tests
```bash
go test ./...
```

### Testing with Examples
```bash
# Test basic functionality
./sqlpp testdata/sql/test.sql

# Test preprocessing
./sqlpp testdata/sql/preprocessor_test.sql

# Test conditional compilation
./sqlpp testdata/sql/conditional_test.sql

# Test schema introspection
./sqlpp testdata/sql/schema_test.sql
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the original C++ sqlpp utility
- Built with the Go ecosystem and modern database drivers
- Uses Cobra for CLI framework and various Go libraries for database connectivity
