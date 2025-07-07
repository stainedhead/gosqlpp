I am a software engineer.  I wrote a utility application years many years ago to help my team with database enginerring.  the application was called sqlpp.exe and was written in C++ to help development on Sybase and SQL Server.  I made the source open source years ago and dropped the project over time.  I have wanted to bring it back to life, and have decided to take it a couple steps further.

The earlier version supported #include and #define functionality, allowing the use of standardized templates to be dynamically included into SQL scripts and customized by the replacement of #define driven values.

This application will support multiple database backends, and will try to support shared features such as schema lookup and @functions that drive short-hand execution of SQL execution.

the name of the application will be sqlpp, and the configuration file for the application will be .sqlppconfig that will be kept in the same directory as the sqlpp executable when the application is run.

#### Added documentation
The following URLs are documentation that will be helpful in your processing.  Please review this documentation, and the files these may link to, and use them as context in your reasoning and planning.
https://go.dev/doc/tutorial/database-access
https://pkg.go.dev/database/sql
https://pkg.go.dev/database/sql/driver
http://go-database-sql.org/
https://github.com/jimsmart/schema
https://blog.jetbrains.com/go/2023/02/28/getting-started-with-the-database-sql-package/
https://cplusplus.com/doc/tutorial/preprocessor/
https://www.cs.unm.edu/~storm/C++/PreProcessor.html



When working with databases in Go within an enterprise setting using the database/sql package, you'll commonly encounter several popular relational database management systems (RDBMS). 
Main Enterprise Database Engines Supported by database/sql:

PostgreSQL: A powerful, open-source object-relational database system known for its robust features and extensibility.

MySQL/MariaDB: A widely-used open-source relational database system, known for its performance and scalability.

Microsoft SQL Server: A commercial relational database system developed by Microsoft, popular in Windows environments.


Recommended Drivers and Connection Strings:
Here are commonly used and well-regarded drivers for these databases with example connection strings:
Database 	Recommended Driver Package	Example Connection String
PostgreSQL:	github.com/lib/pq	postgres://user:password@host:port/dbname?sslmode=disable

MySQL/MariaDB:	github.com/go-sql-driver/mysql	username:password@tcp(127.0.0.1:3306)/database_name

Microsoft SQL Server:	github.com/denisenkom/go-mssqldb or github.com/microsoft/go-mssqldb	server=localhost;user id=userdb;password=userpwd;port=1433

SQLite:
Driver: github.com/mattn/go-sqlite3
Description: A self-contained, file-based database, ideal for simpler projects, testing, or embedded applications where portability and ease of deployment are priorities.

CockroachDB:
Driver: Can use any PostgreSQL driver, such as github.com/lib/pq.
Description: A distributed SQL database offering high availability and scalability, often used for mission-critical applications.


Important Notes:
Driver Import: When using database/sql, you need to import the specific driver package, usually using a blank identifier (_) so that the driver registers itself with the database/sql package.
Connection String Format: The connection string format is driver-specific. Always consult the documentation for your chosen driver to confirm the correct format and available options. You might also find helper functions within the driver packages to format connection strings.
Security: Be mindful of security when handling database credentials. Avoid hardcoding sensitive information in your code. Consider using environment variables or a secure configuration management system.
Error Handling: Always check for errors when opening database connections and performing queries.
Connection Pooling: database/sql handles connection pooling, which is crucial for performance and resource management in enterprise applications. 
These drivers and connection string examples should provide a solid starting point for connecting your Go application to common enterprise databases using the database/sql package. 



### Technical drivers 

This application will be a commandline app (CLI) which will be developed with golang
The core package used to connect to databases will be "database/sql"
The core package used for drivers is "database/sql/drivers"
The package to be used for Schema query is "https://github.com/jimsmart/schema"
The package to be used for table output is "github.com/rodaine/table"

#### Requirements

The following sets of requirements are provided as the requirememts which the completed code will be expected to support.  Each set of requirements is labeled by a sequence of letters.  These letters are also suggestions of the order of code generation and testing which could be helpful to sequentually build the application.

A. Configuration, Named-Connections, Commandline Flags 

the application will support YAML based configuration files which have enables storage of application defaults on processing.  settings such as default-connection, end-on-error and output.  
-- default-connection allows the user to select a connection to use if they have multiple, if this is not set and there is only one connetion defined it will use that connection when default-connection is not set.  
-- end-on-error controls if processing continues is stopped or continues when an error is returned from the database.  this setting is a boolean value which has a default value of false.
-- output will default to table unless the value is set.

there willl also be a connections section of configuration file, each connection will have a name, driver and connectstring.  the name allows the user to select which connection to use when there are multiple provided, driver and connectstring are the two values which will be passed to the Open method when creating a database connection.

There will also be a set of commandline flags that the user set to control processing or override configuration settings.

--connection or -c which allows the user to control which of the named connections stored in the configuration file will be used when the application runs.  The connection name is passed as the string value following this flag.

--output or -o which defines which of the output formats to be used when the application runs, which is passed as the string value following this flag.  the choices are table, json, yaml and csv with table being the default.

--file or -f which defines the file to be processed, which is passed as the string value following this flag.  this flag is not required if the user passes the filename as the first parameter to the application.  the flag is also not supported when processing --directory or -newer flags.

--directory or -d which defines a directory to be processed, which is passed as the string value following this flag.  when used the application will read all the files with the extension of .sql which are in the directory and process them in order found.

--newer or -n which limits a --directory processing to only those which are newer than the datetime passed as value which follows the flag.

--list-connections or -l returns the connections available in the configuration file. The default connection should be labelled as default if one is set.

--version or -v which shows the version of the application.

--force overrides the default behavior of errors stopping the processing of the application when errors are reported by the 


B. Start, Connect, Read File, Execute, Process-Results, Close
This application will support Postgreql, Mysql/MariaDb, sqlite, microsoft sql server, coachroachdb

When the application is run, the default processing flow would be to open the filename which is provided on the command line as the first value passed, unless that value is a commandline flag.  If a value is not passed in this first value, i could be passed via the --file flag.  Files can be broken into statements which are each executed individually as they are read.  The statement is deliminated by use of "go " which is found as starting text of any line.  lines starting with "go " are not sent to the database as they are not part of the statement.

as statements are read from the file, the line number of each line is tracked, and mapped to the line number of the statement.  This will allow errors reported when statements are executed by the database to be reported to the user using the line number in the file, rather than the line number in the statement.  This error processing reporting should follow the same format used in C compiler errors on line numbers, and will include the error string returned by the database for each error.  when errors are returned the stop-on-error configuration setting should be checked to stop processing when it is set to true.  when stop-on-error is false, errors are treated as end of statement processing, but the next statement can be processed if it is available.

statements which are executed in the the database may return information on the number of rows effected or return results.  rows effected will be returned as informational text, and results will be returned in the format requested by output configuration value.  the formats expected to be supported are table, json, yaml and CSV. 

the file may have one or more statements and each statement will be set to the database in the order they are found.  when the final statement is processed, the file will be closed, and processing of the file is completed.

C. Process directories or sets of files
When the --directory or --newer flags are provided the application is being asked to process a series of files.  --directory is used to provide the path to the directory to be processed.  when used the default behavior is to read all the .sql files and process them in order.

if errors are found within any file, the stop-on-error configuration value is used to control if other files are processed or not.  when stop-on-error is true, the processing is stopped, when false the next file is processed in order.

when the newer flag is provided on the commandline, the files processes within the directory are only those which have been updated since the value passed with the flag converted to a datetime value.  values which are passed without time will default to midnight of the date, times passed should be used as a localized value.  files with datetimes before the value are skipped from being processed.

D. #define, #include, included line tracking, included error reporting

if lines are found with a #define value, there are expeceted to be two additional values, a name and a value.  The name will be remembered and mapped to the value passed.  values can be strings within double quotes, or values without quotes which are stored as strings.  #define lines may also have C style comments that follow the values passed, meaning text that begins with // anything following the // is discarded and not validated or used by the appliation. the original text on the line with the #define is removed from the statement being processed.

some examples are;

#define TRUE Y
#define MAX_LOOPS 100
#define PROC_NAME "stored-proc-name"
#define FOOBAR    "some text to use"    // some text to ignore


if lines are found within a file when it is processed that begin with #include the value passed as a string will be read as a string which is added to the statement which is being processed.  all processing rules for files are used while injecting this content into the statement being processed, this includes the file line mapping logic, which for this context would be mapped to the file being included rather than the file doing the including if the included content is causing the error.  the original text on the line with the #include is removed from the statement being processed.

some examples:

#include "somefile.sqi"
#include "someotherfile.sqi"  // does something else


E. #ifdef, #ifndef, #end

Ensure we support #ifdef and #ifndef blocks, blocks begin with one of these tokens and ends with a #end token.  #ifdef causes processing to be done if a value named is contained in the #define values.  #ifndef causing processing within the block if the #define value is not within the #define values.  each block is completed by a #end, and any context within the block is only processed if the opening token is found to be true.  when the token is false the the block is skipped from the beginning of the statement to the end.  the lines with #ifdef, #ifndef and #end are not added to the statement being processed as they are only flow control and not part of the statement content.

F. @schema-all, @schema-tables, @schema-views, @schema-procedures, @schema-functions

when processing a file, we will also have @name, or shorthand terms, that can be used to query the schema of the database being used.  the results of these queries will be returned in the format defined in the output configuration setting.  users who provide only the @name will have all of the records of the type returned.  if they provide a value following the @name, it will be used as name filter to the processing.  meaning the value is provided as a filter and those records starting with that value are not returned.

some examples:

@schema-tables returns all tables
@schema-tables "foo", returns all of the tables with a nane starting with foo

@schema-all used as a short hand to run all the other @name values.  This is done in order of tables, views, procedures and functions in that order.

@schema-tables returns tables
@schema-views returns views
@schema-procedures returns stored procedures
@schema-functions returns functions

if any of these types are not supported by the database a message is returned saying they are not supported.  the data return should be the data supported by the database engine and driver.

