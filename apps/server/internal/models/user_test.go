package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// テスト用テーブル作成
	err = db.AutoMigrate(&User{}, &Post{}, &Like{}, &Follow{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestUser_Creation(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "有効なユーザーの作成",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
				Name:     "Test User",
				Bio:      "This is a test user",
			},
			wantErr: false,
		},
		{
			name: "ユーザー名が空の場合はエラー",
			user: User{
				Email:    "test2@example.com",
				Password: "hashedpassword",
				Name:     "Test User 2",
			},
			wantErr: true,
		},
		{
			name: "メールアドレスが空の場合はエラー",
			user: User{
				Username: "testuser2",
				Password: "hashedpassword",
				Name:     "Test User 2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.Create(&tt.user)

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if result.Error != nil {
					t.Errorf("Expected no error but got: %v", result.Error)
				}
				if tt.user.ID == 0 {
					t.Errorf("Expected user ID to be set")
				}
			}
		})
	}
}

func TestUser_FollowerCount(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	user3 := User{Username: "user3", Email: "user3@test.com", Password: "pass", Name: "User 3"}

	db.Create(&user1)
	db.Create(&user2)
	db.Create(&user3)

	tests := []struct {
		name          string
		targetUser    *User
		followers     []Follow
		expectedCount int64
	}{
		{
			name:          "フォロワーがいない場合",
			targetUser:    &user1,
			followers:     []Follow{},
			expectedCount: 0,
		},
		{
			name:       "フォロワーが1人の場合",
			targetUser: &user1,
			followers: []Follow{
				{FollowerID: user2.ID, FolloweeID: user1.ID},
			},
			expectedCount: 1,
		},
		{
			name:       "フォロワーが2人の場合",
			targetUser: &user1,
			followers: []Follow{
				{FollowerID: user2.ID, FolloweeID: user1.ID},
				{FollowerID: user3.ID, FolloweeID: user1.ID},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// フォローデータをクリア
			db.Where("1 = 1").Delete(&Follow{})

			// テストデータ作成
			for _, follow := range tt.followers {
				db.Create(&follow)
			}

			count := tt.targetUser.FollowerCount(db)
			if count != tt.expectedCount {
				t.Errorf("Expected %d followers, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestUser_FollowingCount(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	user3 := User{Username: "user3", Email: "user3@test.com", Password: "pass", Name: "User 3"}

	db.Create(&user1)
	db.Create(&user2)
	db.Create(&user3)

	tests := []struct {
		name          string
		targetUser    *User
		followings    []Follow
		expectedCount int64
	}{
		{
			name:          "フォローしていない場合",
			targetUser:    &user1,
			followings:    []Follow{},
			expectedCount: 0,
		},
		{
			name:       "1人をフォローしている場合",
			targetUser: &user1,
			followings: []Follow{
				{FollowerID: user1.ID, FolloweeID: user2.ID},
			},
			expectedCount: 1,
		},
		{
			name:       "2人をフォローしている場合",
			targetUser: &user1,
			followings: []Follow{
				{FollowerID: user1.ID, FolloweeID: user2.ID},
				{FollowerID: user1.ID, FolloweeID: user3.ID},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// フォローデータをクリア
			db.Where("1 = 1").Delete(&Follow{})

			// テストデータ作成
			for _, follow := range tt.followings {
				db.Create(&follow)
			}

			count := tt.targetUser.FollowingCount(db)
			if count != tt.expectedCount {
				t.Errorf("Expected %d following, got %d", tt.expectedCount, count)
			}
		})
	}
}
