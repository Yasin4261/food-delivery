package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// MaxImageBytes caps uploaded photos (handlers enforce it with
// http.MaxBytesReader; the service re-checks when decoding).
const MaxImageBytes = 5 << 20 // 5 MiB

// MaxGalleryImages caps how many photos a dish gallery holds (#93).
const MaxGalleryImages = 5

// jpegQuality for re-encoded uploads.
const jpegQuality = 85

// UploadService implements the photo-upload use cases: a chef sets a photo on
// one of their own dishes or on their kitchen profile. Images are decoded and
// re-encoded before storage — that both proves the payload really is a
// JPEG/PNG (whatever the client claimed) and strips EXIF metadata (GPS
// coordinates of someone's home kitchen do not belong on a public URL).
type UploadService struct {
	store domain.FileStore
	chefs domain.ChefRepository
	items domain.MenuItemRepository
}

// NewUploadService builds an UploadService.
func NewUploadService(store domain.FileStore, chefs domain.ChefRepository, items domain.MenuItemRepository) *UploadService {
	return &UploadService{store: store, chefs: chefs, items: items}
}

// UploadDishImage stores a photo for one of the caller's own dishes and
// returns the public URL.
func (s *UploadService) UploadDishImage(ctx context.Context, userID, itemID int, content io.Reader) (string, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return "", err
	}
	if item.ChefID != chef.ID {
		return "", domain.ErrForbidden
	}

	url, err := s.process(ctx, content)
	if err != nil {
		return "", err
	}
	if err := s.items.SetImageURL(ctx, itemID, url); err != nil {
		return "", err
	}
	return url, nil
}

// UploadKitchenImage stores the caller's kitchen photo and returns the public
// URL.
func (s *UploadService) UploadKitchenImage(ctx context.Context, userID int, content io.Reader) (string, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	url, err := s.process(ctx, content)
	if err != nil {
		return "", err
	}
	if err := s.chefs.SetImageURL(ctx, chef.ID, url); err != nil {
		return "", err
	}
	return url, nil
}

// ownedDish resolves a dish and confirms the caller owns it.
func (s *UploadService) ownedDish(ctx context.Context, userID, itemID int) (*domain.MenuItem, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.ChefID != chef.ID {
		return nil, domain.ErrForbidden
	}
	return item, nil
}

// AddDishGalleryImage appends a processed photo to a dish's gallery (owner
// only) and returns the new URL list. The count is capped at
// MaxGalleryImages.
func (s *UploadService) AddDishGalleryImage(ctx context.Context, userID, itemID int, content io.Reader) ([]string, error) {
	item, err := s.ownedDish(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	urls := parseImages(item.Images)
	if len(urls) >= MaxGalleryImages {
		return nil, domain.ErrGalleryFull
	}
	url, err := s.process(ctx, content)
	if err != nil {
		return nil, err
	}
	urls = append(urls, url)
	if err := s.items.SetImages(ctx, itemID, encodeImages(urls)); err != nil {
		return nil, err
	}
	return urls, nil
}

// RemoveDishGalleryImage drops one URL from a dish's gallery (owner only) and
// returns the remaining list.
func (s *UploadService) RemoveDishGalleryImage(ctx context.Context, userID, itemID int, url string) ([]string, error) {
	item, err := s.ownedDish(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	kept := make([]string, 0)
	for _, u := range parseImages(item.Images) {
		if u != url {
			kept = append(kept, u)
		}
	}
	if err := s.items.SetImages(ctx, itemID, encodeImages(kept)); err != nil {
		return nil, err
	}
	return kept, nil
}

// parseImages decodes the JSON-array column into a slice (nil/invalid -> empty).
func parseImages(raw *string) []string {
	if raw == nil || *raw == "" {
		return []string{}
	}
	var urls []string
	if err := json.Unmarshal([]byte(*raw), &urls); err != nil {
		return []string{}
	}
	return urls
}

// encodeImages serialises the slice back to the column (empty -> nil, so the
// column clears rather than storing "[]").
func encodeImages(urls []string) *string {
	if len(urls) == 0 {
		return nil
	}
	b, _ := json.Marshal(urls)
	s := string(b)
	return &s
}

// process decodes the upload (rejecting anything that isn't a real JPEG/PNG),
// re-encodes it (dropping EXIF and any trailing payload), and stores the
// result.
func (s *UploadService) process(ctx context.Context, content io.Reader) (string, error) {
	img, format, err := image.Decode(io.LimitReader(content, MaxImageBytes+1))
	if err != nil {
		return "", domain.ErrUnsupportedImage
	}

	var buf bytes.Buffer
	var ext string
	switch format {
	case "jpeg":
		ext = ".jpg"
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality})
	case "png":
		ext = ".png"
		err = png.Encode(&buf, img)
	default:
		return "", domain.ErrUnsupportedImage
	}
	if err != nil {
		return "", fmt.Errorf("re-encode image: %w", err)
	}
	return s.store.Save(ctx, ext, &buf)
}
