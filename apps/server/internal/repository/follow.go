package repository

import (
	"sns-server/internal/models"
)

type FollowRepository struct {
	*BaseRepository
}

func NewFollowRepository(base *BaseRepository) *FollowRepository {
	return &FollowRepository{BaseRepository: base}
}

// Create creates a new follow relationship
func (r *FollowRepository) Create(follow *models.Follow) error {
	return r.DB.Create(follow).Error
}

// Delete deletes a follow relationship by ID
func (r *FollowRepository) Delete(id uint) error {
	return r.DB.Delete(&models.Follow{}, id).Error
}

// GetByFollowerAndFollowee gets a follow relationship by follower and followee IDs
func (r *FollowRepository) GetByFollowerAndFollowee(followerID, followeeID uint) (*models.Follow, error) {
	var follow models.Follow
	err := r.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&follow).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

// GetFollowersByUserID gets all followers of a user
func (r *FollowRepository) GetFollowersByUserID(userID uint) ([]models.Follow, error) {
	var follows []models.Follow
	err := r.DB.Where("followee_id = ?", userID).Preload("Follower").Find(&follows).Error
	return follows, err
}

// GetFollowingByUserID gets all users that a user is following
func (r *FollowRepository) GetFollowingByUserID(userID uint) ([]models.Follow, error) {
	var follows []models.Follow
	err := r.DB.Where("follower_id = ?", userID).Preload("Followee").Find(&follows).Error
	return follows, err
}

// GetFollowerCountByUserID gets the number of followers for a user
func (r *FollowRepository) GetFollowerCountByUserID(userID uint) int64 {
	var count int64
	r.DB.Model(&models.Follow{}).Where("followee_id = ?", userID).Count(&count)
	return count
}

// GetFollowingCountByUserID gets the number of users that a user is following
func (r *FollowRepository) GetFollowingCountByUserID(userID uint) int64 {
	var count int64
	r.DB.Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&count)
	return count
}
