package handler_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// multipartImage builds a multipart body with an "image" field containing a
// real 2x2 JPEG (or the given raw payload).
func multipartImage(t *testing.T, payload []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("image", "photo.jpg")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	if _, err := fw.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = w.Close()
	return &body, w.FormDataContentType()
}

func realJPEG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}

// doUpload posts a multipart body with the bearer token.
func doUpload(t *testing.T, srv http.Handler, path, token string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestUploadHTTP_DishImage(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")

	// No token -> 401; customer -> 403 (role guard).
	body, ct := multipartImage(t, realJPEG(t))
	if rec := doUpload(t, srv, "/api/v2/menu-items/1/image", "", body, ct); rec.Code != http.StatusUnauthorized {
		t.Errorf("anonymous upload = %d, want 401", rec.Code)
	}
	cust := registerCustomerToken(t, srv, "cust", "cust@example.com")
	body, ct = multipartImage(t, realJPEG(t))
	if rec := doUpload(t, srv, "/api/v2/menu-items/1/image", cust, body, ct); rec.Code != http.StatusForbidden {
		t.Errorf("customer upload = %d, want 403", rec.Code)
	}

	// Owner uploads a real image -> 200 with the URL; the dish reflects it.
	body, ct = multipartImage(t, realJPEG(t))
	rec := doUpload(t, srv, "/api/v2/menu-items/"+itoa(itemID)+"/image", chefToken, body, ct)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload = %d (%s)", rec.Code, rec.Body)
	}
	var out map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if !strings.HasPrefix(out["image_url"], "/uploads/") || !strings.HasSuffix(out["image_url"], ".jpg") {
		t.Fatalf("image_url = %q", out["image_url"])
	}

	// Another chef cannot photograph someone else's dish.
	other, _ := seedChefWithItem(t, srv, "chefb", "chefb@example.com")
	body, ct = multipartImage(t, realJPEG(t))
	if rec := doUpload(t, srv, "/api/v2/menu-items/"+itoa(itemID)+"/image", other, body, ct); rec.Code != http.StatusForbidden {
		t.Errorf("foreign chef upload = %d, want 403", rec.Code)
	}

	// Non-image payload -> 400.
	body, ct = multipartImage(t, []byte("<html>not an image</html>"))
	if rec := doUpload(t, srv, "/api/v2/menu-items/"+itoa(itemID)+"/image", chefToken, body, ct); rec.Code != http.StatusBadRequest {
		t.Errorf("html upload = %d, want 400", rec.Code)
	}

	// The uploaded file is served back; traversal-shaped names are not.
	req := httptest.NewRequest(http.MethodGet, out["image_url"], nil)
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Errorf("serving %q = %d, want 200", out["image_url"], rec2.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/uploads/..%2f..%2fetc%2fpasswd", nil)
	rec2 = httptest.NewRecorder()
	srv.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("traversal request = %d, want 404", rec2.Code)
	}
}

func TestUploadHTTP_KitchenImage(t *testing.T) {
	srv := newTestServer()
	chefToken, _ := seedChefWithItem(t, srv, "chefa", "chefa@example.com")

	body, ct := multipartImage(t, realJPEG(t))
	rec := doUpload(t, srv, "/api/v2/chefs/me/image", chefToken, body, ct)
	if rec.Code != http.StatusOK {
		t.Fatalf("kitchen upload = %d (%s)", rec.Code, rec.Body)
	}
	var out map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &out)

	// The chef profile now carries the URL.
	me := do(t, srv, http.MethodGet, "/api/v2/chefs/me", chefToken, "")
	var chef map[string]any
	_ = json.Unmarshal(me.Body.Bytes(), &chef)
	if chef["image_url"] != out["image_url"] {
		t.Errorf("chef image_url = %v, want %v", chef["image_url"], out["image_url"])
	}
}

func TestUploadHTTP_DishGallery(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	path := "/api/v2/menu-items/" + itoa(itemID) + "/images"

	// Auth: anon 401, customer 403.
	body, ct := multipartImage(t, realJPEG(t))
	if rec := doUpload(t, srv, path, "", body, ct); rec.Code != http.StatusUnauthorized {
		t.Errorf("anon gallery add = %d, want 401", rec.Code)
	}
	cust := registerCustomerToken(t, srv, "cust", "cust@example.com")
	body, ct = multipartImage(t, realJPEG(t))
	if rec := doUpload(t, srv, path, cust, body, ct); rec.Code != http.StatusForbidden {
		t.Errorf("customer gallery add = %d, want 403", rec.Code)
	}

	// Owner appends two, list grows.
	for want := 1; want <= 2; want++ {
		body, ct = multipartImage(t, realJPEG(t))
		rec := doUpload(t, srv, path, chefToken, body, ct)
		if rec.Code != http.StatusOK {
			t.Fatalf("gallery add = %d (%s)", rec.Code, rec.Body)
		}
		var out struct {
			Images []string `json:"images"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		if len(out.Images) != want {
			t.Fatalf("gallery size = %d, want %d", len(out.Images), want)
		}
		if want == 2 {
			// Remove the first; one remains.
			rec = do(t, srv, http.MethodDelete, path+"?url="+out.Images[0], chefToken, "")
			var after struct {
				Images []string `json:"images"`
			}
			_ = json.Unmarshal(rec.Body.Bytes(), &after)
			if rec.Code != http.StatusOK || len(after.Images) != 1 {
				t.Errorf("gallery remove = %d/%d, want 200/1", rec.Code, len(after.Images))
			}
		}
	}

	// Missing url on delete -> 400.
	if rec := do(t, srv, http.MethodDelete, path, chefToken, ""); rec.Code != http.StatusBadRequest {
		t.Errorf("delete without url = %d, want 400", rec.Code)
	}
}
