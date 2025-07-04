-- Comprehensive sqlpp demonstration
-- This file showcases all major features

-- Phase D: Preprocessor definitions
#define APP_NAME "sqlpp Demo"
#define VERSION "1.0"
#define MAX_RECORDS 5
#define DEBUG_MODE 1

SELECT APP_NAME as application, VERSION as version;
go

-- Phase E: Conditional preprocessing
#ifdef DEBUG_MODE
SELECT 'Debug mode is active' as debug_status;
go

#ifndef PRODUCTION
SELECT 'Running in development environment' as env_status;
go
#end

#end

-- Phase D: Include functionality
#include "included_query.sqi"

-- Regular SQL with variable substitution
SELECT * FROM users LIMIT MAX_RECORDS;
go

-- Phase F: Schema introspection
SELECT 'Schema Information:' as section;
go

@schema-tables
go

@schema-views
go

-- Final message
SELECT 'Demo completed successfully!' as final_message;
go
