package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/storage"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(pathToStorage string) (*Storage, error) {
	const op = "storage.Postres.NewStorage"
	db, err := sql.Open("pgx", pathToStorage)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte, uuid string) (int64, error) {
	const op = "storage.postrges.SaveUser"
	query := `INSERT INTO users(email , passHash, confirmation_token) VALUES ($1 , $2 , $3) RETURNING id`

	var id int64

	err := s.db.QueryRowContext(ctx, query, email, passHash, uuid).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)

	}
	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.Postgres.User"
	query := `SELECT id , email, passHash, is_confirmed, confirmation_token FROM users WHERE email = $1`

	var usr models.User

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&usr.ID,
		&usr.Email,
		&usr.PassHash,
		&usr.IsConfirmed,
		&usr.ConfirmationToken,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return usr, nil
}

func (s *Storage) App(ctx context.Context, appId int) (models.App, error) {
	const op = "storage.Postgres.App"

	query := `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App

	err := s.db.QueryRowContext(ctx, query, appId).Scan(
		&app.ID,
		&app.Name,
		&app.Secret,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) Confirm(ctx context.Context, email string) error {
	const op = "storage.Postgres.Confirm"

	query := `UPDATE users SET is_confirmed = true WHERE email = $1`

	res, err := s.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return storage.ErrActivateUser
	}

	return nil

}

func (s *Storage) Create(ctx context.Context, usrId int64, body string) (int64, error) {
	const op = "storage.Postgres.Create"

	query := `INSERT INTO articles(author_id, article_text) VALUES ($1 , $2) RETURNING ID`

	var id int64
	err := s.db.QueryRowContext(ctx, query, usrId, body).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrCreateArticle)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetById(ctx context.Context, id int64) (models.Article, error) {
	const op = "storage.Postgres.getById"

	query := `SELECT id, author_id, article_text FROM articles WHERE id = $1`

	var article models.Article
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&article.Id,
		&article.AuthorId,
		&article.Body)
	if err != nil {
		return models.Article{}, fmt.Errorf("%s: %w", op, err)
	}

	return article, nil
}
func (s *Storage) GetByAuthor(ctx context.Context, authorId int64) ([]models.Article, error) {
	const op = "storage.Postgres.getByAuthor"

	query := `SELECT id , author_id, article_text FROM articles WHERE author_id = $1`

	var articles []models.Article
	res, err := s.db.QueryContext(ctx, query, authorId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer res.Close()

	for res.Next() {
		var article models.Article

		err := res.Scan(&article.Id, &article.AuthorId, &article.Body)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		articles = append(articles, article)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return articles, nil
}

func (s *Storage) GetFeed(ctx context.Context) ([]models.Article, error) {
	const op = "storage.Postgres.getFeed"

	query := `SELECT id , author_id , article_text FROM articles LIMIT 5`

	var articles []models.Article
	res, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer res.Close()

	for res.Next() {
		var acticle models.Article

		err := res.Scan(
			&acticle.Id,
			&acticle.AuthorId,
			&acticle.Body,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		articles = append(articles, acticle)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return articles, nil
}

func (s *Storage) GetArticleAuthor(ctx context.Context, id int64) (int64, error) {
	const op = "storage.Postgres.GetAuthor"

	query := `SELECT author_id FROM articles WHERE id = $1`
	var authorId int64
	err := s.db.QueryRowContext(ctx, query, id).Scan(&authorId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, storage.ErrUserNotFound
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return authorId, nil
}

func (s *Storage) Init(ctx context.Context, name string, secret string) (int64, error) {
	const op = "storage.Postgres.Init"

	query := `INSERT INTO apps(name, secret) VALUES($1 , $2) RETURNING ID`

	var appId int64
	err := s.db.QueryRowContext(ctx, query, name, secret).Scan(&appId)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			appId, err = s.GetAppId(ctx, name)
			if err != nil {
				return 0, fmt.Errorf("%s: %w", op, err)
			}
			return appId, nil
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return appId, nil

}

func (s *Storage) GetAppId(ctx context.Context, name string) (int64, error) {
	const op = "storage.POstgres.getAppId"

	query := `SELECT id FROM apps WHERE name = $1`

	var appId int64
	err := s.db.QueryRowContext(ctx, query, name).Scan(&appId)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return appId, nil
}

func (s *Storage) Edit(ctx context.Context, articleId int64, newBody string) error {
	const op = "storage.Postgres.Edit"

	query := `UPDATE articles SET article_text = $1 WHERE id = $2`

	res, err := s.db.ExecContext(ctx, query, newBody, articleId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return storage.ErrEditArticle
	}

	return nil

}

func (s *Storage) Stop() error {
	const op = "storage.Postgres.Stop"

	if s.db == nil {
		return nil
	}

	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
