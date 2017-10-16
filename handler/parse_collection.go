package handler

import (
	"net/http"

	"strconv"

	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/norman/query"
)

var (
	ASC          = SortOrder("asc")
	DESC         = SortOrder("desc")
	defaultLimit = 100
	maxLimit     = 3000
)

type SortOrder string

type Pagination struct {
	Limit  int
	Marker string
}

type CollectionAttributes struct {
	Sort       string
	Order      SortOrder
	Pagination *Pagination
	Conditions []*query.Condition
}

func ParseCollectionAttributes(req *http.Request, schema client.Schema) *CollectionAttributes {
	if req.Method != http.MethodGet {
		return nil
	}

	result := &CollectionAttributes{}

	result.Order = parseOrder(req)
	result.Sort = parseSort(schema, req)
	result.Pagination = parsePagination(req)
	result.Conditions = parseFilters(schema, req)

	return result
}

func parseOrder(req *http.Request) SortOrder {
	order := req.URL.Query().Get("order")
	if SortOrder(order) == DESC {
		return DESC
	}
	return ASC
}

func parseSort(schema client.Schema, req *http.Request) string {
	sort := req.URL.Query().Get("sort")
	if _, ok := schema.CollectionFilters[sort]; ok {
		return sort
	}
	return ""
}

func parsePagination(req *http.Request) *Pagination {
	q := req.URL.Query()
	limit := q.Get("limit")
	marker := q.Get("marker")

	result := &Pagination{
		Limit:  defaultLimit,
		Marker: marker,
	}

	if limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			return result
		}

		if limitInt > maxLimit {
			result.Limit = maxLimit
		} else if limitInt > 0 {
			result.Limit = limitInt
		}
	}

	return result
}

func parseNameAndOp(value string) (string, string) {
	name := value
	op := "eq"

	idx := strings.LastIndex(value, "_")
	if idx > 0 {
		op = value[idx+1:]
		name = value[0:idx]
	}

	return name, op
}

func parseFilters(schema client.Schema, req *http.Request) []*query.Condition {
	conditions := []*query.Condition{}
	q := req.URL.Query()
	for key, values := range req.URL.Query() {
		name, op := parseNameAndOp(key)
		filter, ok := schema.CollectionFilters[name]
		if !ok {
			continue
		}

		for _, mod := range filter.Modifiers {
			if op != mod || !query.ValidMod(op) {
				continue
			}

			genericValues := []interface{}{}
			for _, value := range values {
				genericValues = append(genericValues, value)
			}

			conditions = append(conditions, query.NewCondition(query.ConditionType(mod), genericValues))
		}
	}

	return conditions
}
