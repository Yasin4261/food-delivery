package handler_test

import (
	"context"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakePromoRepo is an in-memory domain.PromoRepository for HTTP tests.
type fakePromoRepo struct {
	byID   map[int]*domain.PromoCode
	byCode map[string]*domain.PromoCode
	nextID int
}

func newFakePromoRepo() *fakePromoRepo {
	return &fakePromoRepo{byID: map[int]*domain.PromoCode{}, byCode: map[string]*domain.PromoCode{}, nextID: 1}
}

func (f *fakePromoRepo) Create(_ context.Context, p *domain.PromoCode) error {
	if _, ok := f.byCode[p.Code]; ok {
		return domain.ErrPromoExists
	}
	p.ID = f.nextID
	f.nextID++
	cp := *p
	f.byID[p.ID] = &cp
	f.byCode[p.Code] = &cp
	return nil
}

func (f *fakePromoRepo) FindByCode(_ context.Context, code string) (*domain.PromoCode, error) {
	if p, ok := f.byCode[domain.NormaliseCode(code)]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, domain.ErrPromoNotFound
}

func (f *fakePromoRepo) Redeem(_ context.Context, id int) error {
	p, ok := f.byID[id]
	if !ok {
		return domain.ErrPromoNotFound
	}
	if p.UsageLimit > 0 && p.UsedCount >= p.UsageLimit {
		return domain.ErrPromoUsedUp
	}
	p.UsedCount++
	return nil
}

func (f *fakePromoRepo) List(_ context.Context, limit, offset int) ([]*domain.PromoCode, int, error) {
	out := make([]*domain.PromoCode, 0, len(f.byID))
	for _, p := range f.byID {
		cp := *p
		out = append(out, &cp)
	}
	return out, len(out), nil
}

func (f *fakePromoRepo) SetActive(_ context.Context, id int, active bool) error {
	p, ok := f.byID[id]
	if !ok {
		return domain.ErrPromoNotFound
	}
	p.IsActive = active
	return nil
}
