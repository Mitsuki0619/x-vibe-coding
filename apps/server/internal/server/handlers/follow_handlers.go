package handlers

import (
	"sns-server/internal/service"
)

// FollowHandlers contains handlers for follow-related GraphQL operations
type FollowHandlers struct {
	FollowService *service.FollowService
}

// NewFollowHandlers creates a new follow handlers instance
func NewFollowHandlers(followService *service.FollowService) *FollowHandlers {
	return &FollowHandlers{
		FollowService: followService,
	}
}

// HandleFollowUserMutation handles the followUser GraphQL mutation
func (h *FollowHandlers) HandleFollowUserMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	followInput := service.FollowUserInput{
		FolloweeID: getUint(input, "followeeId"),
	}

	follow, err := h.FollowService.FollowUser(followInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("followUser", follow)
}

// HandleUnfollowUserMutation handles the unfollowUser GraphQL mutation
func (h *FollowHandlers) HandleUnfollowUserMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	unfollowInput := service.UnfollowUserInput{
		FolloweeID: getUint(input, "followeeId"),
	}

	result, err := h.FollowService.UnfollowUser(unfollowInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("unfollowUser", result)
}