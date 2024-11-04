package lib

type SidraRequest struct {
	Headers map[string]string
	Body string
	Url string
	Method string
}

type SidraResponse struct {
	StatusCode int
	Headers map[string]string
	Body string
}