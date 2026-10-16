package itemstore

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/askthechain/atcutils/pkg/logger"
	dbmodel "github.com/khetzl/shop-db-models/pkg/storage/dbmodel"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/khetzl/shop-customer-api/internal/domain/models"
	"github.com/khetzl/shop-customer-api/internal/storage/postgres"
)

// newStore connects to the test database (provisioned from shop-db-models'
// db-scripts) and clears the item tables. The tailoring reference rows the
// schema script inserts are left in place. It skips the test when no database
// is configured so the suite stays runnable without external infrastructure.
//
// Set TEST_DATABASE_URL (or DATABASE_URL) to a throwaway PostgreSQL database to
// exercise these tests.
func newStore(t *testing.T) (*Service, *sql.DB) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL or DATABASE_URL to run itemstore Postgres tests")
	}

	db, err := postgres.Connect(dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`TRUNCATE item_relations, parent_item_relations, item_photos, items, parent_item_sizes, parent_items CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	s, err := New(atclogger.NewMockLogger(), db)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	return s, db
}

func TestGetAll_EmptyDatabase(t *testing.T) {
	s, _ := newStore(t)

	items, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("expected an empty, non-nil slice, got %#v", items)
	}
}

func TestGetAll_ResolvesParentItemProperties(t *testing.T) {
	s, db := newStore(t)
	ctx := context.Background()

	tee := &dbmodel.ParentItem{
		ID:          "p-tee",
		Name:        "Tee",
		Description: null.StringFrom("Parent description"),
		CareGuide:   null.StringFrom("Parent care guide"),
		Category:    "tops",
		Material:    "cotton",
		Colour:      "white",
		TailoringID: null.StringFrom("normal"),
		BasePrice:   2500,
	}
	capParent := &dbmodel.ParentItem{
		ID:        "p-cap",
		Name:      "Cap",
		Category:  "accessories",
		Material:  "wool",
		Colour:    "navy",
		BasePrice: 1500,
	}
	for _, p := range []*dbmodel.ParentItem{tee, capParent} {
		if err := p.Insert(ctx, db, boil.Infer()); err != nil {
			t.Fatalf("insert parent item %s: %v", p.ID, err)
		}
	}
	// Inserted out of order to check the sizes come back by position.
	if err := tee.AddParentItemSizes(ctx, db, true,
		&dbmodel.ParentItemSize{ID: "p-tee-l", Size: "L", Position: null.IntFrom(2)},
		&dbmodel.ParentItemSize{ID: "p-tee-s", Size: "S", Position: null.IntFrom(0)},
		&dbmodel.ParentItemSize{ID: "p-tee-m", Size: "M", Position: null.IntFrom(1)},
	); err != nil {
		t.Fatalf("insert sizes: %v", err)
	}
	if err := tee.AddRelatedParentItems(ctx, db, false, capParent); err != nil {
		t.Fatalf("relate parent items: %v", err)
	}

	limited := &dbmodel.Item{
		ID:                      "i-tee-1",
		Name:                    "Tee, limited",
		Description:             null.StringFrom("Own description"),
		TailoringID:             null.StringFrom("slim"),
		ParentItemID:            tee.ID,
		ReleaseDate:             null.TimeFrom(time.Date(2026, 10, 1, 9, 30, 0, 0, time.UTC)),
		LimitedEditionMax:       null.IntFrom(100),
		LimitedEditionAvailable: null.IntFrom(42),
	}
	plain := &dbmodel.Item{ID: "i-tee-2", Name: "Tee", ParentItemID: tee.ID}
	capItem := &dbmodel.Item{ID: "i-cap-1", Name: "Cap", ParentItemID: capParent.ID}
	for _, i := range []*dbmodel.Item{limited, plain, capItem} {
		if err := i.Insert(ctx, db, boil.Infer()); err != nil {
			t.Fatalf("insert item %s: %v", i.ID, err)
		}
	}
	if err := limited.AddItemPhotos(ctx, db, true,
		&dbmodel.ItemPhoto{ID: "ph-back", URL: "https://example.com/back.jpg", Position: null.IntFrom(1)},
		&dbmodel.ItemPhoto{ID: "ph-front", URL: "https://example.com/front.jpg", Position: null.IntFrom(0)},
	); err != nil {
		t.Fatalf("insert photos: %v", err)
	}
	if err := limited.AddRelatedItems(ctx, db, false, capItem); err != nil {
		t.Fatalf("relate items: %v", err)
	}

	got, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	want := []models.Item{
		{
			// No description, care guide, tailoring or sizes anywhere: zero
			// values and empty lists rather than nulls.
			ID:                   "i-cap-1",
			Name:                 "Cap",
			ParentItemID:         "p-cap",
			Category:             "accessories",
			Material:             "wool",
			Colour:               "navy",
			Price:                1500,
			Sizes:                []string{},
			PhotoURLs:            []string{},
			RelatedItemIDs:       []string{},
			RelatedParentItemIDs: []string{},
		},
		{
			// Own description and tailoring win; care guide falls back.
			ID:                   "i-tee-1",
			Name:                 "Tee, limited",
			ParentItemID:         "p-tee",
			Category:             "tops",
			Description:          "Own description",
			CareGuide:            "Parent care guide",
			Material:             "cotton",
			Colour:               "white",
			TailoringID:          "slim",
			Price:                2500,
			Sizes:                []string{"S", "M", "L"},
			PhotoURLs:            []string{"https://example.com/front.jpg", "https://example.com/back.jpg"},
			ReleaseDate:          "2026-10-01",
			LimitedEdition:       &models.LimitedEdition{Max: 100, Available: 42},
			RelatedItemIDs:       []string{"i-cap-1"},
			RelatedParentItemIDs: []string{"p-cap"},
		},
		{
			// Everything inherited from the parent item.
			ID:                   "i-tee-2",
			Name:                 "Tee",
			ParentItemID:         "p-tee",
			Category:             "tops",
			Description:          "Parent description",
			CareGuide:            "Parent care guide",
			Material:             "cotton",
			Colour:               "white",
			TailoringID:          "normal",
			Price:                2500,
			Sizes:                []string{"S", "M", "L"},
			PhotoURLs:            []string{},
			RelatedItemIDs:       []string{},
			RelatedParentItemIDs: []string{"p-cap"},
		},
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d items, got %d: %+v", len(want), len(got), got)
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("item %d:\n got: %+v\nwant: %+v", i, got[i], want[i])
		}
	}
}
