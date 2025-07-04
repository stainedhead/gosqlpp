-- Second test file
SELECT 'Processing second file' as status;
go

UPDATE users SET email = 'updated@example.com' WHERE name = 'Bob Wilson';
go

SELECT name, email FROM users WHERE email LIKE '%updated%';
go
