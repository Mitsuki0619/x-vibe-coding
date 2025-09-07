package models

import (
	"testing"
)

func TestPost_Creation(t *testing.T) {
	db := setupTestDB(t)

	// テスト用ユーザー作成
	user := User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
	}
	db.Create(&user)

	tests := []struct {
		name    string
		post    Post
		wantErr bool
	}{
		{
			name: "有効な投稿の作成",
			post: Post{
				Content:  "これはテスト投稿です",
				AuthorID: user.ID,
			},
			wantErr: false,
		},
		{
			name: "空の内容では投稿できない",
			post: Post{
				Content:  "",
				AuthorID: user.ID,
			},
			wantErr: true,
		},
		{
			name: "280文字を超える投稿はエラー",
			post: Post{
				Content:  generateLongString(281),
				AuthorID: user.ID,
			},
			wantErr: true,
		},
		{
			name: "280文字ちょうどの投稿は成功",
			post: Post{
				Content:  generateLongString(280),
				AuthorID: user.ID,
			},
			wantErr: false,
		},
		{
			name: "作成者IDがない場合はエラー",
			post: Post{
				Content: "テスト投稿",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.Create(&tt.post)

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if result.Error != nil {
					t.Errorf("Expected no error but got: %v", result.Error)
				}
				if tt.post.ID == 0 {
					t.Errorf("Expected post ID to be set")
				}
			}
		})
	}
}

func TestPost_LikeCount(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザーと投稿作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	db.Create(&user1)
	db.Create(&user2)

	post := Post{Content: "テスト投稿", AuthorID: user1.ID}
	db.Create(&post)

	tests := []struct {
		name          string
		likes         []Like
		expectedCount int64
	}{
		{
			name:          "いいねがない場合",
			likes:         []Like{},
			expectedCount: 0,
		},
		{
			name: "いいねが1つの場合",
			likes: []Like{
				{UserID: user2.ID, PostID: post.ID},
			},
			expectedCount: 1,
		},
		{
			name: "いいねが2つの場合",
			likes: []Like{
				{UserID: user1.ID, PostID: post.ID},
				{UserID: user2.ID, PostID: post.ID},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// いいねデータをクリア
			db.Where("1 = 1").Delete(&Like{})

			// テストデータ作成
			for _, like := range tt.likes {
				db.Create(&like)
			}

			count := post.LikeCount(db)
			if count != tt.expectedCount {
				t.Errorf("Expected %d likes, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestPost_IsLikedByUser(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザーと投稿作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	db.Create(&user1)
	db.Create(&user2)

	post := Post{Content: "テスト投稿", AuthorID: user1.ID}
	db.Create(&post)

	tests := []struct {
		name     string
		likes    []Like
		userID   uint
		expected bool
	}{
		{
			name:     "いいねしていない場合",
			likes:    []Like{},
			userID:   user1.ID,
			expected: false,
		},
		{
			name: "いいねしている場合",
			likes: []Like{
				{UserID: user1.ID, PostID: post.ID},
			},
			userID:   user1.ID,
			expected: true,
		},
		{
			name: "他のユーザーがいいねしているが、対象ユーザーはいいねしていない場合",
			likes: []Like{
				{UserID: user2.ID, PostID: post.ID},
			},
			userID:   user1.ID,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// いいねデータをクリア
			db.Where("1 = 1").Delete(&Like{})

			// テストデータ作成
			for _, like := range tt.likes {
				db.Create(&like)
			}

			result := post.IsLikedByUser(db, tt.userID)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPost_Reply(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	db.Create(&user)

	// 親投稿作成
	parentPost := Post{Content: "親投稿", AuthorID: user.ID}
	db.Create(&parentPost)

	// リプライ投稿作成
	replyPost := Post{
		Content:  "リプライ投稿",
		AuthorID: user.ID,
		ParentID: &parentPost.ID,
	}
	result := db.Create(&replyPost)

	if result.Error != nil {
		t.Errorf("Expected no error creating reply, got: %v", result.Error)
	}

	// リプライ数をチェック
	count := parentPost.ReplyCount(db)
	if count != 1 {
		t.Errorf("Expected 1 reply, got %d", count)
	}
}

// ヘルパー関数：指定された長さの文字列を生成
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
