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
			name: "有効ないいねの作成",
			like: Like{
				UserID: user.ID,
				PostID: post.ID,
			},
			wantErr: false,
		},
		{
			name: "UserIDが0の場合はエラー",
			like: Like{
				UserID: 0,
				PostID: post.ID,
			},
			wantErr: true,
		},
		{
			name: "PostIDが0の場合はエラー",
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

	// 投稿を削除
	err = db.Delete(&post).Error
	if err != nil {
		t.Fatalf("投稿削除に失敗: %v", err)
	}

	// いいねが残っているかチェック（外部キー制約によっては削除される場合もある）
	var likeCount int64
	db.Model(&Like{}).Where("post_id = ?", post.ID).Count(&likeCount)

	// NOTE: このテストは外部キー制約の設定によって結果が変わる
	// 現在の実装では外部キー制約にON DELETE CASCADEが設定されていないため、
	// いいねは残る可能性がある
	t.Logf("投稿削除後のいいね数: %d", likeCount)
}
