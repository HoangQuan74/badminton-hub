package model

import (
	"strings"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
)

type ListQuery struct {
	Limit       int
	Offset      int
	Search      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

func NewListQuery(limit, offset int, search, createdFrom, createdTo string) (ListQuery, error) {
	// Apply shared pagination defaults and bounds.
	if limit <= 0 {
		limit = constant.DefaultPageSize
	}
	if limit > constant.MaxPageSize || offset < 0 {
		return ListQuery{}, identityException.ErrInvalidListQuery
	}

	// Parse optional date filters using RFC3339.
	from, err := parseOptionalTime(createdFrom)
	if err != nil {
		return ListQuery{}, identityException.ErrInvalidListQuery
	}
	to, err := parseOptionalTime(createdTo)
	if err != nil {
		return ListQuery{}, identityException.ErrInvalidListQuery
	}
	if from != nil && to != nil && from.After(*to) {
		return ListQuery{}, identityException.ErrInvalidListQuery
	}

	return ListQuery{
		Limit:       limit,
		Offset:      offset,
		Search:      strings.TrimSpace(search),
		CreatedFrom: from,
		CreatedTo:   to,
	}, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
