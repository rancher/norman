package handler

import (
	"sort"

	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/norman/pkg/types/convert"
)

func QueryFilter(opts *types.QueryOptions, schema *types.Schema, data types.APIObject) types.APIObject {
	if opts == nil {
		opts = &types.QueryOptions{}
	}
	result := ApplyQueryOptions(opts, schema, data)
	result.ListRevision = data.ListRevision
	return result
}

func ApplyQueryOptions(options *types.QueryOptions, schema *types.Schema, data types.APIObject) types.APIObject {
	data = ApplyQueryConditions(options.Conditions, schema, data)
	data = ApplySort(options.Sort, data)
	return ApplyPagination(options.Pagination, data)
}

func ApplySort(sortOpts types.Sort, data types.APIObject) types.APIObject {
	name := sortOpts.Name
	if name == "" {
		name = "id"
	}

	dataList := data.List()
	sort.Slice(dataList, func(i, j int) bool {
		left, right := i, j
		if sortOpts.Order == types.DESC {
			left, right = j, i
		}

		return convert.ToString(dataList[left][name]) < convert.ToString(dataList[right][name])
	})

	return data
}

func ApplyQueryConditions(conditions []*types.QueryCondition, schema *types.Schema, objs types.APIObject) types.APIObject {
	var result []map[string]interface{}

outer:
	for _, item := range objs.List() {
		for _, condition := range conditions {
			if !condition.Valid(schema, item) {
				continue outer
			}
		}

		result = append(result, item)
	}

	return types.ToAPI(result)
}

func ApplyPagination(pagination *types.Pagination, data types.APIObject) types.APIObject {
	if pagination == nil || pagination.Limit == nil {
		return data
	}

	limit := *pagination.Limit
	if limit < 0 {
		limit = 0
	}

	dataList := data.List()
	total := int64(len(dataList))

	// Reset fields
	pagination.Next = ""
	pagination.Previous = ""
	pagination.Partial = false
	pagination.Total = &total
	pagination.First = ""

	if len(dataList) == 0 {
		return data
	}

	// startIndex is guaranteed to be a valid index
	startIndex := int64(0)
	if pagination.Marker != "" {
		for i, item := range dataList {
			id, _ := item["id"].(string)
			if id == pagination.Marker {
				startIndex = int64(i)
				break
			}
		}
	}

	previousIndex := startIndex - limit
	if previousIndex <= 0 {
		previousIndex = 0
	}
	nextIndex := startIndex + limit
	if nextIndex > int64(len(dataList)) {
		nextIndex = int64(len(dataList))
	}

	if previousIndex < startIndex {
		pagination.Previous, _ = dataList[previousIndex]["id"].(string)
	}

	if nextIndex > startIndex && nextIndex < int64(len(dataList)) {
		pagination.Next, _ = dataList[nextIndex]["id"].(string)
	}

	if startIndex > 0 || nextIndex < int64(len(dataList)) {
		pagination.Partial = true
	}

	if pagination.Partial {
		pagination.First, _ = dataList[0]["id"].(string)

		lastIndex := int64(len(dataList)) - limit
		if lastIndex > 0 && lastIndex < int64(len(dataList)) {
			pagination.Last, _ = dataList[lastIndex]["id"].(string)
		}
	}

	return types.ToAPI(dataList[startIndex:nextIndex])
}
