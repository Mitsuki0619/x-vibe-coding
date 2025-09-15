package handlers

import (
	"sns-server/internal/service"
)

// LikeHandlers contains handlers for like-related GraphQL operations
type LikeHandlers struct {
	LikeService *service.LikeService
}

// NewLikeHandlers creates a new like handlers instance
func NewLikeHandlers(likeService *service.LikeService) *LikeHandlers {
	return &LikeHandlers{
		LikeService: likeService,
	}
}

// HandleLikePostMutation handles the likePost GraphQL mutation
func (h *LikeHandlers) HandleLikePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	likeInput := service.LikePostInput{
		PostID: getUint(input, "postId"),
	}

	like, err := h.LikeService.LikePost(likeInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("likePost", like)
}

// HandleUnlikePostMutation handles the unlikePost GraphQL mutation
func (h *LikeHandlers) HandleUnlikePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	unlikeInput := service.UnlikePostInput{
		PostID: getUint(input, "postId"),
	}

	success, err := h.LikeService.UnlikePost(unlikeInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("unlikePost", success)
}