package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/stepan2volkov/urlshortener/app/base58"
)

type URL struct {
	ID           int
	OriginalURL  string
	ShortURL     string
	NumRedirects int
}

type Stats struct {
	ShortURL     string
	NumRedirects int
}

// URLStore is responsible for storing and getting url data.
type URLStore interface {
	Create(ctx context.Context, originalURL string) (*URL, error)
	UpdateURL(ctx context.Context, url *URL) error
	GetOriginalURL(ctx context.Context, shortURL string) (*URL, error)
	GetStats(ctx context.Context, shortURL string) (*Stats, error)
	IncreaseNumRedirects(ctx context.Context, shortURL string) error
}

type App struct {
	store URLStore
}

func NewApp(store URLStore) *App {
	return &App{
		store: store,
	}
}

// CreateURL generates short URL and saving it in the store.
func (a *App) CreateURL(ctx context.Context, originalURL string) (*URL, error) {
	url, err := a.store.Create(ctx, originalURL)
	if err != nil {
		return nil, fmt.Errorf("error when creating: %w", err)
	}
	shortURL, err := base58.Decode(url.ID)
	if err != nil {
		return nil, fmt.Errorf("error when generating short URL: %w", err)
	}
	url.ShortURL = shortURL

	if err = a.store.UpdateURL(ctx, url); err != nil {
		return nil, fmt.Errorf("error when saving URL in db: %w", err)
	}
	return url, nil
}

// GetRedirectURL searches short URL in the store and returns original URL to redirect
func (a *App) GetRedirectURL(ctx context.Context, shortURL string) (*URL, error) {
	url, err := a.store.GetOriginalURL(ctx, shortURL)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("URL not found")
		default:
			return nil, fmt.Errorf("error when getting url: %w\n", err)
		}
	}
	a.increaseNumRedirects(ctx, shortURL)

	return url, nil
}

// GetStats searches short URL in the store and returns redirecting stats
func (a *App) GetStats(ctx context.Context, shortURL string) (*Stats, error) {
	stats, err := a.store.GetStats(ctx, shortURL)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("URL not found")
		default:
			return nil, fmt.Errorf("error when getting url: %w\n", err)
		}
	}
	return stats, nil
}

func (a *App) increaseNumRedirects(ctx context.Context, shortURL string) {
	err := a.store.IncreaseNumRedirects(ctx, shortURL)
	if err != nil {
		log.Println(err)
	}
}
