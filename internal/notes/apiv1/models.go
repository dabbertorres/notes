package apiv1

import (
	"time"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/notes"
	"github.com/dabbertorres/notes/internal/users"
	"github.com/dabbertorres/notes/internal/util"
)

type Note struct {
	ID        uuid.UUID    `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	CreatedBy User         `json:"created_by"`
	UpdatedAt time.Time    `json:"updated_at"`
	UpdatedBy User         `json:"updated_by"`
	Title     string       `json:"title,omitempty"`
	Body      string       `json:"body,omitempty"`
	Tags      []Tag        `json:"tags,omitempty"`
	Access    []UserAccess `json:"access,omitempty"`
}

func (n *Note) FromDomain(domain *notes.Note) {
	n.ID = domain.ID
	n.CreatedAt = domain.CreatedAt
	n.CreatedBy.FromDomain(domain.CreatedBy)
	n.UpdatedAt = domain.UpdatedAt
	n.UpdatedBy.FromDomain(domain.UpdatedBy)
	n.Title = domain.Title
	n.Body = domain.Body
	n.Tags = util.MapSlice(domain.Tags, func(t notes.Tag) (out Tag) {
		out.FromDomain(t)
		return out
	})
	n.Access = util.MapSlice(domain.Access, func(a notes.UserAccess) (out UserAccess) {
		out.FromDomain(a)
		return out
	})
}

func (n *Note) ToDomain() *notes.Note {
	return &notes.Note{
		ID:        n.ID,
		CreatedAt: n.CreatedAt,
		CreatedBy: n.CreatedBy.ToDomain(),
		UpdatedAt: n.UpdatedAt,
		UpdatedBy: n.UpdatedBy.ToDomain(),
		Title:     n.Title,
		Body:      n.Body,
		Tags:      util.MapSlice(n.Tags, Tag.ToDomain),
		Access:    util.MapSlice(n.Access, UserAccess.ToDomain),
	}
}

type UserAccess struct {
	User   User   `json:"user"`
	Access string `json:"access"`
}

func (a *UserAccess) FromDomain(domain notes.UserAccess) {
	a.User.FromDomain(domain.User)
	a.Access = string(domain.Access)
}

func (a UserAccess) ToDomain() notes.UserAccess {
	return notes.UserAccess{
		User:   a.User.ToDomain(),
		Access: notes.AccessLevel(a.Access),
	}
}

type Tag struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (t *Tag) FromDomain(domain notes.Tag) {
	t.ID = domain.ID
	t.Name = domain.Name
}

func (t Tag) ToDomain() notes.Tag {
	return notes.Tag{
		ID:   t.ID,
		Name: t.Name,
	}
}

type User struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Active bool      `json:"active"`
}

func (u *User) FromDomain(domain users.User) {
	u.ID = domain.ID
	u.Name = domain.Name
	u.Active = domain.Active
}

func (u User) ToDomain() users.User {
	return users.User{
		ID:     u.ID,
		Name:   u.Name,
		Active: u.Active,
	}
}
