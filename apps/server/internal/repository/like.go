package repository

import (
	"sns-server/internal/models"
)

// LikeRepository handles like data operations
type LikeRepository struct {
	*BaseRepository
}

// NewLikeRepository creates a new like repository
func NewLikeRepository(base *BaseRepository) *LikeRepository {
	return &LikeRepository{BaseRepository: base}
}

// Create creates a new like
func (r *LikeRepository) Create(like *models.Like) error {
	return r.DB.Create(like).Error
}

// FindByUserAndPost finds a like by user and post ID
func (r *LikeRepository) FindByUserAndPost(userID, postID uint) (*models.Like, error) {
	var like models.Like
	err := r.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
	return &like, err
}

// Delete deletes a like
func (r *LikeRepository) Delete(like *models.Like) error {
	return r.DB.Delete(like).Error
}

// FindByIDWithRelations finds a like by ID with user and post preloaded
func (r *LikeRepository) FindByIDWithRelations(id uint) (*models.Like, error) {
	var like models.Like
	err := r.DB.Preload("User").Preload("Post").First(&like, id).Error
	return &like, err
}
