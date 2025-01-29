package dto

type SidraRequest struct {
	Headers  map[string]string
	Body     string
	Url      string
	Method   string
	Upstream string
	Plugins  []string
	IP       string `json:"ip"`
}

type SidraResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}
