-- Comprehensive stdin demonstration file
-- This file can be piped through stdin to test all functionality

SELECT 'Starting stdin demonstration' as message;
go

-- Test basic SQL
SELECT 1 as number, 'Hello World' as greeting;
go

-- Test schema introspection
@drivers
go

-- Test filtered drivers
@drivers "s"
go

-- Test table listing
@schema-tables
go

-- Test comprehensive schema
@schema-all
go

SELECT 'Stdin demonstration completed successfully' as final_message;
go
