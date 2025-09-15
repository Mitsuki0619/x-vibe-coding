package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB は全てのテストで共通して使用されるテストデータベースセットアップ関数
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// SQLiteの外部キー制約を有効化（カスケード削除用）
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("Failed to enable foreign key constraints: %v", err)
	}

	// User、Post、Follow テーブルを最初に作成
	err = db.AutoMigrate(&User{}, &Post{}, &Follow{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Likesテーブルを手動で作成（カスケード削除制約付き）
	if err := db.Migrator().DropTable(&Like{}); err != nil && err.Error() != "table likes not found" {
		t.Logf("Error dropping likes table: %v", err)
	}

	createLikesSQL := `CREATE TABLE likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		created_at DATETIME,
		CONSTRAINT fk_likes_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		CONSTRAINT fk_likes_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		UNIQUE(user_id, post_id)
	)`

	if err := db.Exec(createLikesSQL).Error; err != nil {
		t.Fatalf("Failed to create likes table with cascade constraints: %v", err)
	}

	return db
}
