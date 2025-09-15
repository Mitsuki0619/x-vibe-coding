package models

import (
	"testing"
)

func TestLike_Creation(t *testing.T) {
	db := setupTestDB(t)

	// テスト用ユーザーと投稿を作成
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Name:     "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
	}

	post := &Post{
		Content:  "Test post content",
		AuthorID: user.ID,
	}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("テスト用投稿作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		like    Like
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常なデータでいいね作成すると成功する",
			like: Like{
				UserID: user.ID,
				PostID: post.ID,
			},
			wantErr: false,
		},
		{
			name: "UserIDが0の場合エラーが発生する",
			like: Like{
				UserID: 0,
				PostID: post.ID,
			},
			wantErr: true,
		},
		{
			name: "PostIDが0の場合エラーが発生する",
			like: Like{
				UserID: user.ID,
				PostID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Create(&tt.like).Error
			if tt.wantErr {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				}
				// 作成されたいいねを検証
				if tt.like.ID == 0 {
					t.Error("いいねのIDが設定されていません")
				}
				if tt.like.CreatedAt.IsZero() {
					t.Error("いいねの作成日時が設定されていません")
				}
			}
		})
	}
}

func TestLike_PreventDuplicateLike(t *testing.T) {
	db := setupTestDB(t)

	// テスト用ユーザーと投稿を作成
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Name:     "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
	}

	post := &Post{
		Content:  "Test post content",
		AuthorID: user.ID,
	}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("テスト用投稿作成に失敗: %v", err)
	}

	// 最初のいいねを作成
	like1 := Like{
		UserID: user.ID,
		PostID: post.ID,
	}
	err := db.Create(&like1).Error
	if err != nil {
		t.Fatalf("最初のいいね作成に失敗: %v", err)
	}

	// 同じユーザーが同じ投稿に再度いいねしようとする
	like2 := Like{
		UserID: user.ID,
		PostID: post.ID,
	}
	err = db.Create(&like2).Error
	if err == nil {
		t.Error("重複するいいねが作成されてしまいました")
	}
}

func TestLike_Relations(t *testing.T) {
	db := setupTestDB(t)

	// テスト用ユーザーと投稿を作成
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Name:     "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
	}

	post := &Post{
		Content:  "Test post content",
		AuthorID: user.ID,
	}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("テスト用投稿作成に失敗: %v", err)
	}

	// いいねを作成
	like := Like{
		UserID: user.ID,
		PostID: post.ID,
	}
	err := db.Create(&like).Error
	if err != nil {
		t.Fatalf("いいね作成に失敗: %v", err)
	}

	// リレーションを含めて取得
	var retrievedLike Like
	err = db.Preload("User").Preload("Post").First(&retrievedLike, like.ID).Error
	if err != nil {
		t.Fatalf("いいね取得に失敗: %v", err)
	}

	// User関連をテスト
	if retrievedLike.User.ID != user.ID {
		t.Errorf("User ID が一致しません。期待値: %d, 実際: %d", user.ID, retrievedLike.User.ID)
	}
	if retrievedLike.User.Username != user.Username {
		t.Errorf("Username が一致しません。期待値: %s, 実際: %s", user.Username, retrievedLike.User.Username)
	}

	// Post関連をテスト
	if retrievedLike.Post.ID != post.ID {
		t.Errorf("Post ID が一致しません。期待値: %d, 実際: %d", post.ID, retrievedLike.Post.ID)
	}
	if retrievedLike.Post.Content != post.Content {
		t.Errorf("Post Content が一致しません。期待値: %s, 実際: %s", post.Content, retrievedLike.Post.Content)
	}
}

func TestLike_DeleteCascade(t *testing.T) {
	db := setupTestDB(t)

	t.Run("ユーザー削除時にいいねがカスケード削除される", func(t *testing.T) {
		// テスト用ユーザーと投稿を作成
		user := &User{
			Username: "cascadeuser",
			Email:    "cascade@example.com",
			Password: "password",
			Name:     "Cascade User",
		}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
		}

		otherUser := &User{
			Username: "otheruser",
			Email:    "other@example.com",
			Password: "password",
			Name:     "Other User",
		}
		if err := db.Create(otherUser).Error; err != nil {
			t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
		}

		post := &Post{
			Content:  "Cascade test post",
			AuthorID: otherUser.ID,
		}
		if err := db.Create(post).Error; err != nil {
			t.Fatalf("テスト用投稿作成に失敗: %v", err)
		}

		// いいねを作成
		like := &Like{
			UserID: user.ID,
			PostID: post.ID,
		}
		if err := db.Create(like).Error; err != nil {
			t.Fatalf("いいね作成に失敗: %v", err)
		}

		// いいね数を確認（作成後）
		var likeCount int64
		db.Model(&Like{}).Where("user_id = ?", user.ID).Count(&likeCount)
		if likeCount != 1 {
			t.Errorf("Expected 1 like, got %d", likeCount)
		}

		// ユーザーを物理削除（ソフトデリートではなく実削除）
		if err := db.Unscoped().Delete(user).Error; err != nil {
			t.Fatalf("ユーザー削除に失敗: %v", err)
		}

		// いいね数を確認（ユーザー削除後）- カスケード削除される必要がある
		db.Model(&Like{}).Where("user_id = ?", user.ID).Count(&likeCount)
		if likeCount != 0 {
			t.Errorf("Expected 0 likes after user deletion (cascade), got %d", likeCount)
		}
	})

	t.Run("投稿削除時にいいねがカスケード削除される", func(t *testing.T) {
		// テスト用ユーザーと投稿を作成
		user := &User{
			Username: "postdeleteuser",
			Email:    "postdelete@example.com",
			Password: "password",
			Name:     "Post Delete User",
		}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
		}

		post := &Post{
			Content:  "Post delete test",
			AuthorID: user.ID,
		}
		if err := db.Create(post).Error; err != nil {
			t.Fatalf("テスト用投稿作成に失敗: %v", err)
		}

		// いいねを作成
		like := &Like{
			UserID: user.ID,
			PostID: post.ID,
		}
		if err := db.Create(like).Error; err != nil {
			t.Fatalf("いいね作成に失敗: %v", err)
		}

		// いいね数を確認（作成後）
		var likeCount int64
		db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&likeCount)
		if likeCount != 1 {
			t.Errorf("Expected 1 like, got %d", likeCount)
		}

		// 投稿を物理削除（ソフトデリートではなく実削除）
		if err := db.Unscoped().Delete(post).Error; err != nil {
			t.Fatalf("投稿削除に失敗: %v", err)
		}

		// いいね数を確認（投稿削除後）- カスケード削除される必要がある
		db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&likeCount)
		if likeCount != 0 {
			t.Errorf("Expected 0 likes after post deletion (cascade), got %d", likeCount)
		}
	})

	t.Run("複数のいいねがある場合のカスケード削除", func(t *testing.T) {
		// テスト用ユーザーと投稿を作成
		user1 := &User{
			Username: "multiuser1",
			Email:    "multi1@example.com",
			Password: "password",
			Name:     "Multi User 1",
		}
		if err := db.Create(user1).Error; err != nil {
			t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
		}

		user2 := &User{
			Username: "multiuser2",
			Email:    "multi2@example.com",
			Password: "password",
			Name:     "Multi User 2",
		}
		if err := db.Create(user2).Error; err != nil {
			t.Fatalf("テスト用ユーザー作成に失敗: %v", err)
		}

		post := &Post{
			Content:  "Multi like test post",
			AuthorID: user1.ID,
		}
		if err := db.Create(post).Error; err != nil {
			t.Fatalf("テスト用投稿作成に失敗: %v", err)
		}

		// 複数のいいねを作成
		like1 := &Like{
			UserID: user1.ID,
			PostID: post.ID,
		}
		if err := db.Create(like1).Error; err != nil {
			t.Fatalf("いいね1作成に失敗: %v", err)
		}

		like2 := &Like{
			UserID: user2.ID,
			PostID: post.ID,
		}
		if err := db.Create(like2).Error; err != nil {
			t.Fatalf("いいね2作成に失敗: %v", err)
		}

		// 全いいね数を確認
		var totalLikes int64
		db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&totalLikes)
		if totalLikes != 2 {
			t.Errorf("Expected 2 likes, got %d", totalLikes)
		}

		// 投稿を物理削除（ソフトデリートではなく実削除）
		if err := db.Unscoped().Delete(post).Error; err != nil {
			t.Fatalf("投稿削除に失敗: %v", err)
		}

		// すべてのいいねがカスケード削除される必要がある
		db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&totalLikes)
		if totalLikes != 0 {
			t.Errorf("Expected 0 likes after post deletion (cascade), got %d", totalLikes)
		}
	})
}
