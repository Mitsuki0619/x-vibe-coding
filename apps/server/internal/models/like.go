package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Like struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"not null;uniqueIndex:idx_user_post"`
	PostID    uint      `json:"postId" gorm:"not null;uniqueIndex:idx_user_post"`
	CreatedAt time.Time `json:"createdAt"`

	// リレーション
	User User `json:"user" gorm:"foreignKey:UserID"`
	Post Post `json:"post" gorm:"foreignKey:PostID"`
}

// 複合ユニークキー（同じユーザーが同じ投稿に複数回いいねできないように）
func (Like) TableName() string {
	return "likes"
}

// BeforeCreate はレコード作成前のバリデーション
func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.UserID == 0 {
		return errors.New("ユーザーIDは必須です")
	}
	if l.PostID == 0 {
		return errors.New("投稿IDは必須です")
	}
	return nil
}
