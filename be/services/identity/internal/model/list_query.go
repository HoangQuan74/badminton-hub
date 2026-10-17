package model

import (
	"errors"
	"strings"
	"time"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var ErrInvalidListQuery = errors.New("invalid list query")

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
		limit = DefaultPageSize
	}
	if limit > MaxPageSize || offset < 0 {
		return ListQuery{}, ErrInvalidListQuery
	}

	// Parse optional date filters using RFC3339.
	from, err := parseOptionalTime(createdFrom)
	if err != nil {
		return ListQuery{}, ErrInvalidListQuery
	}
	to, err := parseOptionalTime(createdTo)
	if err != nil {
		return ListQuery{}, ErrInvalidListQuery
	}
	if from != nil && to != nil && from.After(*to) {
		return ListQuery{}, ErrInvalidListQuery
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
