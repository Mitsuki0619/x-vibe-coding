package models

import (
	"testing"
)

func TestFollow_Creation(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	db.Create(&user1)
	db.Create(&user2)

	tests := []struct {
		name    string
		follow  Follow
		wantErr bool
	}{
		{
			name: "有効なフォロー関係の作成",
			follow: Follow{
				FollowerID: user1.ID,
				FolloweeID: user2.ID,
			},
			wantErr: false,
		},
		{
			name: "自分自身をフォローしようとした場合はエラー",
			follow: Follow{
				FollowerID: user1.ID,
				FolloweeID: user1.ID,
			},
			wantErr: true,
		},
		{
			name: "フォロワーIDが0の場合はエラー",
			follow: Follow{
				FollowerID: 0,
				FolloweeID: user2.ID,
			},
			wantErr: true,
		},
		{
			name: "フォロウィーIDが0の場合はエラー",
			follow: Follow{
				FollowerID: user1.ID,
				FolloweeID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.Create(&tt.follow)

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if result.Error != nil {
					t.Errorf("Expected no error but got: %v", result.Error)
				}
				if tt.follow.ID == 0 {
					t.Errorf("Expected follow ID to be set")
				}
			}
		})
	}
}

func TestFollow_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		follow   Follow
		expected bool
	}{
		{
			name: "有効なフォロー関係",
			follow: Follow{
				FollowerID: 1,
				FolloweeID: 2,
			},
			expected: true,
		},
		{
			name: "無効なフォロー関係（自分自身）",
			follow: Follow{
				FollowerID: 1,
				FolloweeID: 1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.follow.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFollow_PreventDuplicateFollow(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	db.Create(&user1)
	db.Create(&user2)

	// 最初のフォローは成功する
	follow1 := Follow{
		FollowerID: user1.ID,
		FolloweeID: user2.ID,
	}
	result1 := db.Create(&follow1)
	if result1.Error != nil {
		t.Fatalf("First follow should succeed: %v", result1.Error)
	}

	// 同じフォロー関係の重複作成は失敗する（ユニークキー制約）
	follow2 := Follow{
		FollowerID: user1.ID,
		FolloweeID: user2.ID,
	}
	result2 := db.Create(&follow2)
	if result2.Error == nil {
		t.Errorf("Duplicate follow should fail due to unique constraint")
	}
}
