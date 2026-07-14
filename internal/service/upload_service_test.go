package service_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeFileStore records saves in memory.
type fakeFileStore struct {
	saved map[string][]byte
	n     int
}

func newFakeFileStore() *fakeFileStore { return &fakeFileStore{saved: map[string][]byte{}} }

func (f *fakeFileStore) Save(_ context.Context, ext string, content io.Reader) (string, error) {
	f.n++
	data, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}
	url := "/uploads/fake-" + strings.Repeat("0", f.n) + ext
	f.saved[url] = data
	return url, nil
}

// tinyJPEG / tinyPNG return minimal real images.
func tinyJPEG(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return &buf
}

func tinyPNG(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return &buf
}

// uploadFixture: user 1 is a chef owning dish 10; user 2 is a chef owning
// nothing; user 3 has no chef profile.
func uploadFixture(t *testing.T) (*service.UploadService, *fakeFileStore, *fakeMenuItemRepo, *fakeChefRepo) {
	t.Helper()
	ctx := context.Background()
	chefs := newFakeChefRepo()
	for _, uid := range []int{1, 2} {
		if err := chefs.Create(ctx, &domain.Chef{UserID: uid, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}
	items := newFakeMenuItemRepo()
	item := domain.NewMenuItem(1, 1, "Soup", 5)
	if err := items.Create(ctx, item); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	store := newFakeFileStore()
	return service.NewUploadService(store, chefs, items), store, items, chefs
}

func TestUploadService_DishImage(t *testing.T) {
	svc, store, items, _ := uploadFixture(t)
	ctx := context.Background()

	url, err := svc.UploadDishImage(ctx, 1, 1, tinyJPEG(t))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !strings.HasSuffix(url, ".jpg") {
		t.Errorf("url = %q, want .jpg", url)
	}
	// The stored bytes are a decodable image (re-encoded, not passthrough).
	if _, format, err := image.Decode(bytes.NewReader(store.saved[url])); err != nil || format != "jpeg" {
		t.Errorf("stored content not a jpeg: %v/%q", err, format)
	}
	// The dish row points at the new URL.
	item, _ := items.FindByID(ctx, 1)
	if item.ImageURL == nil || *item.ImageURL != url {
		t.Errorf("image_url not persisted: %+v", item.ImageURL)
	}

	// PNG keeps its format.
	pngURL, err := svc.UploadDishImage(ctx, 1, 1, tinyPNG(t))
	if err != nil || !strings.HasSuffix(pngURL, ".png") {
		t.Errorf("png upload = %q (%v), want .png", pngURL, err)
	}
}

func TestUploadService_Authorization(t *testing.T) {
	svc, store, _, _ := uploadFixture(t)
	ctx := context.Background()

	// Another chef's dish -> forbidden, nothing stored.
	if _, err := svc.UploadDishImage(ctx, 2, 1, tinyJPEG(t)); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("foreign dish = %v, want ErrForbidden", err)
	}
	// No chef profile -> chef not found.
	if _, err := svc.UploadDishImage(ctx, 3, 1, tinyJPEG(t)); !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("no profile = %v, want ErrChefNotFound", err)
	}
	if store.n != 0 {
		t.Errorf("store touched %d times on rejected uploads, want 0", store.n)
	}
}

func TestUploadService_RejectsNonImages(t *testing.T) {
	svc, store, _, _ := uploadFixture(t)
	ctx := context.Background()

	payloads := map[string]io.Reader{
		"html":       strings.NewReader("<script>alert(1)</script>"),
		"empty":      strings.NewReader(""),
		"fake magic": strings.NewReader("\xff\xd8\xff not really a jpeg"),
	}
	for name, r := range payloads {
		if _, err := svc.UploadDishImage(ctx, 1, 1, r); !errors.Is(err, domain.ErrUnsupportedImage) {
			t.Errorf("%s = %v, want ErrUnsupportedImage", name, err)
		}
	}
	if store.n != 0 {
		t.Errorf("store touched for invalid payloads")
	}
}

// Re-encoding strips trailing/embedded metadata: bytes after the JPEG EOI
// (where EXIF-hidden payloads often ride) never reach storage.
func TestUploadService_ReencodeDropsTrailingPayload(t *testing.T) {
	svc, store, _, _ := uploadFixture(t)
	ctx := context.Background()

	img := tinyJPEG(t)
	img.WriteString("SECRET-PAYLOAD-AFTER-EOI")
	url, err := svc.UploadKitchenImage(ctx, 1, img)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if bytes.Contains(store.saved[url], []byte("SECRET-PAYLOAD-AFTER-EOI")) {
		t.Error("trailing payload survived re-encoding")
	}
}

func TestUploadService_KitchenImage(t *testing.T) {
	svc, _, _, chefs := uploadFixture(t)
	ctx := context.Background()

	url, err := svc.UploadKitchenImage(ctx, 1, tinyPNG(t))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	chef, _ := chefs.FindByUserID(ctx, 1)
	if chef.ImageURL == nil || *chef.ImageURL != url {
		t.Errorf("kitchen image_url not persisted: %+v", chef.ImageURL)
	}
}
