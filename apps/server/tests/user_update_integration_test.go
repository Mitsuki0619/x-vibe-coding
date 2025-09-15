package tests

import (
	"testing"

	"sns-server/internal/models"
	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestUserUpdateIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	// テスト開始前にデータをクリア
	db.Exec("DELETE FROM follows")
	db.Exec("DELETE FROM likes")
	db.Exec("DELETE FROM posts")
	db.Exec("DELETE FROM users")

	// テスト用ユーザー登録
	userReq := testutil.GraphQLRequest{
		Query: `mutation { 
			register(input: { 
				username: "testuser", 
				email: "test@example.com", 
				password: "password123", 
				name: "Test User",
				bio: "Original Bio"
			}) { 
				user { 
					id 
					username
					email
					name
					bio
				} 
			} 
		}`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
				"bio":      "Original Bio",
			},
		},
	}
	userResp := testutil.ExecuteGraphQLRequest(t, srv, userReq)
	testutil.AssertNoErrors(t, userResp)

	userData := testutil.GetDataMap(t, userResp)
	userInfo := testutil.GetMap(t, userData, "register")
	userDetailInfo := testutil.GetMap(t, userInfo, "user")
	testUserID := uint(userDetailInfo["id"].(float64))

	t.Run("正常なデータでユーザー更新すると成功する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					name: "Updated Name",
					bio: "Updated Bio"
				}) { 
					success
					message
					user {
						id
						name
						bio
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": testUserID,
					"name":   "Updated Name",
					"bio":    "Updated Bio",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		updateResult := testutil.GetMap(t, data, "updateUser")

		if updateResult["success"] != true {
			t.Error("User update should succeed")
		}

		user := testutil.GetMap(t, updateResult, "user")
		if user["name"] != "Updated Name" {
			t.Errorf("Expected name to be 'Updated Name', got %v", user["name"])
		}
		if user["bio"] != "Updated Bio" {
			t.Errorf("Expected bio to be 'Updated Bio', got %v", user["bio"])
		}
	})

	t.Run("ユーザー名を更新できる", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					username: "updated_username"
				}) { 
					success
					user {
						username
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId":   testUserID,
					"username": "updated_username",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		updateResult := testutil.GetMap(t, data, "updateUser")
		user := testutil.GetMap(t, updateResult, "user")

		if user["username"] != "updated_username" {
			t.Errorf("Expected username to be 'updated_username', got %v", user["username"])
		}
	})

	t.Run("メールアドレスを更新できる", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					email: "updated@example.com"
				}) { 
					success
					user {
						email
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": testUserID,
					"email":  "updated@example.com",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		updateResult := testutil.GetMap(t, data, "updateUser")
		user := testutil.GetMap(t, updateResult, "user")

		if user["email"] != "updated@example.com" {
			t.Errorf("Expected email to be 'updated@example.com', got %v", user["email"])
		}
	})

	t.Run("複数フィールドを同時に更新できる", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					username: "multi_update",
					name: "Multi Update Name",
					bio: "Multi Update Bio"
				}) { 
					success
					user {
						username
						name
						bio
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId":   testUserID,
					"username": "multi_update",
					"name":     "Multi Update Name",
					"bio":      "Multi Update Bio",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		updateResult := testutil.GetMap(t, data, "updateUser")
		user := testutil.GetMap(t, updateResult, "user")

		if user["username"] != "multi_update" {
			t.Errorf("Expected username to be 'multi_update', got %v", user["username"])
		}
		if user["name"] != "Multi Update Name" {
			t.Errorf("Expected name to be 'Multi Update Name', got %v", user["name"])
		}
		if user["bio"] != "Multi Update Bio" {
			t.Errorf("Expected bio to be 'Multi Update Bio', got %v", user["bio"])
		}
	})

	t.Run("存在しないユーザーIDで更新しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 999999,
					name: "Should Fail"
				}) { 
					success
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("空のユーザー名で更新しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					username: ""
				}) { 
					success
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId":   testUserID,
					"username": "",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("空のメールアドレスで更新しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					email: ""
				}) { 
					success
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": testUserID,
					"email":  "",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("無効なメールアドレス形式で更新しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					email: "invalid-email-format"
				}) { 
					success
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": testUserID,
					"email":  "invalid-email-format",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("重複するユーザー名で更新しようとするとエラーが発生する", func(t *testing.T) {
		// 別のユーザーを作成
		duplicateUserReq := testutil.GraphQLRequest{
			Query: `mutation { 
				register(input: { 
					username: "duplicate_user", 
					email: "duplicate@example.com", 
					password: "password456", 
					name: "Duplicate User" 
				}) { 
					user { 
						id 
						username
					} 
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"username": "duplicate_user",
					"email":    "duplicate@example.com",
					"password": "password456",
					"name":     "Duplicate User",
				},
			},
		}
		duplicateResp := testutil.ExecuteGraphQLRequest(t, srv, duplicateUserReq)
		testutil.AssertNoErrors(t, duplicateResp)

		// 作成されたユーザーを確認
		dupData := testutil.GetDataMap(t, duplicateResp)
		dupUser := testutil.GetMap(t, dupData, "register")
		dupUserInfo := testutil.GetMap(t, dupUser, "user")
		
		if dupUserInfo["username"] != "duplicate_user" {
			t.Fatalf("Failed to create duplicate test user: expected 'duplicate_user', got %v", dupUserInfo["username"])
		}

		// 既存のユーザー名で更新を試行
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					username: "duplicate_user"
				}) { 
					success
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId":   testUserID,
					"username": "duplicate_user",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("SQLインジェクション試行が安全に処理される", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 1,
					name: "'; DROP TABLE users; --"
				}) { 
					success
					user {
						name
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": testUserID,
					"name":   "'; DROP TABLE users; --",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		// テーブルが削除されていないことを確認
		var count int64
		db.Model(&models.User{}).Count(&count)
		if count == 0 {
			t.Error("Users table should not be dropped by SQL injection attempt")
		}

		// 名前が安全に保存されていることを確認
		data := testutil.GetDataMap(t, resp)
		updateResult := testutil.GetMap(t, data, "updateUser")
		user := testutil.GetMap(t, updateResult, "user")
		if user["name"] != "'; DROP TABLE users; --" {
			t.Errorf("Expected name to be stored safely, got %v", user["name"])
		}
	})

	t.Run("ユーザーIDが0の場合エラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				updateUser(input: { 
					userId: 0,
					name: "Should Fail"
				}) { 
					success
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})
}