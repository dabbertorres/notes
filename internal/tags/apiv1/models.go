package apiv1

import (
	"errors"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/tags"
	"github.com/dabbertorres/notes/internal/users"
	"github.com/dabbertorres/notes/internal/util"
)

type Tag struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Access []UserAccess `json:"access,omitempty"`
}

func TagFromDomain(domain *tags.Tag) (t Tag) {
	t.ID = domain.ID.String()
	t.Name = domain.Name
	t.Access = util.MapSlice(domain.Access, UserAccessFromDomain)
	return t
}

func (t *Tag) ToDomain() (*tags.Tag, error) {
	var errs []error

	out := &tags.Tag{
		ID:     apiv1.Validate(".id", t.ID, &errs, uuid.Parse),
		Name:   t.Name,
		Access: apiv1.ValidateSlice(".access", t.Access, &errs, UserAccess.ToDomain),
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return out, nil
}

// WritableTag contains only the subset of fields on [Tag] that an API user can modify.
type WritableTag struct {
	Name   string       `json:"name,omitempty"`
	Access []UserAccess `json:"access,omitempty"`
}

func (n *WritableTag) ToDomain() (*tags.Tag, error) {
	var errs []error

	out := &tags.Tag{
		Name:   n.Name,
		Access: apiv1.ValidateSlice(".access", n.Access, &errs, UserAccess.ToDomain),
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return out, nil
}

type UserAccess struct {
	User   User   `json:"user"`
	Access string `json:"access"`
}

func UserAccessFromDomain(domain users.Access) (a UserAccess) {
	a.User = UserFromDomain(domain.User)
	a.Access = domain.Access.String()
	return a
}

func (a UserAccess) ToDomain() (users.Access, error) {
	var errs []error

	out := users.Access{
		User:   apiv1.Validate(".user", a.User, &errs, User.ToDomain),
		Access: apiv1.Validate(".access", a.Access, &errs, users.ParseAccessLevel),
	}

	if len(errs) != 0 {
		return users.Access{}, errors.Join(errs...)
	}

	return out, nil
}

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func UserFromDomain(domain users.User) (u User) {
	u.ID = domain.ID.String()
	u.Name = domain.Name
	u.Active = domain.Active
	return u
}

func (u User) ToDomain() (users.User, error) {
	var errs []error

	out := users.User{
		ID:     apiv1.Validate(".id", u.ID, &errs, uuid.Parse),
		Name:   u.Name,
		Active: u.Active,
	}

	if len(errs) != 0 {
		return users.User{}, errors.Join(errs...)
	}

	return out, nil
}

type ListTagsPageTokenData struct {
	LastTagID uuid.NullUUID
	Search    string
}

func (d *ListTagsPageTokenData) EncodePager() ([][]byte, error) {
	var out [2][]byte

	if d.LastTagID.Valid {
		out[0] = []byte(d.LastTagID.UUID.String())
	}

	if d.Search != "" {
		out[1] = []byte(d.Search)
	}

	return out[:], nil
}

func (d *ListTagsPageTokenData) DecodePager(data [][]byte) error {
	if len(data) != 2 {
		return errors.New("invalid page token format (incorrect number of parts)")
	}

	if len(data[0]) != 0 {
		lastTagID, err := uuid.ParseBytes(data[0])
		if err != nil {
			return err
		}

		d.LastTagID.UUID = lastTagID
		d.LastTagID.Valid = true
	}

	if len(data[1]) != 0 {
		d.Search = string(data[1])
	}

	return nil
}
