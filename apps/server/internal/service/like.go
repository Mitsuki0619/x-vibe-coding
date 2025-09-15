package service

import (
	"fmt"

	"sns-server/internal/models"
	"sns-server/internal/repository"
)

// LikeService handles like business logic
type LikeService struct {
	likeRepo    *repository.LikeRepository
	postRepo    *repository.PostRepository
	userService *UserService
}

// NewLikeService creates a new like service
func NewLikeService(likeRepo *repository.LikeRepository, postRepo *repository.PostRepository, userService *UserService) *LikeService {
	return &LikeService{
		likeRepo:    likeRepo,
		postRepo:    postRepo,
		userService: userService,
	}
}

// LikePostInput represents like post input
type LikePostInput struct {
	PostID uint
}

// UnlikePostInput represents unlike post input
type UnlikePostInput struct {
	PostID uint
}

// getDefaultUser returns the first available user for operations (testing compatibility)
func (s *LikeService) getDefaultUser() (*models.User, error) {
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
		return user, nil
	}

	return &users[0], nil
}

// validatePostID validates that post ID is valid
func (s *LikeService) validatePostID(postID uint) error {
	if postID == 0 {
		return fmt.Errorf("Post ID is required")
	}
	return nil
}

// LikePost creates a like for a post
func (s *LikeService) LikePost(input LikePostInput) (*models.Like, error) {
	if err := s.validatePostID(input.PostID); err != nil {
		return nil, err
	}

	user, err := s.getDefaultUser()
	if err != nil {
		return nil, err
	}

	// Check if post exists
	_, err = s.postRepo.FindByID(input.PostID)
	if err != nil {
		return nil, fmt.Errorf("Post not found")
	}

	// Create the like
	like := models.Like{
		UserID: user.ID,
		PostID: input.PostID,
	}

	if err := s.likeRepo.Create(&like); err != nil {
		return nil, fmt.Errorf("Failed to like post: %v", err)
	}

	// Load the like with relations
	likeWithRelations, err := s.likeRepo.FindByIDWithRelations(like.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to load like: %v", err)
	}

	return likeWithRelations, nil
}

// UnlikePost removes a like from a post
func (s *LikeService) UnlikePost(input UnlikePostInput) (bool, error) {
	if err := s.validatePostID(input.PostID); err != nil {
		return false, err
	}

	user, err := s.getDefaultUser()
	if err != nil {
		return false, err
	}

	// Find the like
	like, err := s.likeRepo.FindByUserAndPost(user.ID, input.PostID)
	if err != nil {
		return false, fmt.Errorf("Like not found")
	}

	// Delete the like
	if err := s.likeRepo.Delete(like); err != nil {
		return false, fmt.Errorf("Failed to unlike post: %v", err)
	}

	return true, nil
}
