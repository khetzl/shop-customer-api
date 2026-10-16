package models

// Item is a sellable article as shown to customers, with the properties it
// inherits from its parent item already resolved.
type Item struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ParentItemID string `json:"parentItemId"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	CareGuide    string `json:"careGuide"`
	Material     string `json:"material"`
	Colour       string `json:"colour"`
	// TailoringID is the item's fit (e.g. "slim"); empty if it has none.
	TailoringID string `json:"tailoringId"`
	// Price is in pennies.
	Price       int      `json:"price"`
	Sizes       []string `json:"sizes"`
	PhotoURLs   []string `json:"photoUrls"`
	ReleaseDate string   `json:"releaseDate"`
	// LimitedEdition is nil unless the item is a limited edition.
	LimitedEdition       *LimitedEdition `json:"limitedEdition,omitempty"`
	RelatedItemIDs       []string        `json:"relatedItemIds"`
	RelatedParentItemIDs []string        `json:"relatedParentItemIds"`
}

type LimitedEdition struct {
	Max       int `json:"max"`
	Available int `json:"available"`
}
