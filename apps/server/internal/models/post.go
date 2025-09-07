package models

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Post struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Content   string         `json:"content" gorm:"not null;size:280"` // Twitter風の文字制限
	AuthorID  uint           `json:"authorId" gorm:"not null"`
	ParentID  *uint          `json:"parentId"` // リプライ用（NULLable）
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // ソフトデリート

	// リレーション
	Author  User   `json:"author" gorm:"foreignKey:AuthorID"`
	Parent  *Post  `json:"parent" gorm:"foreignKey:ParentID"` // リプライ元
	Replies []Post `json:"replies" gorm:"foreignKey:ParentID"`
	Likes   []Like `json:"likes" gorm:"foreignKey:PostID"`
}

// いいね数を取得
func (p *Post) LikeCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Like{}).Where("post_id = ?", p.ID).Count(&count)
	return count
}

// リプライ数を取得
func (p *Post) ReplyCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Post{}).Where("parent_id = ?", p.ID).Count(&count)
	return count
}

// ユーザーがいいねしているかチェック
func (p *Post) IsLikedByUser(db *gorm.DB, userID uint) bool {
	var count int64
	db.Model(&Like{}).Where("post_id = ? AND user_id = ?", p.ID, userID).Count(&count)
	return count > 0
}

// バリデーション
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	// 内容が空でないかチェック
	if strings.TrimSpace(p.Content) == "" {
		return errors.New("content cannot be empty")
	}

	// 280文字制限チェック
	if len([]rune(p.Content)) > 280 {
		return errors.New("content exceeds 280 characters")
	}

	// 作成者IDが設定されているかチェック
	if p.AuthorID == 0 {
		return errors.New("author ID is required")
	}

	return nil
}
