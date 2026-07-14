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

// Sort orders valid per search type. The dish list additionally allows price
// ordering.
var (
	chefSorts = map[string]bool{
		domain.SortDefault: true, domain.SortRating: true, domain.SortPopular: true,
	}
	dishSorts = map[string]bool{
		domain.SortDefault: true, domain.SortRating: true, domain.SortPopular: true,
		domain.SortPriceAsc: true, domain.SortPriceDesc: true,
	}
)

// validateFilters rejects out-of-range or unknown filter values before they
// reach the adapter. allowed is the sort whitelist for the search type.
func validateFilters(f domain.SearchFilters, allowed map[string]bool) error {
	if !allowed[f.Sort] {
		return ValidationError{Msg: "unknown sort: must be rating, popular, price_asc or price_desc"}
	}
	if f.MinRating < 0 || f.MinRating > 5 {
		return ValidationError{Msg: "min_rating must be between 0 and 5"}
	}
	if f.MinPrice < 0 || f.MaxPrice < 0 {
		return ValidationError{Msg: "prices cannot be negative"}
	}
	if f.MaxPrice > 0 && f.MinPrice > f.MaxPrice {
		return ValidationError{Msg: "min_price cannot exceed max_price"}
	}
	return nil
}

// Chefs searches chefs, returning a page and the total.
func (s *SearchService) Chefs(ctx context.Context, q string, f domain.SearchFilters, limit, offset int) ([]*domain.Chef, int, error) {
	q, limit, offset, err := s.normalise(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := validateFilters(f, chefSorts); err != nil {
		return nil, 0, err
	}
	return s.search.SearchChefs(ctx, q, f, limit, offset)
}

// Foods searches dishes, returning a page and the total.
func (s *SearchService) Foods(ctx context.Context, q string, f domain.SearchFilters, limit, offset int) ([]*domain.MenuItem, int, error) {
	q, limit, offset, err := s.normalise(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := validateFilters(f, dishSorts); err != nil {
		return nil, 0, err
	}
	return s.search.SearchMenuItems(ctx, q, f, limit, offset)
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
