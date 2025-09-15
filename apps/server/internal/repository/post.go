package repository

import (
	"sns-server/internal/models"
	"gorm.io/gorm"
)

// PostRepository handles post data operations
type PostRepository struct {
	*BaseRepository
}

// NewPostRepository creates a new post repository
func NewPostRepository(base *BaseRepository) *PostRepository {
	return &PostRepository{BaseRepository: base}
}

// FindAll returns all posts with author preloaded
func (r *PostRepository) FindAll() ([]models.Post, error) {
	var posts []models.Post
	err := r.DB.Preload("Author").Find(&posts).Error
	return posts, err
}

// Create creates a new post
func (r *PostRepository) Create(post *models.Post) error {
	return r.DB.Create(post).Error
}

// FindByID finds a post by ID with author preloaded
func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.DB.Preload("Author").First(&post, id).Error
	return &post, err
}

// Delete deletes a post by ID (soft delete with cascade deletion of likes)
func (r *PostRepository) Delete(id uint) error {
	// トランザクションで実行
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 関連するいいねを先に削除
		if err := tx.Where("post_id = ?", id).Delete(&models.Like{}).Error; err != nil {
			return err
		}
		
		// 投稿をソフトデリート
		return tx.Delete(&models.Post{}, id).Error
	})
}

// Exists checks if a post exists by ID
func (r *PostRepository) Exists(id uint) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Post{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
