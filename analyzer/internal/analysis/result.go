package analysis

type LayerResult struct {
	Digest    string `json:"digest"`
	SizeBytes int64  `json:"size_bytes"`
	MediaType string `json:"media_type"`
}

type Result struct {
	Image           string           `json:"image"`
	Tag             string           `json:"tag"`
	MediaType       string           `json:"media_type"`
	Layers          []LayerResult    `json:"layers"`
	TotalSizeBytes  int64            `json:"total_size_bytes"`
	Insights        Insights         `json:"insights"`
	Recommendations []Recommendation `json:"recommendations"`
}
