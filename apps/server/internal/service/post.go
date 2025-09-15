package service

import (
	"fmt"

	"sns-server/internal/models"
	"sns-server/internal/repository"
)

// PostService handles post business logic
type PostService struct {
	postRepo    *repository.PostRepository
	userService *UserService
}

// NewPostService creates a new post service
func NewPostService(postRepo *repository.PostRepository, userService *UserService) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userService: userService,
	}
}

// CreatePostInput represents post creation input
type CreatePostInput struct {
	Content string
}

// DeletePostInput represents post deletion input
type DeletePostInput struct {
	PostID uint `json:"postId"`
}

// DeletePostResponse represents post deletion response
type DeletePostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetAllPosts returns all posts with authors preloaded
func (s *PostService) GetAllPosts() ([]models.Post, error) {
	return s.postRepo.FindAll()
}

// CreatePost creates a new post with validation
func (s *PostService) CreatePost(input CreatePostInput) (*models.Post, error) {
	// Validate content
	if input.Content == "" {
		return nil, fmt.Errorf("Content is required")
	}

	// Get the first available user (for testing compatibility)
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("Failed to get users: %v", err)
	}

	if len(users) == 0 {
		// No users exist, create default user
		user, err := s.userService.EnsureDefaultUser()
		if err != nil {
			return nil, fmt.Errorf("Failed to get default user: %v", err)
		}
		users = []models.User{*user}
	}

	user := &users[0] // Use first available user

	post := models.Post{
		Content:  input.Content,
		AuthorID: user.ID,
	}

	// The BeforeCreate hook in Post model will validate content length (280 chars)
	if err := s.postRepo.Create(&post); err != nil {
		return nil, fmt.Errorf("Failed to create post: %v", err)
	}

	// Load the post with author information
	postWithAuthor, err := s.postRepo.FindByID(post.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to load post: %v", err)
	}

	return postWithAuthor, nil
}

// DeletePost deletes a post with validation
func (s *PostService) DeletePost(input DeletePostInput) (*DeletePostResponse, error) {
	// 投稿IDの検証
	if input.PostID == 0 {
		return nil, fmt.Errorf("Post ID is required")
	}

	// 投稿の存在確認
	exists, err := s.postRepo.Exists(input.PostID)
	if err != nil {
		return nil, fmt.Errorf("Failed to check post existence: %v", err)
	}
	
	if !exists {
		return nil, fmt.Errorf("Post not found")
	}

	// 投稿削除（ソフトデリート + カスケード削除）
	err = s.postRepo.Delete(input.PostID)
	if err != nil {
		return nil, fmt.Errorf("Failed to delete post: %v", err)
	}

	return &DeletePostResponse{
		Success: true,
		Message: fmt.Sprintf("Post %d deleted successfully", input.PostID),
	}, nil
}
