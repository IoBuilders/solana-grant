package pagination

type PaginatedQueryResponse[K interface{}] struct {
	Items         []K
	TotalElements int
	PaginationParams
}
