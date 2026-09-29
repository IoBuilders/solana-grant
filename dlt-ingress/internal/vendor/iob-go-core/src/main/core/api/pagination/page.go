package pagination

import (
	"math"
)

// Page contains pagination metadata for a paginated response.
type Page struct {
	// Number of items returned in the current page
	Size int `json:"size" example:"20"`

	// Current page index (starting from 0)
	Page int `json:"page" example:"0"`

	// Total number of available pages
	TotalPages int `json:"totalPages" example:"5"`

	// Total number of elements matching the query
	TotalElements int `json:"totalElements" example:"93"`
} // @name Page

// PageResponse represents a generic paginated API response.
type PageResponse struct {
	// List of items returned in the current page
	Data any `json:"data"`

	// Pagination metadata
	Page *Page `json:"page"`
} // @name PageResponse

func BuildResponse(data any, pageSize, offset, totalElements int) *PageResponse {
	page := Page{
		Size:          pageSize,
		Page:          offset / pageSize,
		TotalElements: totalElements,
		TotalPages:    int(math.Ceil(float64(totalElements) / float64(pageSize))),
	}

	return &PageResponse{
		Data: data,
		Page: &page,
	}
}

// MapAndBuildResponse maps every item using the mapper function
func MapAndBuildResponse[T any, U any](
	paginatedResp PaginatedQueryResponse[T],
	mapper func(T) U,
) *PageResponse {
	responseItems := make([]U, len(paginatedResp.Items))
	for i, item := range paginatedResp.Items {
		responseItems[i] = mapper(item)
	}

	return BuildResponse(responseItems, paginatedResp.PageSize, paginatedResp.Offset, paginatedResp.TotalElements)
}
