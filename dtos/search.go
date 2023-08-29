package dtos

type SearchResult struct {
	TotalHits int     `json:"totalHits"`
	Page      int     `json:"page"`
	PageSize  int     `json:"pageSize"`
	Hits      []Study `json:"hits"`
}
