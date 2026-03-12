package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const openRouterChatCompletionsURL = "https://openrouter.ai/api/v1/chat/completions"
const openRouterModel = "google/gemini-2.5-flash"
const maxPromptRows = 200

var forbiddenSQLPattern = regexp.MustCompile(`(?i)\b(insert|update|delete|drop|alter|create|truncate|grant|revoke|copy|call|do|execute|prepare|deallocate|vacuum|analyze|refresh|set|show|begin|commit|rollback|lock)\b`)
var tablePattern = regexp.MustCompile(`(?i)\b(from|join)\s+([a-zA-Z0-9_\.\"]+)`)

// @Summary      Prompt SQL Assistant
// @Description  Generates a read-only SQL query from a natural language prompt, executes it, then returns a summary
// @Tags         Stations
// @Accept       json
// @Produce      json
// @Param        body  body      model.PromptRequest  true  "Prompt Request"
// @Success      200   {object}  model.PromptResponse
// @Failure      400   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /prompt [post]
func (environment *Environment) PromptHandler(context *gin.Context) {
	var request model.PromptRequest
	if bindError := context.ShouldBindJSON(&request); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	request.Prompt = strings.TrimSpace(request.Prompt)
	if request.Prompt == "" {
		api.SendErrorResponse(context, http.StatusBadRequest, "Prompt Is Required")
		return
	}

	if environment.OpenRouterAPIKey == "" {
		api.SendErrorResponse(context, http.StatusInternalServerError, "OpenRouter API Key Not Configured")
		return
	}

	sqlQuery, generationError := environment.generateSQLFromPrompt(context, request.Prompt)
	if generationError != nil {
		api.SendErrorResponse(context, http.StatusBadGateway, "Failed To Generate SQL Query: "+generationError.Error())
		return
	}

	sqlQuery = sanitizeGeneratedSQL(sqlQuery)
	if ok := isReadOnlyAllowedSQL(sqlQuery); !ok {
		api.SendErrorResponse(context, http.StatusBadRequest, "Generated SQL Query Is Not Allowed")
		return
	}

	rows, queryError := environment.Database.Query(context, sqlQuery)
	if queryError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Execute Generated SQL Query")
		return
	}
	defer rows.Close()

	queryRows, parseError := rowsToMaps(rows, maxPromptRows)
	if parseError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Query Result")
		return
	}

	resultJSON, marshalError := json.Marshal(queryRows)
	if marshalError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Serialize Query Result")
		return
	}

	summary, summaryError := environment.generateSummaryFromResult(context, request.Prompt, sqlQuery, string(resultJSON))
	if summaryError != nil {
		summary = "Query executed successfully, but summary generation failed."
	}

	context.JSON(http.StatusOK, model.PromptResponse{
		SQL:     sqlQuery,
		Summary: summary,
		Rows:    queryRows,
	})
}

func sanitizeGeneratedSQL(query string) string {
	trimmed := strings.TrimSpace(query)
	trimmed = strings.TrimPrefix(trimmed, "```sql")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.TrimSuffix(trimmed, ";")
	return strings.TrimSpace(trimmed)
}

func isReadOnlyAllowedSQL(query string) bool {
	normalized := strings.TrimSpace(strings.ToLower(query))
	if normalized == "" {
		return false
	}

	if strings.Contains(normalized, ";") {
		return false
	}

	if !(strings.HasPrefix(normalized, "select") || strings.HasPrefix(normalized, "with")) {
		return false
	}

	if forbiddenSQLPattern.MatchString(normalized) {
		return false
	}

	matches := tablePattern.FindAllStringSubmatch(normalized, -1)
	if len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		tableName := strings.Trim(match[2], `"`)
		if strings.Contains(tableName, ".") {
			parts := strings.Split(tableName, ".")
			tableName = parts[len(parts)-1]
		}

		if tableName != "stations" && tableName != "stations_analysis" {
			return false
		}
	}

	return true
}

func rowsToMaps(rows pgx.Rows, limit int) ([]map[string]interface{}, error) {
	fieldDescriptions := rows.FieldDescriptions()
	rowsAsMaps := make([]map[string]interface{}, 0)

	for rows.Next() {
		values, valuesError := rows.Values()
		if valuesError != nil {
			return nil, valuesError
		}

		rowData := make(map[string]interface{}, len(fieldDescriptions))
		for index, field := range fieldDescriptions {
			fieldName := string(field.Name)
			if index >= len(values) {
				rowData[fieldName] = nil
				continue
			}

			if bytesValue, ok := values[index].([]byte); ok {
				rowData[fieldName] = string(bytesValue)
			} else {
				rowData[fieldName] = values[index]
			}
		}

		rowsAsMaps = append(rowsAsMaps, rowData)
		if len(rowsAsMaps) >= limit {
			break
		}
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, rowsError
	}

	return rowsAsMaps, nil
}

func (environment *Environment) generateSQLFromPrompt(context *gin.Context, prompt string) (string, error) {
	systemPrompt := "You are a PostgreSQL SQL assistant. Generate exactly one read-only SQL query for the user's request. Only query from tables stations and stations_analysis. Never use INSERT, UPDATE, DELETE, DROP, ALTER, CREATE, TRUNCATE, GRANT, REVOKE, COPY, CALL, DO, EXECUTE, PREPARE, DEALLOCATE, VACUUM, ANALYZE, REFRESH, SET, SHOW, BEGIN, COMMIT, ROLLBACK or LOCK. Return SQL only, no markdown or explanation. Prefer including LIMIT 200 or less when many rows may be returned. Schema: stations(tiploc, nlc, name, crs, latitude, longitude); stations_analysis(tiploc, service_count, delay_average_commute, delay_rank_commute, delay_average, delay_rank)."
	return environment.openRouterChatCompletion(context, systemPrompt, prompt)
}

func (environment *Environment) generateSummaryFromResult(context *gin.Context, prompt string, sqlQuery string, resultJSON string) (string, error) {
	systemPrompt := "You summarize SQL query results for API responses. Keep the summary concise and factual in plain text."
	userPrompt := fmt.Sprintf("User prompt: %s\n\nSQL query: %s\n\nRows JSON: %s\n\nReturn a concise summary.", prompt, sqlQuery, resultJSON)
	return environment.openRouterChatCompletion(context, systemPrompt, userPrompt)
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

type openRouterResponse struct {
	Choices []struct {
		Message openRouterMessage `json:"message"`
	} `json:"choices"`
}

func (environment *Environment) openRouterChatCompletion(context *gin.Context, systemPrompt string, userPrompt string) (string, error) {
	requestBody := openRouterRequest{
		Model: openRouterModel,
		Messages: []openRouterMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	encodedBody, marshalError := json.Marshal(requestBody)
	if marshalError != nil {
		return "", marshalError
	}

	httpRequest, requestError := http.NewRequestWithContext(context.Request.Context(), http.MethodPost, openRouterChatCompletionsURL, bytes.NewReader(encodedBody))
	if requestError != nil {
		return "", requestError
	}

	httpRequest.Header.Set("Authorization", "Bearer "+environment.OpenRouterAPIKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, doError := http.DefaultClient.Do(httpRequest)
	if doError != nil {
		return "", doError
	}

	bodyBytes, readError := io.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if readError != nil {
		return "", readError
	}

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return "", fmt.Errorf("openrouter model %s failed with status %d: %s", openRouterModel, httpResponse.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var response openRouterResponse
	if unmarshalError := json.Unmarshal(bodyBytes, &response); unmarshalError != nil {
		return "", unmarshalError
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("openrouter model %s returned no choices", openRouterModel)
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}
