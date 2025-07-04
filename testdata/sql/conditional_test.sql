-- Test file with conditional preprocessing
#define DEBUG 1
#define TABLE_NAME users

SELECT 'Starting conditional test' as message;
go

#ifdef DEBUG
SELECT 'Debug mode is enabled' as debug_info;
go
#end

#ifndef PRODUCTION
SELECT 'Not in production mode' as env_info;
go
#end

#ifdef UNDEFINED_VAR
SELECT 'This should not appear' as hidden_message;
go
#end

-- Nested conditionals
#ifdef DEBUG
SELECT 'In debug block' as status;
go

#ifndef PRODUCTION
SELECT 'Debug and not production' as nested_status;
go
#end

SELECT 'Still in debug block' as status;
go
#end

SELECT 'Final message' as message;
go
