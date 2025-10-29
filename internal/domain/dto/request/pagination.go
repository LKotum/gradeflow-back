package request

// PaginationQuery captures pagination/filter parameters for list endpoints.
type PaginationQuery struct {
	Limit  int     `form:"limit,default=20"`
	Offset int     `form:"offset,default=0"`
	Search *string `form:"search"`
}

// Normalize bounds pagination values to sane defaults.
func (p *PaginationQuery) Normalize(maxLimit int) {
	if maxLimit <= 0 {
		maxLimit = 100
	}
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > maxLimit {
		p.Limit = maxLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}
