package testutil

import (
	"log"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sns-server/internal/config"
	"sns-server/internal/models"
)

// SetupTestDB はテスト用データベースを設定します
func SetupTestDB(t *testing.T) *gorm.DB {
	// テスト設定をロード
	cfg := config.LoadTest()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// テスト前にテーブルをクリア
	CleanupDB(t, db)

	// マイグレーション実行
	err = db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Like{},
		&models.Follow{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupDB はテスト用データベースをクリーンアップします
func CleanupDB(t *testing.T, db *gorm.DB) {
	// 外部キー制約があるため、順序に注意してテーブルを削除
	tables := []string{"likes", "follows", "posts", "users"}

	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// CreateTestUser はテスト用ユーザーを作成します
func CreateTestUser(t *testing.T, db *gorm.DB, username, email, name string) *models.User {
	user := &models.User{
		Username: username,
		Email:    email,
		Password: "test_password",
		Name:     name,
		Bio:      "Test user bio",
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestPost はテスト用投稿を作成します
func CreateTestPost(t *testing.T, db *gorm.DB, authorID uint, content string) *models.Post {
	post := &models.Post{
		Content:  content,
		AuthorID: authorID,
	}

	if err := db.Create(post).Error; err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// 作成者をプリロード
	if err := db.Preload("Author").First(post, post.ID).Error; err != nil {
		t.Fatalf("Failed to load test post: %v", err)
	}

	return post
}

// WaitForTestDB はテスト用データベースが利用可能になるまで待機します
func WaitForTestDB() {
	// テスト設定をロード
	cfg := config.LoadTest()

	for i := 0; i < 30; i++ { // 最大30秒待機
		db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
		if err == nil {
			sqlDB, _ := db.DB()
			if sqlDB.Ping() == nil {
				sqlDB.Close()
				log.Println("Test database is ready")
				return
			}
			sqlDB.Close()
		}

		log.Printf("Waiting for test database... (%d/30)", i+1)
		// time.Sleep(1 * time.Second) // 1秒待機
	}

	log.Fatal("Test database is not available after 30 seconds")
}
