package handler

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
	"github.com/Yasin4261/food-delivery/internal/storage"
)

// UploadHandler exposes photo upload (chef role; ownership enforced in the
// service) and serves the stored files.
type UploadHandler struct {
	uploads   *service.UploadService
	uploadDir string
}

// NewUploadHandler builds an UploadHandler. uploadDir is where the local
// store keeps files (used by Serve).
func NewUploadHandler(uploads *service.UploadService, uploadDir string) *UploadHandler {
	return &UploadHandler{uploads: uploads, uploadDir: uploadDir}
}

// imageFromRequest extracts the multipart "image" file, capped at
// MaxImageBytes. It returns nil after writing the error response.
func (h *UploadHandler) imageFromRequest(w http.ResponseWriter, r *http.Request) *http.Request {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxImageBytes+4096) // + form overhead
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			respondError(w, http.StatusRequestEntityTooLarge, "image must be at most 5 MB")
			return nil
		}
		respondError(w, http.StatusBadRequest, "expected multipart form data with an image field")
		return nil
	}
	return r
}

// DishImage handles POST /api/v2/menu-items/{id}/image (chef role).
func (h *UploadHandler) DishImage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	itemID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu item id")
		return
	}
	if r = h.imageFromRequest(w, r); r == nil {
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing image field")
		return
	}
	defer file.Close()

	url, err := h.uploads.UploadDishImage(r.Context(), claims.UserID, itemID, file)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"image_url": url})
}

// DishGalleryAdd handles POST /api/v2/menu-items/{id}/images (chef role) —
// append a photo to the dish gallery; returns the URL list.
func (h *UploadHandler) DishGalleryAdd(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	itemID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu item id")
		return
	}
	if r = h.imageFromRequest(w, r); r == nil {
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing image field")
		return
	}
	defer file.Close()

	urls, err := h.uploads.AddDishGalleryImage(r.Context(), claims.UserID, itemID, file)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string][]string{"images": urls})
}

// DishGalleryRemove handles DELETE /api/v2/menu-items/{id}/images?url=… (chef
// role) — drop one photo; returns the remaining URL list.
func (h *UploadHandler) DishGalleryRemove(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	itemID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu item id")
		return
	}
	url := r.URL.Query().Get("url")
	if url == "" {
		respondError(w, http.StatusBadRequest, "url query param is required")
		return
	}
	urls, err := h.uploads.RemoveDishGalleryImage(r.Context(), claims.UserID, itemID, url)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string][]string{"images": urls})
}

// KitchenImage handles POST /api/v2/chefs/me/image (chef role).
func (h *UploadHandler) KitchenImage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	if r = h.imageFromRequest(w, r); r == nil {
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing image field")
		return
	}
	defer file.Close()

	url, err := h.uploads.UploadKitchenImage(r.Context(), claims.UserID, file)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"image_url": url})
}

// Serve handles GET /uploads/{file} (public). Only names the store could have
// generated are served — anything else (dotfiles, traversal attempts, other
// patterns) is a 404 before touching the filesystem.
func (h *UploadHandler) Serve(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("file")
	if !storage.ValidName(name) {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	http.ServeFile(w, r, filepath.Join(h.uploadDir, name))
}
