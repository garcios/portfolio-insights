-- Rename 'username' column back to 'name' in customers.users table
ALTER TABLE customers.users RENAME COLUMN username TO name;
