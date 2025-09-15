package models

import (
	"errors"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
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
	// 必須フィールドチェック
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required")
	}

	// ユーザー名のバリデーション
	if err := u.validateUsername(); err != nil {
		return err
	}

	// メールアドレスのバリデーション
	if err := u.validateEmail(); err != nil {
		return err
	}

	// 名前のバリデーション
	if err := u.validateName(); err != nil {
		return err
	}

	// XSS対策：HTMLエスケープ
	u.sanitizeFields()

	return nil
}

// ユーザー名バリデーション
func (u *User) validateUsername() error {
	username := strings.TrimSpace(u.Username)

	// 長さチェック
	if utf8.RuneCountInString(username) < 2 {
		return errors.New("username must be at least 2 characters")
	}
	if utf8.RuneCountInString(username) > 50 {
		return errors.New("username must be at most 50 characters")
	}

	// 特殊文字チェック（英数字、日本語文字、アンダースコア、ハイフンのみ許可）
	// Unicode文字に対応するためより包括的なパターンを使用
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9\p{L}\p{N}_-]+$`)
	if !validUsername.MatchString(username) {
		return errors.New("username contains invalid characters")
	}

	return nil
}

// メールアドレスバリデーション
func (u *User) validateEmail() error {
	email := strings.TrimSpace(u.Email)

	// 基本的なメール形式チェック
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// 名前バリデーション
func (u *User) validateName() error {
	name := strings.TrimSpace(u.Name)

	// 長さチェック
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be at most 100 characters")
	}

	return nil
}

// XSS対策とフィールドサニタイズ
func (u *User) sanitizeFields() {
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.Name = html.EscapeString(strings.TrimSpace(u.Name))
	u.Bio = html.EscapeString(strings.TrimSpace(u.Bio))
}
