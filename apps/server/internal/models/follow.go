package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Follow struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FollowerID uint      `json:"followerId" gorm:"not null;uniqueIndex:idx_follower_followee"` // フォローする人
	FolloweeID uint      `json:"followeeId" gorm:"not null;uniqueIndex:idx_follower_followee"` // フォローされる人
	CreatedAt  time.Time `json:"createdAt"`

	// リレーション
	Follower User `json:"follower" gorm:"foreignKey:FollowerID"`
	Followee User `json:"followee" gorm:"foreignKey:FolloweeID"`
}

// 複合ユニークキー（同じユーザーを複数回フォローできないように）
func (Follow) TableName() string {
	return "follows"
}

// 自分自身をフォローできないようにバリデーション
func (f *Follow) IsValid() bool {
	return f.FollowerID != f.FolloweeID
}

// バリデーション
func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	// フォロワーIDが設定されているかチェック
	if f.FollowerID == 0 {
		return errors.New("follower ID is required")
	}

	// フォロウィーIDが設定されているかチェック
	if f.FolloweeID == 0 {
		return errors.New("followee ID is required")
	}

	// 自分自身をフォローしようとしていないかチェック
	if !f.IsValid() {
		return errors.New("cannot follow yourself")
	}

	return nil
}
