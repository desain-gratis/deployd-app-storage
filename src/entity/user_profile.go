package entity

import (
	"time"

	"github.com/desain-gratis/common/delivery/mycontent-api/mycontent"
)

var _ mycontent.Data = &UserProfile{}

// UserProfile represents raft UserProfile configuration for a particular service
type UserProfile struct {
	Ns string `json:"namespace"`

	Id string `json:"id"`

	Name     string `json:"name"`
	NickName string `json:"nick_name"`

	Hobby []string `json:"hobby"`

	PublishedAt time.Time `json:"published_at"`
	URLx        string    `json:"url"`
}

func (a *UserProfile) CreatedTime() time.Time {
	return a.PublishedAt
}

func (a *UserProfile) ID() string {
	return a.Id
}

func (a *UserProfile) Namespace() string {
	return a.Ns
}

func (a *UserProfile) RefIDs() []string {
	return nil
}

func (a *UserProfile) URL() string {
	return a.URLx
}

func (a *UserProfile) Validate() error {
	return nil
}

func (a *UserProfile) WithCreatedTime(t time.Time) mycontent.Data {
	a.PublishedAt = t
	return a
}

func (a *UserProfile) WithID(id string) mycontent.Data {
	a.Id = id
	return a
}

func (a *UserProfile) WithNamespace(id string) mycontent.Data {
	a.Ns = id
	return a
}

func (a *UserProfile) WithURL(url string) mycontent.Data {
	a.URLx = url
	return a
}

func (a *UserProfile) WithVersion(ver uint64) mycontent.Data {
	return a
}

func (a *UserProfile) DGVersion() *uint64 {
	return nil
}
