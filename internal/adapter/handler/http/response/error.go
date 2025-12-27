package response

// ErrorResponse represent the standard JSON representation of the API error.
type ErrorResponse struct {
	Code     string   `json:"code"`
	Messages []string `json:"messages"`
}
