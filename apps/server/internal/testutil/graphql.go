package testutil

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"sns-server/internal/server"
)

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// ExecuteGraphQLRequest executes a GraphQL request against the server
func ExecuteGraphQLRequest(t *testing.T, srv *server.Server, req GraphQLRequest) GraphQLResponse {
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	httpReq := httptest.NewRequest("POST", "/query", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	srv.HandleGraphQL(recorder, httpReq)

	var resp GraphQLResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return resp
}

// AssertNoErrors checks that the GraphQL response has no errors
func AssertNoErrors(t *testing.T, resp GraphQLResponse) {
	if resp.Errors != nil {
		t.Errorf("Unexpected errors: %v", resp.Errors)
	}
}

// AssertHasErrors checks that the GraphQL response has errors
func AssertHasErrors(t *testing.T, resp GraphQLResponse) {
	if resp.Errors == nil {
		t.Error("Expected errors but got none")
	}
}

// GetDataMap extracts data as map[string]interface{} from GraphQL response
func GetDataMap(t *testing.T, resp GraphQLResponse) map[string]interface{} {
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data is not a map")
	}
	return data
}

// GetArray extracts an array field from a map
func GetArray(t *testing.T, data map[string]interface{}, field string) []interface{} {
	arr, ok := data[field].([]interface{})
	if !ok {
		t.Fatalf("%s field is not an array", field)
	}
	return arr
}

// GetMap extracts a map field from a map
func GetMap(t *testing.T, data map[string]interface{}, field string) map[string]interface{} {
	m, ok := data[field].(map[string]interface{})
	if !ok {
		t.Fatalf("%s field is not a map", field)
	}
	return m
}

// GetString extracts a string field from a map
func GetString(t *testing.T, data map[string]interface{}, field string) string {
	s, ok := data[field].(string)
	if !ok {
		t.Fatalf("%s field is not a string", field)
	}
	return s
}

// GetBool extracts a boolean field from a map
func GetBool(t *testing.T, data map[string]interface{}, field string) bool {
	b, ok := data[field].(bool)
	if !ok {
		t.Fatalf("%s field is not a boolean", field)
	}
	return b
}
