package dtos

type Study struct {
	Accession    string `json:"accession"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Links        int    `json:"links"`
	Files        int    `json:"files"`
	Release_date string `json:"release_date"`
	Views        int    `json:"views"`
	IsPublic     bool   `json:"isPublic"`
}
