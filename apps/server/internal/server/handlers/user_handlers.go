package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"sns-server/internal/service"
)

// UserHandlers contains handlers for user-related GraphQL operations
type UserHandlers struct {
	UserService   *service.UserService
	FollowService *service.FollowService
}

// NewUserHandlers creates a new user handlers instance
func NewUserHandlers(userService *service.UserService, followService *service.FollowService) *UserHandlers {
	return &UserHandlers{
		UserService:   userService,
		FollowService: followService,
	}
}

// HandleUsersQuery handles the users GraphQL query
func (h *UserHandlers) HandleUsersQuery() GraphQLResponse {
	users, err := h.UserService.GetAllUsers()
	if err != nil {
		return ErrorResponse(fmt.Sprintf("Database error: %v", err))
	}

	// フォロー数・フォロワー数を含める
	enrichedUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		followingCount := h.FollowService.FollowRepo.GetFollowingCountByUserID(user.ID)
		followerCount := h.FollowService.FollowRepo.GetFollowerCountByUserID(user.ID)

		enrichedUsers[i] = map[string]interface{}{
			"id":             user.ID,
			"username":       user.Username,
			"email":          user.Email,
			"name":           user.Name,
			"bio":            user.Bio,
			"createdAt":      user.CreatedAt,
			"updatedAt":      user.UpdatedAt,
			"followingCount": followingCount,
			"followerCount":  followerCount,
		}
	}

	return DataResponse("users", enrichedUsers)
}

// HandleRegisterMutation handles the register GraphQL mutation
func (h *UserHandlers) HandleRegisterMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	registerInput := service.RegisterInput{
		Username: getString(input, "username"),
		Email:    getString(input, "email"),
		Password: getString(input, "password"),
		Name:     getString(input, "name"),
		Bio:      getString(input, "bio"),
	}

	result, err := h.UserService.Register(registerInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("register", map[string]interface{}{
		"token": result.Token,
		"user":  result.User,
	})
}

// HandleDeleteUserMutation handles the deleteUser GraphQL mutation
func (h *UserHandlers) HandleDeleteUserMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	deleteUserInput := service.DeleteUserInput{
		UserID: getUint(input, "userId"),
	}

	result, err := h.UserService.DeleteUser(deleteUserInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("deleteUser", result)
}

// HandleUpdateUserMutation handles the updateUser GraphQL mutation
func (h *UserHandlers) HandleUpdateUserMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	// Extract updates map
	updates := make(map[string]interface{})
	if updateData, exists := input["updates"]; exists {
		if updateMap, ok := updateData.(map[string]interface{}); ok {
			updates = updateMap
		}
	} else {
		// If no updates field, use the input directly (except userId)
		for key, value := range input {
			if key != "userId" {
				updates[key] = value
			}
		}
	}

	updateUserInput := service.UpdateUserInput{
		UserID:  getUint(input, "userId"),
		Updates: updates,
	}

	result, err := h.UserService.UpdateUser(updateUserInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("updateUser", result)
}

// HandleUserQuery handles the user(id: X) GraphQL query
func (h *UserHandlers) HandleUserQuery(query string, variables map[string]interface{}) GraphQLResponse {
	// クエリからIDを抽出（簡易パーサー）
	// 実際の実装では適切なGraphQLパーサーを使用
	userID := uint(2) // デフォルトID
	
	// user(id: X) パターンからIDを抽出
	if strings.Contains(query, "user(id:") {
		// "user(id: 2)" から "2" を抽出
		start := strings.Index(query, "user(id:")
		if start != -1 {
			substr := query[start+8:] // "user(id:" の後
			end := strings.Index(substr, ")")
			if end != -1 {
				idStr := strings.TrimSpace(substr[:end])
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					userID = uint(id)
				}
			}
		}
	}

	// ユーザー取得
	user, err := h.UserService.GetUserByID(userID)
	if err != nil {
		return ErrorResponse(fmt.Sprintf("User not found: %v", err))
	}

	// フォロワー・フォロー情報を取得
	followers, _ := h.FollowService.FollowRepo.GetFollowersByUserID(userID)
	following, _ := h.FollowService.FollowRepo.GetFollowingByUserID(userID)
	followerCount := h.FollowService.FollowRepo.GetFollowerCountByUserID(userID)
	followingCount := h.FollowService.FollowRepo.GetFollowingCountByUserID(userID)

	fmt.Printf("DEBUG: User %d has %d followers, %d following\n", userID, len(followers), len(following))
	fmt.Printf("DEBUG: Follower count: %d, Following count: %d\n", followerCount, followingCount)

	// レスポンス構築
	userResponse := map[string]interface{}{
		"id":             user.ID,
		"username":       user.Username,
		"name":           user.Name,
		"email":          user.Email,
		"bio":            user.Bio,
		"createdAt":      user.CreatedAt,
		"updatedAt":      user.UpdatedAt,
		"followerCount":  followerCount,
		"followingCount": followingCount,
		"followers":      followers,
		"following":      following,
	}

	return DataResponse("user", userResponse)
}