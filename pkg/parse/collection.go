package parse

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/types"
)

func QueryOptions(apiOp *types.APIRequest, schema *types.Schema) types.QueryOptions {
	req := apiOp.Request
	if req.Method != http.MethodGet {
		return types.QueryOptions{}
	}

	result := &types.QueryOptions{
		Options: map[string]string{},
	}

	result.Sort = parseSort(schema, apiOp)
	result.Pagination = parsePagination(apiOp)
	result.Conditions = parseFilters(schema, apiOp)

	return *result
}

func parseOrder(apiOp *types.APIRequest) types.SortOrder {
	order := apiOp.Query.Get("order")
	if types.SortOrder(order) == types.DESC {
		return types.DESC
	}
	return types.ASC
}

func parseSort(schema *types.Schema, apiOp *types.APIRequest) types.Sort {
	sortField := apiOp.Query.Get("sort")
	if _, ok := schema.CollectionFilters[sortField]; !ok {
		sortField = ""
	}
	return types.Sort{
		Order: parseOrder(apiOp),
		Name:  sortField,
	}
}

func parsePagination(apiOp *types.APIRequest) *types.Pagination {
	if apiOp.Pagination != nil {
		return apiOp.Pagination
	}

	q := apiOp.Query
	limit := q.Get("limit")
	marker := q.Get("marker")

	result := &types.Pagination{
		Marker: marker,
	}

	if limit != "" {
		limitInt, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return result
		}

		if limitInt >= 0 {
			result.Limit = &limitInt
		}
	}

	return result
}

func parseNameAndOp(value string) (string, types.ModifierType) {
	name := value
	op := "eq"

	idx := strings.LastIndex(value, "_")
	if idx > 0 {
		op = value[idx+1:]
		name = value[0:idx]
	}

	return name, types.ModifierType(op)
}

func parseFilters(schema *types.Schema, apiOp *types.APIRequest) []*types.QueryCondition {
	var conditions []*types.QueryCondition
	for key, values := range apiOp.Query {
		if key == "namespaces" || key == "namespace" {
			continue
		}

		name, op := parseNameAndOp(key)
		filter, ok := schema.CollectionFilters[name]
		if !ok {
			continue
		}

		for _, mod := range filter.Modifiers {
			if op != mod || !types.ValidMod(op) {
				continue
			}

			conditions = append(conditions, types.NewConditionFromString(name, mod, values...))
		}
	}

	return conditions
}
