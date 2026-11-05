package entity

import (
	"time"

	"github.com/desain-gratis/common/delivery/mycontent-api/mycontent"
)

var _ mycontent.Data = &Product{}

// Product a simple product data~
type Product struct {
	Ns string `json:"namespace"`

	Id string `json:"id"`

	Name        string `json:"name"`
	Description string `json:"description"`

	Price    string `json:"price"`
	Currency string `json:"currency"`

	PublishedAt time.Time `json:"published_at"`
	URLx        string    `json:"url"`
}

func (a *Product) CreatedTime() time.Time {
	return a.PublishedAt
}

func (a *Product) ID() string {
	return a.Id
}

func (a *Product) Namespace() string {
	return a.Ns
}

func (a *Product) RefIDs() []string {
	return nil
}

func (a *Product) URL() string {
	return a.URLx
}

func (a *Product) Validate() error {
	return nil
}

func (a *Product) WithCreatedTime(t time.Time) mycontent.Data {
	a.PublishedAt = t
	return a
}

func (a *Product) WithID(id string) mycontent.Data {
	a.Id = id
	return a
}

func (a *Product) WithNamespace(id string) mycontent.Data {
	a.Ns = id
	return a
}

func (a *Product) WithURL(url string) mycontent.Data {
	a.URLx = url
	return a
}

func (a *Product) WithVersion(ver uint64) mycontent.Data {
	return a
}

func (a *Product) DGVersion() *uint64 {
	return nil
}
