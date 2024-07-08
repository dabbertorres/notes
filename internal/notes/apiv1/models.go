package apiv1

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/notes"
	"github.com/dabbertorres/notes/internal/tags"
	"github.com/dabbertorres/notes/internal/users"
	"github.com/dabbertorres/notes/internal/util"
)

type Note struct {
	ID        string       `json:"id"`
	CreatedAt string       `json:"created_at"`
	CreatedBy User         `json:"created_by"`
	UpdatedAt string       `json:"updated_at"`
	UpdatedBy User         `json:"updated_by"`
	Title     string       `json:"title,omitempty"`
	Body      string       `json:"body,omitempty"`
	Tags      []Tag        `json:"tags,omitempty"`
	Access    []UserAccess `json:"access,omitempty"`
}

func NoteFromDomain(domain *notes.Note) (n Note) {
	n.ID = domain.ID.String()
	n.CreatedAt = domain.CreatedAt.Format(time.RFC3339)
	n.CreatedBy = UserFromDomain(domain.CreatedBy)
	n.UpdatedAt = domain.UpdatedAt.Format(time.RFC3339)
	n.UpdatedBy = UserFromDomain(domain.UpdatedBy)
	n.Title = domain.Title
	n.Body = domain.Body
	n.Tags = util.MapSlice(domain.Tags, TagFromDomain)
	n.Access = util.MapSlice(domain.Access, UserAccessFromDomain)
	return n
}

func (n *Note) ToDomain() (*notes.Note, error) {
	var errs []error

	out := &notes.Note{
		ID:        apiv1.ValidateOptional(".id", n.ID, &errs, uuid.Parse),
		CreatedAt: apiv1.Validate(".created_at", n.CreatedAt, &errs, apiv1.ParseRFC3339),
		CreatedBy: apiv1.Validate(".created_by", n.CreatedBy, &errs, User.ToDomain),
		UpdatedAt: apiv1.Validate(".updated_at", n.UpdatedAt, &errs, apiv1.ParseRFC3339),
		UpdatedBy: apiv1.Validate(".updated_by", n.UpdatedBy, &errs, User.ToDomain),
		Title:     n.Title,
		Body:      n.Body,
		Tags:      util.MapSlice(n.Tags, Tag.ToDomain),
		Access:    apiv1.ValidateSlice(".access", n.Access, &errs, UserAccess.ToDomain),
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return out, nil
}

// WritableNote contains only the subset of fields on [Note] that an API user can modify.
type WritableNote struct {
	Title  string       `json:"title,omitempty"`
	Body   string       `json:"body,omitempty"`
	Tags   []Tag        `json:"tags,omitempty"`
	Access []UserAccess `json:"access,omitempty"`
}

func (n *WritableNote) ToDomain() (*notes.Note, error) {
	var errs []error

	out := &notes.Note{
		Title:  n.Title,
		Body:   n.Body,
		Tags:   util.MapSlice(n.Tags, Tag.ToDomain),
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
		Access: apiv1.ValidateOptional(".access", a.Access, &errs, users.ParseAccessLevel),
	}

	if len(errs) != 0 {
		return users.Access{}, errors.Join(errs...)
	}

	return out, nil
}

type Tag struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func TagFromDomain(domain tags.Tag) (t Tag) {
	t.ID = domain.ID
	t.Name = domain.Name
	return t
}

func (t Tag) ToDomain() tags.Tag {
	return tags.Tag{
		ID:   t.ID,
		Name: t.Name,
	}
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

type ListNotesPageTokenData struct {
	LastNoteID uuid.NullUUID
	LastRank   float32
	TextSearch string
	TagSearch  uuid.NullUUID
}

func (d *ListNotesPageTokenData) EncodePager() ([][]byte, error) {
	if d == nil {
		return nil, nil
	}

	var out [4][]byte

	if d.LastNoteID.Valid {
		out[0] = []byte(d.LastNoteID.UUID.String())
	}

	if d.LastRank > 0.0 {
		out[1] = strconv.AppendFloat(out[1], float64(d.LastRank), 'g', -1, 32)
	}

	if d.TextSearch != "" {
		out[2] = []byte(d.TextSearch)
	}

	if d.TagSearch.Valid {
		out[3] = []byte(d.TagSearch.UUID.String())
	}

	return out[:], nil
}

func (t *ListNotesPageTokenData) DecodePager(data [][]byte) (err error) {
	if len(data) != 4 {
		return errors.New("invalid page token format (incorrect number of parts)")
	}

	if len(data[0]) != 0 {
		lastNoteID, err := uuid.ParseBytes(data[0])
		if err != nil {
			return err
		}

		t.LastNoteID.UUID = lastNoteID
		t.LastNoteID.Valid = true
	}

	if len(data[1]) != 0 {
		lastRank, err := strconv.ParseFloat(string(data[1]), 32)
		if err != nil {
			return err
		}

		t.LastRank = float32(lastRank)
	}

	if len(data[2]) != 0 {
		t.TextSearch = string(data[2])
	}

	if len(data[3]) != 0 {
		tagSearch, err := uuid.ParseBytes(data[3])
		if err != nil {
			return err
		}

		t.TagSearch.UUID = tagSearch
		t.TagSearch.Valid = true
	}

	return nil
}
