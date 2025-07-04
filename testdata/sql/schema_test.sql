-- Test file for schema introspection
SELECT 'Testing schema introspection' as message;
go

-- List all tables
@schema-tables
go

-- List tables starting with 'u'
@schema-tables "u"
go

-- List all views (should be empty for SQLite)
@schema-views
go

-- List all stored procedures (not supported by SQLite)
@schema-procedures
go

-- List all functions (not supported by SQLite)
@schema-functions
go

-- Show all schema information
@schema-all
go
