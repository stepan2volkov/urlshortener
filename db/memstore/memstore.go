package memstore

import (
	"context"
	"database/sql"
	"sync"

	"github.com/stepan2volkov/urlshortener/app"
)

var _ app.URLStore = &MemStore{}

type MemStore struct {
	sync.Mutex
	shortMap    map[string]app.URL
	originalMap map[string]app.URL
}

func NewMemStore() *MemStore {
	return &MemStore{
		shortMap:    make(map[string]app.URL),
		originalMap: make(map[string]app.URL),
	}
}

func (us *MemStore) Create(ctx context.Context, originalURL string) (*app.URL, error) {
	us.Lock()
	defer us.Unlock()

	url := app.URL{
		ID:          len(us.shortMap) + len(us.originalMap),
		OriginalURL: originalURL,
	}

	us.originalMap[originalURL] = url
	return &url, nil
}

func (us *MemStore) UpdateURL(ctx context.Context, url *app.URL) error {
	us.Lock()
	defer us.Unlock()

	delete(us.originalMap, url.OriginalURL)

	us.shortMap[url.ShortURL] = *url
	return nil
}
func (us *MemStore) GetOriginalURL(ctx context.Context, shortURL string) (*app.URL, error) {
	us.Lock()
	defer us.Unlock()

	if url, found := us.shortMap[shortURL]; found {
		return &url, nil
	}

	return nil, sql.ErrNoRows
}

func (us *MemStore) GetStats(ctx context.Context, shortURL string) (*app.Stats, error) {
	us.Lock()
	defer us.Unlock()

	if url, found := us.shortMap[shortURL]; found {
		return &app.Stats{
			ShortURL:     url.ShortURL,
			NumRedirects: url.NumRedirects,
		}, nil
	}

	return nil, sql.ErrNoRows
}

func (us *MemStore) IncreaseNumRedirects(ctx context.Context, shortURL string) error {
	us.Lock()
	defer us.Unlock()

	if foundUser, found := us.shortMap[shortURL]; found {
		foundUser.NumRedirects += 1
		us.shortMap[shortURL] = foundUser
		return nil
	}
	return sql.ErrNoRows
}
