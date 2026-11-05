package entity

import (
	"time"

	"github.com/desain-gratis/common/delivery/mycontent-api/mycontent"
)

var _ mycontent.Data = &ProductReview{}

// ProductReview a simple ProductReview data~
type ProductReview struct {
	Ns string `json:"namespace"`

	Id string `json:"id"`

	ProductID string `json:"product_id"`

	Rating int    `json:"rating"`
	Review string `json:"review"`

	Reviewer string `json:"reviewer"`

	PublishedAt time.Time `json:"published_at"`
	URLx        string    `json:"url"`
}

func (a *ProductReview) CreatedTime() time.Time {
	return a.PublishedAt
}

func (a *ProductReview) ID() string {
	return a.Id
}

func (a *ProductReview) Namespace() string {
	return a.Ns
}

func (a *ProductReview) RefIDs() []string {
	// references the product ID
	return []string{a.ProductID}
}

func (a *ProductReview) URL() string {
	return a.URLx
}

func (a *ProductReview) Validate() error {
	return nil
}

func (a *ProductReview) WithCreatedTime(t time.Time) mycontent.Data {
	a.PublishedAt = t
	return a
}

func (a *ProductReview) WithID(id string) mycontent.Data {
	a.Id = id
	return a
}

func (a *ProductReview) WithNamespace(id string) mycontent.Data {
	a.Ns = id
	return a
}

func (a *ProductReview) WithURL(url string) mycontent.Data {
	a.URLx = url
	return a
}

func (a *ProductReview) WithVersion(ver uint64) mycontent.Data {
	return a
}

func (a *ProductReview) DGVersion() *uint64 {
	return nil
}
