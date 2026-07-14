package service

import (
	"bytes"
	"context"
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
