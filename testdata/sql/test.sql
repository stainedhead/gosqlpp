-- Test SQL file
SELECT 'Hello, World!' as message;
go

SELECT * FROM users;
go

INSERT INTO users (name, email) VALUES ('Bob Wilson', 'bob@example.com');
go

SELECT COUNT(*) as user_count FROM users;
go
