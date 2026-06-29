package service

import (
	"context"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// SearchService implements catalogue search. It depends only on the
// domain.SearchRepository port.
type SearchService struct {
	search domain.SearchRepository
}

// NewSearchService builds a SearchService.
func NewSearchService(search domain.SearchRepository) *SearchService {
	return &SearchService{search: search}
}

// Chefs searches chefs, returning a page and the total.
func (s *SearchService) Chefs(ctx context.Context, q string, limit, offset int) ([]*domain.Chef, int, error) {
	q, limit, offset, err := s.normalise(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return s.search.SearchChefs(ctx, q, limit, offset)
}

// Foods searches dishes, returning a page and the total.
func (s *SearchService) Foods(ctx context.Context, q string, limit, offset int) ([]*domain.MenuItem, int, error) {
	q, limit, offset, err := s.normalise(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return s.search.SearchMenuItems(ctx, q, limit, offset)
}

// Users searches users (admin-only at the handler layer), returning a page and
// the total. Password hashes are cleared before returning.
func (s *SearchService) Users(ctx context.Context, q string, limit, offset int) ([]*domain.User, int, error) {
	q, limit, offset, err := s.normalise(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	users, total, err := s.search.SearchUsers(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	return users, total, nil
}

// normalise trims the query (rejecting empty) and clamps the paging window.
func (s *SearchService) normalise(q string, limit, offset int) (string, int, int, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "", 0, 0, ValidationError{Msg: "q (search query) is required"}
	}
	limit, offset = normalisePaging(limit, offset)
	return q, limit, offset, nil
}
