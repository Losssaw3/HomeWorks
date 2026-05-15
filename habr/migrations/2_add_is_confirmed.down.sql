-- Active: 1778328872255@@127.0.0.1@5432@sso
ALTER TABLE users DROP COLUMN is_confirmed;
ALTER TABLE users DROP COLUMN confirmation_token;