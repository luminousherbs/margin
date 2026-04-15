package service

import (
	"context"
	"sync"
	"time"

	"margin.at/internal/domain"
)

const defaultProfileTTL = 5 * time.Minute

type profileCacheEntry struct {
	author    domain.Author
	expiresAt time.Time
}

type ProfileService struct {
	repo  domain.ProfileRepository
	ttl   time.Duration
	mu    sync.RWMutex
	cache map[string]profileCacheEntry
}

func NewProfileService(repo domain.ProfileRepository) *ProfileService {
	return &ProfileService{
		repo:  repo,
		ttl:   defaultProfileTTL,
		cache: make(map[string]profileCacheEntry),
	}
}

func (s *ProfileService) GetProfiles(ctx context.Context, dids []string) (map[string]domain.Author, error) {
	now := time.Now()
	result := make(map[string]domain.Author, len(dids))
	var missing []string

	s.mu.RLock()
	for _, did := range dids {
		if e, ok := s.cache[did]; ok && now.Before(e.expiresAt) {
			result[did] = e.author
		} else {
			missing = append(missing, did)
		}
	}
	s.mu.RUnlock()

	if len(missing) == 0 {
		return result, nil
	}

	fetched, err := s.repo.GetProfiles(ctx, missing)
	if err != nil {
		return result, err
	}

	expiry := now.Add(s.ttl)
	s.mu.Lock()
	for did, author := range fetched {
		s.cache[did] = profileCacheEntry{author: author, expiresAt: expiry}
		result[did] = author
	}
	s.mu.Unlock()

	return result, nil
}

func (s *ProfileService) GetProfile(ctx context.Context, did string) (*domain.Profile, error) {
	return s.repo.GetProfile(ctx, did)
}

func (s *ProfileService) UpsertProfile(ctx context.Context, p *domain.Profile) error {
	if err := s.repo.UpsertProfile(ctx, p); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.cache, p.AuthorDID)
	s.mu.Unlock()
	return nil
}
