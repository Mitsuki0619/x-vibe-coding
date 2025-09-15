package handlers

// GraphQLResponse represents a GraphQL response structure
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error structure
type GraphQLError struct {
	Message string `json:"message"`
}

// ErrorResponse creates a GraphQL error response
func ErrorResponse(message string) GraphQLResponse {
	return GraphQLResponse{
		Errors: []GraphQLError{{Message: message}},
	}
}

// DataResponse creates a GraphQL data response
func DataResponse(key string, data interface{}) GraphQLResponse {
	return GraphQLResponse{
		Data: map[string]interface{}{
			key: data,
		},
	}
}

// getString extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getUint extracts a uint value from a map
func getUint(m map[string]interface{}, key string) uint {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return uint(num)
		}
		if num, ok := val.(int); ok {
			return uint(num)
		}
	}
	return 0
}