package handler

import (
	"net/http"
	"runtime"
)

// VersionHandler exposes build metadata. The version string is injected at
// build time via -ldflags "-X main.version=$(git describe --tags)" and passed
// down from the composition root.
type VersionHandler struct {
	version string
}

// NewVersionHandler builds a VersionHandler.
func NewVersionHandler(version string) *VersionHandler {
	return &VersionHandler{version: version}
}

// Version handles GET /version.
func (h *VersionHandler) Version(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"version": h.version,
		"go":      runtime.Version(),
	})
}
