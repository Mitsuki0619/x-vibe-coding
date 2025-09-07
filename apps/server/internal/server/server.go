package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gorm.io/gorm"
	"sns-server/internal/config"
	"sns-server/internal/models"
)

type Server struct {
	DB     *gorm.DB
	Config *config.Config
}

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

func (s *Server) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid JSON")
		return
	}

	// 簡単なクエリルーティング
	response := s.executeQuery(req.Query, req.Variables)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) executeQuery(query string, variables map[string]interface{}) GraphQLResponse {
	// 非常にシンプルなクエリパーサー（実際のプロジェクトでは適切なGraphQLライブラリを使用）

	// ユーザー一覧クエリ
	if contains(query, "users") && !contains(query, "mutation") {
		return s.handleUsersQuery()
	}

	// ユーザー登録ミューテーション
	if contains(query, "register") && contains(query, "mutation") {
		return s.handleRegisterMutation(variables)
	}

	// 投稿作成ミューテーション
	if contains(query, "createPost") && contains(query, "mutation") {
		return s.handleCreatePostMutation(variables)
	}

	// 投稿一覧クエリ
	if contains(query, "posts") && !contains(query, "mutation") {
		return s.handlePostsQuery()
	}

	// いいね取り消しミューテーション（先にチェック）
	if contains(query, "unlikePost") && contains(query, "mutation") {
		return s.handleUnlikePostMutation(variables)
	}

	// いいねミューテーション
	if contains(query, "likePost") && contains(query, "mutation") {
		return s.handleLikePostMutation(variables)
	}

	return GraphQLResponse{
		Errors: []GraphQLError{{Message: "Query not implemented"}},
	}
}

func (s *Server) handleUsersQuery() GraphQLResponse {
	var users []models.User
	if err := s.DB.Find(&users).Error; err != nil {
		return errorResponse(fmt.Sprintf("Database error: %v", err))
	}
	return dataResponse("users", users)
}

func (s *Server) handleRegisterMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return errorResponse("Invalid input format")
	}

	user := models.User{
		Username: getString(input, "username"),
		Email:    getString(input, "email"),
		Password: getString(input, "password"), // TODO: ハッシュ化
		Name:     getString(input, "name"),
		Bio:      getString(input, "bio"),
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return errorResponse(fmt.Sprintf("Failed to create user: %v", err))
	}

	return dataResponse("register", map[string]interface{}{
		"token": "temp_token_" + strconv.Itoa(int(user.ID)),
		"user":  user,
	})
}

func (s *Server) handleCreatePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return errorResponse("Invalid input format - variables required")
	}

	content := getString(input, "content")
	if content == "" {
		return errorResponse("Content is required")
	}

	// デフォルトユーザーを取得または作成
	user, err := s.ensureDefaultUser()
	if err != nil {
		return errorResponse(err.Error())
	}

	post := models.Post{
		Content:  content,
		AuthorID: user.ID,
	}

	if err := s.DB.Create(&post).Error; err != nil {
		return errorResponse(fmt.Sprintf("Failed to create post: %v", err))
	}

	// 作成者情報をプリロード
	s.DB.Preload("Author").First(&post, post.ID)

	return dataResponse("createPost", post)
}

func (s *Server) handlePostsQuery() GraphQLResponse {
	var posts []models.Post
	if err := s.DB.Preload("Author").Order("created_at DESC").Find(&posts).Error; err != nil {
		return errorResponse(fmt.Sprintf("Database error: %v", err))
	}
	return dataResponse("posts", posts)
}

func (s *Server) sendError(w http.ResponseWriter, message string) {
	response := GraphQLResponse{
		Errors: []GraphQLError{{Message: message}},
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			contains(s[1:], substr))))
}

// デフォルトユーザーIDの定数
const defaultUserID uint = 1

// デフォルトユーザーを取得または作成する共通関数
func (s *Server) ensureDefaultUser() (*models.User, error) {
	var user models.User
	if err := s.DB.First(&user, defaultUserID).Error; err != nil {
		// デフォルトユーザーを作成
		user = models.User{
			Username: "default_user",
			Email:    "default@example.com",
			Password: "password",
			Name:     "Default User",
		}
		if createErr := s.DB.Create(&user).Error; createErr != nil {
			return nil, fmt.Errorf("デフォルトユーザーの作成に失敗: %v", createErr)
		}
	}
	return &user, nil
}

// GraphQLエラーレスポンスを作成するヘルパー関数
func errorResponse(message string) GraphQLResponse {
	return GraphQLResponse{
		Errors: []GraphQLError{{Message: message}},
	}
}

// GraphQLデータレスポンスを作成するヘルパー関数
func dataResponse(key string, data interface{}) GraphQLResponse {
	return GraphQLResponse{
		Data: map[string]interface{}{
			key: data,
		},
	}
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

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

func (s *Server) handleLikePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return errorResponse("Invalid input format")
	}

	postID := getUint(input, "postId")
	if postID == 0 {
		return errorResponse("Post ID is required")
	}

	// デフォルトユーザーを取得または作成
	user, err := s.ensureDefaultUser()
	if err != nil {
		return errorResponse(err.Error())
	}

	// 投稿が存在するかチェック
	var post models.Post
	if err := s.DB.First(&post, postID).Error; err != nil {
		return errorResponse("Post not found")
	}

	// いいねを作成
	like := models.Like{
		UserID: user.ID,
		PostID: postID,
	}

	if err := s.DB.Create(&like).Error; err != nil {
		return errorResponse(fmt.Sprintf("Failed to like post: %v", err))
	}

	// 作成されたいいねをリレーション込みで取得
	s.DB.Preload("User").Preload("Post").First(&like, like.ID)

	return dataResponse("likePost", like)
}

func (s *Server) handleUnlikePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return errorResponse("Invalid input format")
	}

	postID := getUint(input, "postId")
	if postID == 0 {
		return errorResponse("Post ID is required")
	}

	// デフォルトユーザーを取得または作成
	user, err := s.ensureDefaultUser()
	if err != nil {
		return errorResponse(err.Error())
	}

	// いいねを削除
	result := s.DB.Where("user_id = ? AND post_id = ?", user.ID, postID).Delete(&models.Like{})
	if result.Error != nil {
		return errorResponse(fmt.Sprintf("Failed to unlike post: %v", result.Error))
	}

	if result.RowsAffected == 0 {
		return errorResponse("Like not found")
	}

	return dataResponse("unlikePost", true)
}
