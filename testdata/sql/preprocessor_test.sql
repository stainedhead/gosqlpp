-- Test file with preprocessor directives
#define TABLE_NAME users
#define LIMIT_COUNT 5

SELECT 'Testing preprocessor' as message;
go

SELECT * FROM TABLE_NAME LIMIT LIMIT_COUNT;
go

#include "included_query.sqi"

SELECT COUNT(*) as total_TABLE_NAME FROM TABLE_NAME;
go
