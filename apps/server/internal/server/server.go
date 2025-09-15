package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"
	"sns-server/internal/config"
	"sns-server/internal/repository"
	"sns-server/internal/service"
	"sns-server/internal/server/handlers"
)

type Server struct {
	DB     *gorm.DB
	Config *config.Config

	// Services
	UserService   *service.UserService
	PostService   *service.PostService
	LikeService   *service.LikeService
	FollowService *service.FollowService

	// Handlers
	UserHandlers   *handlers.UserHandlers
	PostHandlers   *handlers.PostHandlers
	LikeHandlers   *handlers.LikeHandlers
	FollowHandlers *handlers.FollowHandlers
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// NewServer creates a new server with all services and handlers initialized
func NewServer(db *gorm.DB, config *config.Config) *Server {
	// Initialize repositories
	baseRepo := repository.NewBaseRepository(db)
	userRepo := repository.NewUserRepository(baseRepo)
	postRepo := repository.NewPostRepository(baseRepo)
	likeRepo := repository.NewLikeRepository(baseRepo)
	followRepo := repository.NewFollowRepository(baseRepo)

	// Initialize services
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userService)
	likeService := service.NewLikeService(likeRepo, postRepo, userService)
	followService := service.NewFollowService(followRepo, userRepo)

	// Initialize handlers
	userHandlers := handlers.NewUserHandlers(userService, followService)
	postHandlers := handlers.NewPostHandlers(postService)
	likeHandlers := handlers.NewLikeHandlers(likeService)
	followHandlers := handlers.NewFollowHandlers(followService)

	return &Server{
		DB:            db,
		Config:        config,
		UserService:   userService,
		PostService:   postService,
		LikeService:   likeService,
		FollowService: followService,
		UserHandlers:   userHandlers,
		PostHandlers:   postHandlers,
		LikeHandlers:   likeHandlers,
		FollowHandlers: followHandlers,
	}
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

func (s *Server) executeQuery(query string, variables map[string]interface{}) handlers.GraphQLResponse {
	// 非常にシンプルなクエリパーサー（実際のプロジェクトでは適切なGraphQLライブラリを使用）

	// ユーザー一覧クエリ
	if contains(query, "users") && !contains(query, "mutation") {
		return s.UserHandlers.HandleUsersQuery()
	}

	// ユーザー登録ミューテーション
	if contains(query, "register") && contains(query, "mutation") {
		return s.UserHandlers.HandleRegisterMutation(variables)
	}

	// ユーザー削除ミューテーション
	if contains(query, "deleteUser") && contains(query, "mutation") {
		return s.UserHandlers.HandleDeleteUserMutation(variables)
	}

	// ユーザー更新ミューテーション
	if contains(query, "updateUser") && contains(query, "mutation") {
		return s.UserHandlers.HandleUpdateUserMutation(variables)
	}

	// 投稿作成ミューテーション
	if contains(query, "createPost") && contains(query, "mutation") {
		return s.PostHandlers.HandleCreatePostMutation(variables)
	}

	// 投稿削除ミューテーション
	if contains(query, "deletePost") && contains(query, "mutation") {
		return s.PostHandlers.HandleDeletePostMutation(variables)
	}

	// 投稿一覧クエリ
	if contains(query, "posts") && !contains(query, "mutation") {
		return s.PostHandlers.HandlePostsQuery()
	}

	// いいね取り消しミューテーション（先にチェック）
	if contains(query, "unlikePost") && contains(query, "mutation") {
		return s.LikeHandlers.HandleUnlikePostMutation(variables)
	}

	// いいねミューテーション
	if contains(query, "likePost") && contains(query, "mutation") {
		return s.LikeHandlers.HandleLikePostMutation(variables)
	}

	// フォロー解除ミューテーション（先にチェック）
	if contains(query, "unfollowUser") && contains(query, "mutation") {
		return s.FollowHandlers.HandleUnfollowUserMutation(variables)
	}

	// フォローミューテーション
	if contains(query, "followUser") && contains(query, "mutation") {
		return s.FollowHandlers.HandleFollowUserMutation(variables)
	}

	// ユーザー詳細クエリ（フォロー関係含む）
	if contains(query, "user(") && !contains(query, "mutation") {
		return s.UserHandlers.HandleUserQuery(query, variables)
	}

	return handlers.GraphQLResponse{
		Errors: []handlers.GraphQLError{{Message: "Query not implemented"}},
	}
}

func (s *Server) sendError(w http.ResponseWriter, message string) {
	response := handlers.GraphQLResponse{
		Errors: []handlers.GraphQLError{{Message: message}},
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}