-- Test @schema-all with new @drivers command
SELECT 'Testing enhanced @schema-all command' as message;
go

@schema-all
go

SELECT 'Schema-all completed' as final_message;
go
