package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // JSONに含めない
	Name      string         `json:"name" gorm:"not null"`
	Bio       string         `json:"bio"`
	Avatar    string         `json:"avatar"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // ソフトデリート

	// リレーション
	Posts     []Post   `json:"posts" gorm:"foreignKey:AuthorID"`
	Likes     []Like   `json:"likes" gorm:"foreignKey:UserID"`
	Following []Follow `json:"following" gorm:"foreignKey:FollowerID"`
	Followers []Follow `json:"followers" gorm:"foreignKey:FolloweeID"`
}

// フォロワー数を取得
func (u *User) FollowerCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Follow{}).Where("followee_id = ?", u.ID).Count(&count)
	return count
}

// フォロー数を取得
func (u *User) FollowingCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Follow{}).Where("follower_id = ?", u.ID).Count(&count)
	return count
}

// 投稿数を取得
func (u *User) PostCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Post{}).Where("author_id = ?", u.ID).Count(&count)
	return count
}

// バリデーション
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
