package repository

import (
	"sns-server/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles user data operations
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(base *BaseRepository) *UserRepository {
	return &UserRepository{BaseRepository: base}
}

// FindAll returns all users
func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.DB.Find(&users).Error
	return users, err
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

// FindByUsernameOrEmail finds user by username or email
func (r *UserRepository) FindByUsernameOrEmail(username, email string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("username = ? OR email = ?", username, email).First(&user).Error
	return &user, err
}

// FindByUsername finds user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetByID finds user by ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAll returns all users
func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.DB.Find(&users).Error
	return users, err
}

// Delete deletes a user by ID (soft delete with cascade deletion)
func (r *UserRepository) Delete(id uint) error {
	// トランザクションで実行
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 関連データを先に削除
		// 1. ユーザーのいいねを削除
		if err := tx.Where("user_id = ?", id).Delete(&models.Like{}).Error; err != nil {
			return err
		}
		
		// 2. ユーザーの投稿を削除（投稿削除時のカスケード処理も含む）
		if err := tx.Where("author_id = ?", id).Delete(&models.Post{}).Error; err != nil {
			return err
		}
		
		// 3. フォロー関係を削除（フォロワーとフォロイーの両方）
		if err := tx.Where("follower_id = ? OR followee_id = ?", id, id).Delete(&models.Follow{}).Error; err != nil {
			return err
		}
		
		// 4. ユーザー本体を削除（ソフトデリート）
		return tx.Delete(&models.User{}, id).Error
	})
}

// Exists checks if a user exists by ID
func (r *UserRepository) Exists(id uint) (bool, error) {
	var count int64
	err := r.DB.Model(&models.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Update updates user fields by ID
func (r *UserRepository) Update(id uint, updates map[string]interface{}) error {
	// Check if user exists
	exists, err := r.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}

	// Perform update
	result := r.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
