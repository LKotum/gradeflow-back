package response

// PageMeta provides pagination metadata for list responses.
type PageMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// Paginated wraps items with pagination metadata.
type Paginated[T any] struct {
	Data []T     `json:"data"`
	Meta PageMeta `json:"meta"`
}
