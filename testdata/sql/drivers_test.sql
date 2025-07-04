-- Test file for @drivers command
SELECT 'Testing @drivers command' as message;
go

-- List all available drivers
@drivers
go

-- List drivers starting with 's' (should show sqlite3 and sqlserver)
@drivers "s"
go

-- List drivers starting with 'p' (should show postgres)
@drivers "p"
go

SELECT 'Driver listing completed' as final_message;
go
