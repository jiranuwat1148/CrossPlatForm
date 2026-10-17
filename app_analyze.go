package main

type ImageTags struct {
	File  string   `json:"file"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
	Error string   `json:"error,omitempty"`
}

type AnalyzeProgress struct {
	Done  int    `json:"done"`
	Total int    `json:"total"`
	File  string `json:"file"`
}
