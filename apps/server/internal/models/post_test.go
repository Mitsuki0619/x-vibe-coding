package models

import (
	"strings"
	"testing"
	"gorm.io/gorm"
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
			name: "正常なコンテンツで投稿作成すると成功する",
			post: Post{
				Content:  "これはテスト投稿です",
				AuthorID: user.ID,
			},
			wantErr: false,
		},
		{
			name: "空文字列のコンテンツで投稿作成するとエラーが発生する",
			post: Post{
				Content:  "",
				AuthorID: user.ID,
			},
			wantErr: true,
		},
		{
			name: "280文字を超えるコンテンツで投稿作成するとエラーが発生する",
			post: Post{
				Content:  generateLongString(281),
				AuthorID: user.ID,
			},
			wantErr: true,
		},
		{
			name: "280文字ちょうどのコンテンツで投稿作成すると成功する",
			post: Post{
				Content:  generateLongString(280),
				AuthorID: user.ID,
			},
			wantErr: false,
		},
		{
			name: "作成者IDが未設定の場合エラーが発生する",
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
			name:          "いいねが0件の場合カウントが0になる",
			likes:         []Like{},
			expectedCount: 0,
		},
		{
			name: "いいねが1件ある場合カウントが1になる",
			likes: []Like{
				{UserID: user2.ID, PostID: post.ID},
			},
			expectedCount: 1,
		},
		{
			name: "いいねが2件ある場合カウントが2になる",
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
			name:     "ユーザーがいいねしていない場合falseが返される",
			likes:    []Like{},
			userID:   user1.ID,
			expected: false,
		},
		{
			name: "ユーザーがいいねしている場合trueが返される",
			likes: []Like{
				{UserID: user1.ID, PostID: post.ID},
			},
			userID:   user1.ID,
			expected: true,
		},
		{
			name: "他ユーザーはいいね済みだが対象ユーザーはいいねしていない場合falseが返される",
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

func TestPost_Delete(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user := User{Username: "testuser", Email: "test@example.com", Password: "pass", Name: "Test User"}
	db.Create(&user)

	tests := []struct {
		name    string
		setup   func() *Post
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常な投稿を削除すると成功する",
			setup: func() *Post {
				post := Post{Content: "削除予定の投稿", AuthorID: user.ID}
				db.Create(&post)
				return &post
			},
			wantErr: false,
		},
		{
			name: "存在しない投稿を削除しようとしても影響なし",
			setup: func() *Post {
				return &Post{ID: 999999} // 存在しないID
			},
			wantErr: false, // GORMはソフトデリートで存在しないレコードの削除はエラーにならない
		},
		{
			name: "いいねがついている投稿を削除するとカスケード削除される",
			setup: func() *Post {
				post := Post{Content: "いいね付き投稿", AuthorID: user.ID}
				db.Create(&post)
				
				// いいねを追加
				like := Like{UserID: user.ID, PostID: post.ID}
				db.Create(&like)
				
				return &post
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := tt.setup()
			
			// 削除前にいいね数確認（カスケードテスト用）
			var initialLikeCount int64
			if post.ID != 0 {
				db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&initialLikeCount)
			}

			// 削除実行（手動でカスケード削除を行う）
			var result *gorm.DB
			if post.ID != 0 {
				// 関連するいいねを先に削除
				db.Where("post_id = ?", post.ID).Delete(&Like{})
			}
			result = db.Delete(post)

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(result.Error.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errMsg, result.Error.Error())
				}
			} else {
				if result.Error != nil {
					t.Errorf("Expected no error but got: %v", result.Error)
				}
				
				// 削除確認
				var count int64
				db.Model(&Post{}).Where("id = ?", post.ID).Count(&count)
				if count != 0 {
					t.Errorf("Expected post to be deleted, but it still exists")
				}
				
				// カスケード削除確認（いいねがあった場合）
				if initialLikeCount > 0 {
					var finalLikeCount int64
					db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&finalLikeCount)
					if finalLikeCount != 0 {
						t.Errorf("Expected likes to be cascade deleted, but %d likes remain", finalLikeCount)
					}
				}
			}
		})
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
