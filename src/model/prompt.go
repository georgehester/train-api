package model

type PromptRequest struct {
	Prompt string `json:"prompt"`
} // @name PromptRequest

type PromptResponse struct {
	SQL     string                   `json:"sql"`
	Summary string                   `json:"summary"`
	Rows    []map[string]interface{} `json:"rows"`
} // @name PromptResponse
