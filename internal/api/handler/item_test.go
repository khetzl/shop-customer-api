package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/askthechain/atcutils/pkg/logger"

	"github.com/khetzl/shop-customer-api/internal/api/handler"
	"github.com/khetzl/shop-customer-api/internal/api/router"
	"github.com/khetzl/shop-customer-api/internal/domain/models"
)

type fakeItemStore struct {
	items []models.Item
	err   error
}

func (f fakeItemStore) GetAll() ([]models.Item, error) {
	return f.items, f.err
}

func newRouter(t *testing.T, store handler.ItemStore) http.Handler {
	t.Helper()

	h, err := handler.New(atclogger.NewMockLogger(), store)
	if err != nil {
		t.Fatalf("handler.New: %v", err)
	}
	r, err := router.New(atclogger.NewMockLogger(), h)
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}
	return r
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestItemRoutes_ReturnAllItems(t *testing.T) {
	items := []models.Item{
		{ID: "i-1", Name: "Tee", Price: 2500, Sizes: []string{"S", "M"}},
		{ID: "i-2", Name: "Cap", Price: 1500, LimitedEdition: &models.LimitedEdition{Max: 10, Available: 3}},
	}
	r := newRouter(t, fakeItemStore{items: items})

	for _, path := range []string{router.Items, router.ForYou} {
		t.Run(path, func(t *testing.T) {
			rec := get(r, path)
			if rec.Code != http.StatusOK {
				t.Fatalf("status: got %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var body struct {
				Items []models.Item `json:"items"`
				Error any           `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
			}
			if !reflect.DeepEqual(body.Items, items) {
				t.Errorf("items:\n got: %+v\nwant: %+v", body.Items, items)
			}
			if body.Error != nil {
				t.Errorf("error: got %v, want nil", body.Error)
			}
		})
	}
}

func TestItemRoutes_EmptyListIsJSONArray(t *testing.T) {
	r := newRouter(t, fakeItemStore{items: []models.Item{}})

	rec := get(r, router.Items)
	if got, want := rec.Body.String(), `{"error":null,"items":[]}`; got != want {
		t.Errorf("body: got %s, want %s", got, want)
	}
}

func TestItemRoutes_StoreErrorIs500(t *testing.T) {
	r := newRouter(t, fakeItemStore{err: errors.New("db down")})

	for _, path := range []string{router.Items, router.ForYou} {
		if rec := get(r, path); rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: status got %d, want %d", path, rec.Code, http.StatusInternalServerError)
		}
	}
}
