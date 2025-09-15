package handlers

import (
	"fmt"

	"sns-server/internal/service"
)

// PostHandlers contains handlers for post-related GraphQL operations
type PostHandlers struct {
	PostService *service.PostService
}

// NewPostHandlers creates a new post handlers instance
func NewPostHandlers(postService *service.PostService) *PostHandlers {
	return &PostHandlers{
		PostService: postService,
	}
}

// HandleCreatePostMutation handles the createPost GraphQL mutation
func (h *PostHandlers) HandleCreatePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format - variables required")
	}

	createPostInput := service.CreatePostInput{
		Content: getString(input, "content"),
	}

	post, err := h.PostService.CreatePost(createPostInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("createPost", post)
}

// HandlePostsQuery handles the posts GraphQL query
func (h *PostHandlers) HandlePostsQuery() GraphQLResponse {
	posts, err := h.PostService.GetAllPosts()
	if err != nil {
		return ErrorResponse(fmt.Sprintf("Database error: %v", err))
	}
	return DataResponse("posts", posts)
}

// HandleDeletePostMutation handles the deletePost GraphQL mutation
func (h *PostHandlers) HandleDeletePostMutation(variables map[string]interface{}) GraphQLResponse {
	input, ok := variables["input"].(map[string]interface{})
	if !ok {
		return ErrorResponse("Invalid input format")
	}

	deletePostInput := service.DeletePostInput{
		PostID: getUint(input, "postId"),
	}

	result, err := h.PostService.DeletePost(deletePostInput)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return DataResponse("deletePost", result)
}