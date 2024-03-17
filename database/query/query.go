package query

import (
	"errors"

	"strings"
)

// QueryType defines the type of query
type QueryType int

const (
	// QueryUnknown represents an unknown query type
	QueryUnknown QueryType = iota
	// QueryInsert represents an insert operation
	QueryInsert
	// QuerySelect represents a select operation
	QuerySelect
)

// Query represents a parsed query
type Query struct {
	Type  QueryType
	Key   string
	Value string // Only used for insert queries
}

// ParseQuery takes a string and attempts to parse it into a Query object
func ParseQuery(queryString string) (*Query, error) {
	tokens := strings.Fields(strings.TrimSpace(queryString))
	if len(tokens) == 0 {
		return nil, errors.New("query is empty")
	}

	switch strings.ToUpper(tokens[0]) {
	case "INSERT":
		if len(tokens) != 3 {
			return nil, errors.New("invalid insert query format")
		}
		return &Query{Type: QueryInsert, Key: tokens[1], Value: tokens[2]}, nil
	case "SELECT":
		if len(tokens) != 2 {
			return nil, errors.New("invalid select query format")
		}
		return &Query{Type: QuerySelect, Key: tokens[1]}, nil
	default:
		return nil, errors.New("unknown query type")
	}
}
