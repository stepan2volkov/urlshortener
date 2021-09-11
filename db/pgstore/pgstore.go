package pgstore

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // PostgreSQL Driver

	"github.com/stepan2volkov/urlshortener/app"
)

var _ app.URLStore = &PgStore{}

type PgURL struct {
	ID           int       `db:"id"`
	CreatedAt    time.Time `db:"created_at"`
	OriginalURL  string    `db:"original_url"`
	ShortURL     string    `db:"short_url"`
	NumRedirects int       `db:"num_redirects"`
}

type PgStats struct {
	ShortURL     string `db:"short_url"`
	NumRedirects int    `db:"num_redirects"`
}

type PgStore struct {
	db *sql.DB
}

func NewPgStore(dsn string) (*PgStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	ps := &PgStore{db: db}
	if err = ps.migrate(); err != nil {
		return nil, err
	}
	return ps, nil
}

func (s *PgStore) migrate() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		id            bigserial primary key,
		created_at    timestamp with time zone,
		original_url  varchar,
		short_url     varchar,
		num_redirects bigint default 0
	);`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS urls_short_url_uindex ON urls (short_url);`)
	return err
}

func (s *PgStore) Close() error {
	return s.db.Close()
}

func (s *PgStore) Create(ctx context.Context, originalURL string) (*app.URL, error) {
	pgURL := &PgURL{
		CreatedAt:   time.Now(),
		OriginalURL: originalURL,
	}

	row := s.db.QueryRowContext(ctx, `INSERT INTO urls (created_at, original_url) VALUES ($1, $2) RETURNING id`, pgURL.CreatedAt, pgURL.OriginalURL)

	var id int
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return &app.URL{
		ID:          id,
		OriginalURL: pgURL.OriginalURL,
	}, nil
}

func (s *PgStore) UpdateURL(ctx context.Context, url *app.URL) error {
	pgURL := &PgURL{
		ID:       url.ID,
		ShortURL: url.ShortURL,
	}
	_, err := s.db.ExecContext(ctx, "UPDATE urls SET short_url = $1 WHERE id = $2", pgURL.ShortURL, pgURL.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PgStore) GetOriginalURL(ctx context.Context, shortURL string) (*app.URL, error) {
	pgURL := &PgURL{}

	row := s.db.QueryRowContext(ctx, `SELECT id, created_at, original_url, short_url, num_redirects
		FROM urls WHERE short_url = $1`, shortURL)
	err := row.Scan(&pgURL.ID, &pgURL.CreatedAt, &pgURL.OriginalURL, &pgURL.ShortURL, &pgURL.NumRedirects)
	if err != nil {
		return nil, err
	}
	return &app.URL{
		ID:           pgURL.ID,
		OriginalURL:  pgURL.OriginalURL,
		ShortURL:     pgURL.ShortURL,
		NumRedirects: pgURL.NumRedirects,
	}, nil
}

func (s *PgStore) GetStats(ctx context.Context, shortURL string) (*app.Stats, error) {
	stats := &PgStats{}

	row := s.db.QueryRowContext(ctx, `SELECT short_url, num_redirects FROM urls WHERE short_url = $1`, shortURL)
	err := row.Scan(&stats.ShortURL, &stats.NumRedirects)
	if err != nil {
		return nil, err
	}
	return &app.Stats{
		ShortURL:     stats.ShortURL,
		NumRedirects: stats.NumRedirects,
	}, nil
}

func (s *PgStore) IncreaseNumRedirects(ctx context.Context, shortURL string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE urls SET num_redirects = num_redirects + 1 WHERE short_url = $1", shortURL)
	if err != nil {
		return err
	}
	return nil
}
