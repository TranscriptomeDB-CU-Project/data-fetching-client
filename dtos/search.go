package dtos

type SearchResult struct {
	TotalHits int     `json:"totalHits"`
	Page      int     `json:"page"`
	PageSize  int     `json:"pageSize"`
	Hits      []Study `json:"hits"`
}

type FileMetadata struct {
	Name string `json:"Name"`
	Type string `json:"Type"`
}

type SearchFileResult struct {
	Draw            int            `json:"draw"`
	RecordsTotal    int            `json:"recordsTotal"`
	RecordsFiltered int            `json:"recordsFiltered"`
	Files           []FileMetadata `json:"data"`
}
