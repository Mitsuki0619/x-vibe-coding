package server_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func TestServerIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := &server.Server{
		DB: db,
	}

	t.Run("ユーザー一覧取得（空の場合）", func(t *testing.T) {
		req := GraphQLRequest{
			Query: `{ users { id username name } }`,
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		users, ok := data["users"].([]interface{})
		if !ok {
			t.Fatal("Users field is not an array")
		}

		if len(users) != 0 {
			t.Errorf("Expected 0 users, got %d", len(users))
		}
	})

	t.Run("ユーザー登録", func(t *testing.T) {
		req := GraphQLRequest{
			Query: `mutation { register(input: { username: "testuser", email: "test@example.com", password: "password123", name: "Test User" }) { token user { id username name email } } }`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"username": "testuser",
					"email":    "test@example.com",
					"password": "password123",
					"name":     "Test User",
				},
			},
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		register, ok := data["register"].(map[string]interface{})
		if !ok {
			t.Fatal("Register field is not a map")
		}

		// トークンの確認
		token, ok := register["token"].(string)
		if !ok || token == "" {
			t.Error("Token is missing or empty")
		}

		// ユーザー情報の確認
		user, ok := register["user"].(map[string]interface{})
		if !ok {
			t.Fatal("User field is not a map")
		}

		if user["username"] != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", user["username"])
		}

		if user["name"] != "Test User" {
			t.Errorf("Expected name 'Test User', got %v", user["name"])
		}
	})

	t.Run("ユーザー一覧取得（登録後）", func(t *testing.T) {
		req := GraphQLRequest{
			Query: `{ users { id username name email } }`,
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		users, ok := data["users"].([]interface{})
		if !ok {
			t.Fatal("Users field is not an array")
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		user, ok := users[0].(map[string]interface{})
		if !ok {
			t.Fatal("First user is not a map")
		}

		if user["username"] != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", user["username"])
		}
	})

	t.Run("投稿作成", func(t *testing.T) {
		req := GraphQLRequest{
			Query: `mutation { createPost(input: { content: "Hello, World!" }) { id content author { username } } }`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"content": "Hello, World!",
				},
			},
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		createPost, ok := data["createPost"].(map[string]interface{})
		if !ok {
			t.Fatal("createPost field is not a map")
		}

		if createPost["content"] != "Hello, World!" {
			t.Errorf("Expected content 'Hello, World!', got %v", createPost["content"])
		}

		author, ok := createPost["author"].(map[string]interface{})
		if !ok {
			t.Fatal("Author field is not a map")
		}

		if author["username"] == "" {
			t.Error("Author username should not be empty")
		}
	})

	t.Run("投稿一覧取得", func(t *testing.T) {
		// 直接データベースにテスト投稿を作成
		user := testutil.CreateTestUser(t, db, "postuser", "post@example.com", "Post User")
		testutil.CreateTestPost(t, db, user.ID, "Test post content")

		req := GraphQLRequest{
			Query: `{ posts { id content author { username } } }`,
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		posts, ok := data["posts"].([]interface{})
		if !ok {
			t.Fatal("Posts field is not an array")
		}

		if len(posts) < 1 {
			t.Errorf("Expected at least 1 post, got %d", len(posts))
		}

		post, ok := posts[0].(map[string]interface{})
		if !ok {
			t.Fatal("First post is not a map")
		}

		if post["content"] != "Test post content" {
			t.Errorf("Expected content 'Test post content', got %v", post["content"])
		}
	})

	t.Run("投稿にいいね", func(t *testing.T) {
		// まず投稿IDを取得する必要がある（前のテストで作成された投稿を使用）
		req := GraphQLRequest{
			Query: `mutation {
				likePost(input: {
					postId: 1
				}) {
					id
					user {
						id
						username
					}
					post {
						id
						content
					}
				}
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": 1,
				},
			},
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		likePost, ok := data["likePost"].(map[string]interface{})
		if !ok {
			t.Fatal("likePost field is not a map")
		}

		if likePost["id"] == nil {
			t.Error("Expected like ID")
		}

		user, ok := likePost["user"].(map[string]interface{})
		if !ok {
			t.Fatal("User field is not a map")
		}

		if user["username"] == nil {
			t.Error("Expected user username")
		}
	})

	t.Run("投稿のいいねを取り消し", func(t *testing.T) {
		req := GraphQLRequest{
			Query: `mutation {
				unlikePost(input: {
					postId: 1
				})
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": 1,
				},
			},
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors != nil {
			t.Errorf("Unexpected errors: %v", resp.Errors)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		unlikePost, ok := data["unlikePost"].(bool)
		if !ok {
			t.Fatal("unlikePost field is not a boolean")
		}

		if !unlikePost {
			t.Error("Expected unlikePost to be true")
		}
	})

	t.Run("無効なクエリ", func(t *testing.T) {
		req := GraphQLRequest{
			Query: `{ invalidField }`,
		}

		resp := executeGraphQLRequest(t, srv, req)

		if resp.Errors == nil {
			t.Error("Expected errors for invalid query")
		}
	})
}

func executeGraphQLRequest(t *testing.T, srv *server.Server, req GraphQLRequest) GraphQLResponse {
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
