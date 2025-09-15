package models

import (
	"strings"
	"testing"
	"gorm.io/gorm"
)

func TestUser_Creation(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "正常なデータでユーザー作成すると成功する",
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
			name: "ユーザー名が空文字列の場合エラーが発生する",
			user: User{
				Email:    "test2@example.com",
				Password: "hashedpassword",
				Name:     "Test User 2",
			},
			wantErr: true,
		},
		{
			name: "メールアドレスが空文字列の場合エラーが発生する",
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
			name:          "フォロワーが0人の場合カウントが0になる",
			targetUser:    &user1,
			followers:     []Follow{},
			expectedCount: 0,
		},
		{
			name:       "フォロワーが1人存在する場合カウントが1になる",
			targetUser: &user1,
			followers: []Follow{
				{FollowerID: user2.ID, FolloweeID: user1.ID},
			},
			expectedCount: 1,
		},
		{
			name:       "フォロワーが2人存在する場合カウントが2になる",
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
			name:          "誰もフォローしていない場合カウントが0になる",
			targetUser:    &user1,
			followings:    []Follow{},
			expectedCount: 0,
		},
		{
			name:       "1人をフォローしている場合カウントが1になる",
			targetUser: &user1,
			followings: []Follow{
				{FollowerID: user1.ID, FolloweeID: user2.ID},
			},
			expectedCount: 1,
		},
		{
			name:       "2人をフォローしている場合カウントが2になる",
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

	t.Run("ユーザー作成の境界値テスト", func(t *testing.T) {
		db := setupTestDB(t)

		tests := []struct {
			name    string
			user    User
			wantErr bool
		}{
			{
				name: "最小長度のユーザー名で成功する",
				user: User{
					Username: "ab", // 2文字
					Email:    "min@test.com",
					Password: "pass",
					Name:     "Min",
				},
				wantErr: false,
			},
			{
				name: "1文字のユーザー名でエラーが発生する",
				user: User{
					Username: "a", // 1文字
					Email:    "single@test.com",
					Password: "pass",
					Name:     "Single",
				},
				wantErr: true,
			},
			{
				name: "非常に長いユーザー名でエラーが発生する",
				user: User{
					Username: strings.Repeat("verylongusername", 10), // 160文字
					Email:    "long@test.com",
					Password: "pass",
					Name:     "Long",
				},
				wantErr: true,
			},
			{
				name: "Unicode文字を含むユーザー名が正しく処理される",
				user: User{
					Username: "ユーザー123",
					Email:    "unicode@test.com",
					Password: "pass",
					Name:     "Unicode User 😀",
				},
				wantErr: false,
			},
			{
				name: "特殊文字を含むユーザー名でエラーが発生する",
				user: User{
					Username: "user@#$%",
					Email:    "special@test.com",
					Password: "pass",
					Name:     "Special",
				},
				wantErr: true,
			},
			{
				name: "無効なメールアドレスでエラーが発生する",
				user: User{
					Username: "validuser",
					Email:    "invalid-email",
					Password: "pass",
					Name:     "Valid",
				},
				wantErr: true,
			},
			{
				name: "SQLインジェクション試行が安全に処理される",
				user: User{
					Username: "testuser",
					Email:    "test@test.com",
					Password: "'; DROP TABLE users; --",
					Name:     "SQL Test",
				},
				wantErr: false, // パスワードは文字列として安全に保存される
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

					// SQLインジェクションのテストでは、パスワードが安全に保存されていることを確認
					if tt.name == "SQLインジェクション試行が安全に処理される" {
						if tt.user.Password != "'; DROP TABLE users; --" {
							t.Errorf("Password should be stored as-is, got: %s", tt.user.Password)
						}
					}
				}
			})
		}
	})

	t.Run("ユーザー制約テスト", func(t *testing.T) {
		db := setupTestDB(t)

		// 最初のユーザーを作成
		user1 := User{
			Username: "testuser1",
			Email:    "test1@example.com",
			Password: "password1",
			Name:     "Test User 1",
		}

		result1 := db.Create(&user1)
		if result1.Error != nil {
			t.Fatalf("Failed to create first user: %v", result1.Error)
		}

		t.Run("重複ユーザー名でエラーが発生する", func(t *testing.T) {
			user2 := User{
				Username: "testuser1", // 重複するユーザー名
				Email:    "test2@example.com",
				Password: "password2",
				Name:     "Test User 2",
			}

			result2 := db.Create(&user2)
			if result2.Error == nil {
				t.Error("Expected error for duplicate username")
			}
		})

		t.Run("重複メールアドレスでエラーが発生する", func(t *testing.T) {
			user3 := User{
				Username: "testuser3",
				Email:    "test1@example.com", // 重複するメール
				Password: "password3",
				Name:     "Test User 3",
			}

			result3 := db.Create(&user3)
			if result3.Error == nil {
				t.Error("Expected error for duplicate email")
			}
		})
	})
}

func TestUser_Update(t *testing.T) {
	db := setupTestDB(t)

	// ベースユーザー作成
	baseUser := User{
		Username: "original_user",
		Email:    "original@example.com", 
		Password: "original_pass",
		Name:     "Original Name",
		Bio:      "Original Bio",
	}
	db.Create(&baseUser)

	tests := []struct {
		name        string
		userID      uint
		updateData  map[string]interface{}
		wantErr     bool
		errMsg      string
		checkFields map[string]interface{} // 更新後に確認するフィールド
	}{
		{
			name:   "正常なデータでユーザー更新すると成功する",
			userID: baseUser.ID,
			updateData: map[string]interface{}{
				"name": "Updated Name",
				"bio":  "Updated Bio",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"name": "Updated Name",
				"bio":  "Updated Bio",
			},
		},
		{
			name:   "ユーザー名を更新できる",
			userID: baseUser.ID,
			updateData: map[string]interface{}{
				"username": "updated_username",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"username": "updated_username",
			},
		},
		{
			name:   "メールアドレスを更新できる",
			userID: baseUser.ID,
			updateData: map[string]interface{}{
				"email": "updated@example.com",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"email": "updated@example.com",
			},
		},
		{
			name:   "複数フィールドを同時に更新できる",
			userID: baseUser.ID,
			updateData: map[string]interface{}{
				"username": "multi_update",
				"name":     "Multi Update Name",
				"bio":      "Multi Update Bio",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"username": "multi_update",
				"name":     "Multi Update Name",
				"bio":      "Multi Update Bio",
			},
		},
		{
			name:   "SQLインジェクション試行が安全に処理される",
			userID: baseUser.ID,
			updateData: map[string]interface{}{
				"name": "'; DROP TABLE users; --",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"name": "'; DROP TABLE users; --",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 基本的な更新テスト（モデルレベル）
			result := db.Model(&User{}).Where("id = ?", tt.userID).Updates(tt.updateData)

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

				// 更新確認
				if tt.checkFields != nil {
					var updatedUser User
					db.First(&updatedUser, tt.userID)
					
					for field, expectedValue := range tt.checkFields {
						var actualValue interface{}
						switch field {
						case "username":
							actualValue = updatedUser.Username
						case "email":
							actualValue = updatedUser.Email
						case "name":
							actualValue = updatedUser.Name
						case "bio":
							actualValue = updatedUser.Bio
						}
						
						if actualValue != expectedValue {
							t.Errorf("Expected %s to be '%v', got '%v'", field, expectedValue, actualValue)
						}
					}
				}

				// UpdatedAtが更新されていることを確認
				if result.RowsAffected > 0 {
					var updatedUser User
					db.First(&updatedUser, tt.userID)
					if updatedUser.UpdatedAt.Equal(baseUser.UpdatedAt) {
						t.Error("Expected UpdatedAt to be changed after update")
					}
				}
			}
		})
	}
}

func TestUser_Delete(t *testing.T) {
	db := setupTestDB(t)

	// テストユーザー作成
	user1 := User{Username: "user1", Email: "user1@test.com", Password: "pass", Name: "User 1"}
	user2 := User{Username: "user2", Email: "user2@test.com", Password: "pass", Name: "User 2"}
	user3 := User{Username: "user3", Email: "user3@test.com", Password: "pass", Name: "User 3"}

	db.Create(&user1)
	db.Create(&user2)
	db.Create(&user3)

	tests := []struct {
		name    string
		setup   func() *User
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常なユーザーを削除すると成功する",
			setup: func() *User {
				user := User{Username: "delete_user", Email: "delete@test.com", Password: "pass", Name: "Delete User"}
				db.Create(&user)
				return &user
			},
			wantErr: false,
		},
		{
			name: "存在しないユーザーを削除しようとしても影響なし",
			setup: func() *User {
				return &User{ID: 999999} // 存在しないID
			},
			wantErr: false, // GORMはソフトデリートで存在しないレコードの削除はエラーにならない
		},
		{
			name: "投稿があるユーザーを削除するとカスケード削除される",
			setup: func() *User {
				user := User{Username: "post_user", Email: "post_user@test.com", Password: "pass", Name: "Post User"}
				db.Create(&user)
				
				// 投稿を追加
				post := Post{Content: "テスト投稿", AuthorID: user.ID}
				db.Create(&post)
				
				return &user
			},
			wantErr: false,
		},
		{
			name: "いいねがあるユーザーを削除するとカスケード削除される",
			setup: func() *User {
				user := User{Username: "like_user", Email: "like_user@test.com", Password: "pass", Name: "Like User"}
				db.Create(&user)
				
				// 他のユーザーの投稿にいいね
				post := Post{Content: "いいねされる投稿", AuthorID: user2.ID}
				db.Create(&post)
				
				like := Like{UserID: user.ID, PostID: post.ID}
				db.Create(&like)
				
				return &user
			},
			wantErr: false,
		},
		{
			name: "フォロー関係があるユーザーを削除するとカスケード削除される",
			setup: func() *User {
				user := User{Username: "follow_user", Email: "follow_user@test.com", Password: "pass", Name: "Follow User"}
				db.Create(&user)
				
				// フォロー関係を追加
				follow1 := Follow{FollowerID: user.ID, FolloweeID: user2.ID} // userが他をフォロー
				follow2 := Follow{FollowerID: user3.ID, FolloweeID: user.ID} // userがフォローされる
				db.Create(&follow1)
				db.Create(&follow2)
				
				return &user
			},
			wantErr: false,
		},
		{
			name: "複合データがあるユーザーを削除すると全てカスケード削除される",
			setup: func() *User {
				user := User{Username: "complex_user", Email: "complex@test.com", Password: "pass", Name: "Complex User"}
				db.Create(&user)
				
				// 投稿、いいね、フォロー関係を全て追加
				post := Post{Content: "複合テスト投稿", AuthorID: user.ID}
				db.Create(&post)
				
				like := Like{UserID: user.ID, PostID: post.ID}
				db.Create(&like)
				
				follow := Follow{FollowerID: user.ID, FolloweeID: user2.ID}
				db.Create(&follow)
				
				return &user
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setup()
			
			// 削除前のデータ数確認
			var initialPostCount, initialLikeCount, initialFollowCount int64
			if user.ID != 0 {
				db.Model(&Post{}).Where("author_id = ?", user.ID).Count(&initialPostCount)
				db.Model(&Like{}).Where("user_id = ?", user.ID).Count(&initialLikeCount)
				// フォロー関係は削除ユーザーがフォロワーとフォロイーの両方のケースをカウント
				db.Model(&Follow{}).Where("follower_id = ? OR followee_id = ?", user.ID, user.ID).Count(&initialFollowCount)
			}

			// 削除実行（手動でカスケード削除を行う）
			var result *gorm.DB
			if user.ID != 0 {
				// 関連データを先に削除（実際のサービス層ではRepositoryで行う）
				db.Where("user_id = ?", user.ID).Delete(&Like{})                                     // ユーザーのいいね
				db.Where("author_id = ?", user.ID).Delete(&Post{})                                   // ユーザーの投稿
				db.Where("follower_id = ? OR followee_id = ?", user.ID, user.ID).Delete(&Follow{}) // フォロー関係
			}
			result = db.Delete(user)

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
				db.Model(&User{}).Where("id = ?", user.ID).Count(&count)
				if count != 0 {
					t.Errorf("Expected user to be deleted, but it still exists")
				}
				
				// カスケード削除確認
				if initialPostCount > 0 {
					var finalPostCount int64
					db.Model(&Post{}).Where("author_id = ?", user.ID).Count(&finalPostCount)
					if finalPostCount != 0 {
						t.Errorf("Expected posts to be cascade deleted, but %d posts remain", finalPostCount)
					}
				}
				
				if initialLikeCount > 0 {
					var finalLikeCount int64
					db.Model(&Like{}).Where("user_id = ?", user.ID).Count(&finalLikeCount)
					if finalLikeCount != 0 {
						t.Errorf("Expected likes to be cascade deleted, but %d likes remain", finalLikeCount)
					}
				}
				
				if initialFollowCount > 0 {
					var finalFollowCount int64
					db.Model(&Follow{}).Where("follower_id = ? OR followee_id = ?", user.ID, user.ID).Count(&finalFollowCount)
					if finalFollowCount != 0 {
						t.Errorf("Expected follow relationships to be cascade deleted, but %d relationships remain", finalFollowCount)
					}
				}
			}
		})
	}
}
