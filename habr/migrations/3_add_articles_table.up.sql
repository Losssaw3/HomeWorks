-- Active: 1778328872255@@127.0.0.1@5432@sso

CREATE TABLE IF NOT EXISTS articles
(   id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id INT NOT NULL,
    article_text TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_articles_author_id ON articles(author_id);