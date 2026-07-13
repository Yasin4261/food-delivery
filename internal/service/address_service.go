package service

import (
	"context"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AddressService implements the customer address-book use cases. Every
// operation is scoped to the caller: reading, editing or deleting another
// user's address returns domain.ErrForbidden.
type AddressService struct {
	addresses domain.AddressRepository
}

// NewAddressService builds an AddressService.
func NewAddressService(addresses domain.AddressRepository) *AddressService {
	return &AddressService{addresses: addresses}
}

// AddressInput is the editable slice of an address-book entry.
type AddressInput struct {
	Label     string
	Address   string
	City      string
	Latitude  *float64
	Longitude *float64
	IsDefault bool
}

func (in AddressInput) apply(a *domain.Address) {
	a.Label = in.Label
	a.Address = in.Address
	a.City = optional(in.City)
	a.Latitude = in.Latitude
	a.Longitude = in.Longitude
	a.IsDefault = in.IsDefault
}

// Create adds an address to the caller's book. The first address becomes the
// default automatically.
func (s *AddressService) Create(ctx context.Context, userID int, in AddressInput) (*domain.Address, error) {
	a := &domain.Address{UserID: userID}
	in.apply(a)
	if err := a.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.addresses.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		a.IsDefault = true
	}

	if err := s.addresses.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// List returns the caller's addresses (default first).
func (s *AddressService) List(ctx context.Context, userID int) ([]*domain.Address, error) {
	return s.addresses.ListByUser(ctx, userID)
}

// Update edits one of the caller's own addresses.
func (s *AddressService) Update(ctx context.Context, userID, addressID int, in AddressInput) (*domain.Address, error) {
	a, err := s.owned(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}
	in.apply(a)
	if err := a.Validate(); err != nil {
		return nil, err
	}
	if err := s.addresses.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Delete removes one of the caller's own addresses.
func (s *AddressService) Delete(ctx context.Context, userID, addressID int) error {
	if _, err := s.owned(ctx, userID, addressID); err != nil {
		return err
	}
	return s.addresses.Delete(ctx, addressID)
}

// owned loads an address and enforces that the caller owns it.
func (s *AddressService) owned(ctx context.Context, userID, addressID int) (*domain.Address, error) {
	a, err := s.addresses.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if a.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return a, nil
}
