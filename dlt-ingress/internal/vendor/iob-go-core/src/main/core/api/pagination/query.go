package pagination

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

type PaginationParams struct {
	PageSize int
	Offset   int
	SortBy   string
	SortDir  string
}

func (p PaginationParams) OrderClause() string {
	sortBy := p.SortBy
	sortDir := strings.ToUpper(p.SortDir)

	if sortBy == "" {
		sortBy = "created_at"
	}

	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}

	sortBy = camelToSnake(sortBy)

	return sortBy + " " + sortDir
}

func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// GetPaginationParams extracts pagination parameters `pageSize`, `offset`, `sortBy` and `sortDir` from the query parameters of the HTTP request context.
// - If `size`, `page`, `sortBy` or `sortDirection` query parameters are not provided, it defaults to 10 for `pageSize`, 0 for `offset`, "createdAt" for `sortBy` and "DESC" for `sortDir`.
// - Returns an error if the provided `pageSize`, `offset`, `sortBy` or `sortDirection` values are not valid integers.
func GetPaginationParams(c *gin.Context) (PaginationParams, error) {
	params := PaginationParams{
		PageSize: 10,
		Offset:   0,
		SortBy:   "created_at",
		SortDir:  "DESC",
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		parsedLimit, err := strconv.Atoi(sizeStr)
		if err != nil {
			return params, err
		}
		params.PageSize = parsedLimit
	}

	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil {
			return params, err
		}
		params.Offset = params.PageSize * parsedPage
	}

	if sortBy := c.Query("sortBy"); sortBy != "" {
		params.SortBy = camelToSnake(sortBy)
	}

	// sort direction (sortDirection=asc|desc)
	if sortDirection := c.Query("sortDirection"); sortDirection != "" {
		sortDirection = strings.ToUpper(sortDirection)

		if sortDirection != "ASC" && sortDirection != "DESC" {
			sortDirection = "DESC"
		}
		params.SortDir = sortDirection
	}

	return params, nil
}
