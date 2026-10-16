package itemstore

import (
	"context"
	"database/sql"

	"github.com/askthechain/atcutils/pkg/logger"
	dbmodel "github.com/khetzl/shop-db-models/pkg/storage/dbmodel"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/khetzl/shop-customer-api/internal/domain/models"
)

// Service loads items through the shop-db-models library. Items are read from
// the item_details view, which already resolves the properties an item
// inherits from its parent item (category, material, sizes, price, colour and
// the description/care guide/tailoring fallbacks) and lists its photos and
// related items.
type Service struct {
	logger atclogger.Logger
	db     *sql.DB
}

func New(logger atclogger.Logger, db *sql.DB) (*Service, error) {
	return &Service{
		logger: logger,
		db:     db,
	}, nil
}

// GetAll returns every item, ordered by id.
func (s *Service) GetAll() ([]models.Item, error) {
	ctx := context.Background()

	rows, err := dbmodel.ItemDetails(qm.OrderBy(dbmodel.ItemDetailColumns.ID)).All(ctx, s.db)
	if err != nil {
		return nil, err
	}

	out := make([]models.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, toModel(row))
	}
	return out, nil
}

// toModel maps an item_details row onto the API item model. The view's
// columns are all nullable as far as SQLBoiler knows, so NULLs become zero
// values and NULL arrays become empty lists.
func toModel(d *dbmodel.ItemDetail) models.Item {
	out := models.Item{
		ID:                   d.ID.String,
		Name:                 d.Name.String,
		ParentItemID:         d.ParentItemID.String,
		Category:             d.Category.String,
		Description:          d.Description.String,
		CareGuide:            d.CareGuide.String,
		Material:             d.Material.String,
		Colour:               d.Colour.String,
		TailoringID:          d.TailoringID.String,
		Price:                d.BasePrice.Int,
		Sizes:                nonNil(d.AvailableSizes),
		PhotoURLs:            nonNil(d.PhotoUrls),
		RelatedItemIDs:       nonNil(d.RelatedItemIds),
		RelatedParentItemIDs: nonNil(d.RelatedParentItemIds),
	}
	if d.ReleaseDate.Valid {
		out.ReleaseDate = d.ReleaseDate.Time.Format("2006-01-02")
	}
	if d.LimitedEditionMax.Valid {
		out.LimitedEdition = &models.LimitedEdition{
			Max:       d.LimitedEditionMax.Int,
			Available: d.LimitedEditionAvailable.Int,
		}
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
