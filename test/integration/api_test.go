// Package integration wires the real handler/router/store stack together
// behind an httptest server and drives it through HTTP the way a client
// would. It is intentionally a black-box test at the API surface.
package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/askthechain/atcutils/pkg/logger"
	dbmodel "github.com/khetzl/shop-db-models/pkg/storage/dbmodel"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/khetzl/shop-customer-api/internal/api/handler"
	"github.com/khetzl/shop-customer-api/internal/api/router"
	"github.com/khetzl/shop-customer-api/internal/domain/models"
	"github.com/khetzl/shop-customer-api/internal/services/itemstore"
	"github.com/khetzl/shop-customer-api/internal/storage/postgres"
)

// newTestServer wires the full stack behind an httptest server, backed by a
// database holding one parent item with two items.
func newTestServer(t *testing.T) string {
	t.Helper()

	// The full stack is backed by PostgreSQL. Skip when no database is
	// configured so the suite stays runnable without external infrastructure.
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL or DATABASE_URL to run the integration suite")
	}

	db, err := postgres.Connect(dsn)
	if err != nil {
		t.Fatalf("postgres.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`TRUNCATE item_relations, parent_item_relations, item_photos, items, parent_item_sizes, parent_items CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	ctx := context.Background()
	parent := &dbmodel.ParentItem{ID: "p-1", Name: "Hoodie", Category: "tops", Material: "cotton", Colour: "grey", BasePrice: 4500}
	if err := parent.Insert(ctx, db, boil.Infer()); err != nil {
		t.Fatalf("insert parent item: %v", err)
	}
	for _, i := range []*dbmodel.Item{
		{ID: "i-2", Name: "Hoodie, zip", ParentItemID: parent.ID},
		{ID: "i-1", Name: "Hoodie", ParentItemID: parent.ID},
	} {
		if err := i.Insert(ctx, db, boil.Infer()); err != nil {
			t.Fatalf("insert item %s: %v", i.ID, err)
		}
	}

	logger := atclogger.NewMockLogger()
	itemStore, err := itemstore.New(logger, db)
	if err != nil {
		t.Fatalf("itemstore.New: %v", err)
	}
	h, err := handler.New(logger, itemStore)
	if err != nil {
		t.Fatalf("handler.New: %v", err)
	}
	r, err := router.New(logger, h)
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestAPI_ListItems(t *testing.T) {
	url := newTestServer(t)

	for _, path := range []string{router.Items, router.ForYou} {
		t.Run(path, func(t *testing.T) {
			res, err := http.Get(url + path)
			if err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
			defer res.Body.Close()
			raw, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status: got %d, want %d; body=%s", res.StatusCode, http.StatusOK, raw)
			}

			var body struct {
				Items []models.Item `json:"items"`
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("decode: %v; body=%s", err, raw)
			}
			if len(body.Items) != 2 {
				t.Fatalf("expected 2 items, got %d: %s", len(body.Items), raw)
			}
			for i, wantID := range []string{"i-1", "i-2"} {
				got := body.Items[i]
				if got.ID != wantID {
					t.Errorf("item %d: id got %q, want %q", i, got.ID, wantID)
				}
				if got.Category != "tops" || got.Price != 4500 || got.Colour != "grey" {
					t.Errorf("item %d: parent properties not resolved: %+v", i, got)
				}
			}
		})
	}
}
