package tests

import (
	"testing"

	"sns-server/internal/models"
	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestUserDeleteIntegration(t *testing.T) {
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
	user1Req := testutil.GraphQLRequest{
		Query: `mutation { 
			register(input: { 
				username: "user1", 
				email: "user1@example.com", 
				password: "password123", 
				name: "User 1" 
			}) { 
				token 
				user { 
					id 
					username 
				} 
			} 
		}`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"username": "user1",
				"email":    "user1@example.com",
				"password": "password123",
				"name":     "User 1",
			},
		},
	}
	user1Resp := testutil.ExecuteGraphQLRequest(t, srv, user1Req)
	testutil.AssertNoErrors(t, user1Resp)

	user2Req := testutil.GraphQLRequest{
		Query: `mutation { 
			register(input: { 
				username: "user2", 
				email: "user2@example.com", 
				password: "password456", 
				name: "User 2" 
			}) { 
				token 
				user { 
					id 
					username 
				} 
			} 
		}`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"username": "user2",
				"email":    "user2@example.com",
				"password": "password456",
				"name":     "User 2",
			},
		},
	}
	user2Resp := testutil.ExecuteGraphQLRequest(t, srv, user2Req)
	testutil.AssertNoErrors(t, user2Resp)

	// ユーザー1のIDを取得
	user1Data := testutil.GetDataMap(t, user1Resp)
	user1Info := testutil.GetMap(t, user1Data, "register")
	user1UserInfo := testutil.GetMap(t, user1Info, "user")
	user1ID := uint(user1UserInfo["id"].(float64))

	t.Run("正常なユーザーIDで削除すると成功する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				deleteUser(input: { 
					userId: 1
				}) { 
					success
					message
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": user1ID,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		deleteResult := testutil.GetMap(t, data, "deleteUser")

		if deleteResult["success"] != true {
			t.Error("User delete should succeed")
		}

		message := testutil.GetString(t, deleteResult, "message")
		if message == "" {
			t.Error("Delete message should not be empty")
		}

		// ユーザーが削除されたことを確認
		var count int64
		db.Model(&models.User{}).Where("id = ?", user1ID).Count(&count)
		if count != 0 {
			t.Errorf("Expected user to be deleted, but it still exists")
		}
	})

	t.Run("存在しないユーザーIDで削除しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				deleteUser(input: { 
					userId: 999999
				}) { 
					success
					message
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("ユーザーIDが0の場合エラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				deleteUser(input: { 
					userId: 0
				}) { 
					success
					message
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("複合データがあるユーザーを削除すると全てカスケード削除される", func(t *testing.T) {
		// 新しいユーザー作成
		userReq := testutil.GraphQLRequest{
			Query: `mutation { 
				register(input: { 
					username: "complex_user", 
					email: "complex@example.com", 
					password: "password789", 
					name: "Complex User" 
				}) { 
					user { 
						id 
					} 
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"username": "complex_user",
					"email":    "complex@example.com",
					"password": "password789",
					"name":     "Complex User",
				},
			},
		}
		userResp := testutil.ExecuteGraphQLRequest(t, srv, userReq)
		testutil.AssertNoErrors(t, userResp)

		userData := testutil.GetDataMap(t, userResp)
		userInfo := testutil.GetMap(t, userData, "register")
		userDetailInfo := testutil.GetMap(t, userInfo, "user")
		complexUserID := uint(userDetailInfo["id"].(float64))

		// 投稿作成
		postReq := testutil.GraphQLRequest{
			Query: `mutation { 
				createPost(input: { 
					content: "複合ユーザーの投稿" 
				}) { 
					id
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"content": "複合ユーザーの投稿",
				},
			},
		}
		postResp := testutil.ExecuteGraphQLRequest(t, srv, postReq)
		testutil.AssertNoErrors(t, postResp)

		postData := testutil.GetDataMap(t, postResp)
		post := testutil.GetMap(t, postData, "createPost")
		postID := uint(post["id"].(float64))

		// いいね追加
		likeReq := testutil.GraphQLRequest{
			Query: `mutation { 
				likePost(input: { 
					postId: 1
				}) { 
					id
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": postID,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, likeReq)

		// フォロー関係追加
		followReq := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 2
				}) { 
					id
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"followeeId": 2, // user2をフォロー
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, followReq)

		// 削除前のデータ数確認
		var initialPostCount, initialLikeCount, initialFollowCount int64
		db.Model(&models.Post{}).Where("author_id = ?", complexUserID).Count(&initialPostCount)
		db.Model(&models.Like{}).Where("user_id = ?", complexUserID).Count(&initialLikeCount)
		db.Model(&models.Follow{}).Where("follower_id = ? OR followee_id = ?", complexUserID, complexUserID).Count(&initialFollowCount)

		if initialPostCount == 0 && initialLikeCount == 0 && initialFollowCount == 0 {
			t.Log("Warning: No associated data found for cascade delete test")
		}

		// ユーザー削除
		deleteReq := testutil.GraphQLRequest{
			Query: `mutation { 
				deleteUser(input: { 
					userId: 1
				}) { 
					success
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"userId": complexUserID,
				},
			},
		}

		deleteResp := testutil.ExecuteGraphQLRequest(t, srv, deleteReq)
		testutil.AssertNoErrors(t, deleteResp)

		// カスケード削除確認
		var finalPostCount, finalLikeCount, finalFollowCount int64
		db.Model(&models.Post{}).Where("author_id = ?", complexUserID).Count(&finalPostCount)
		db.Model(&models.Like{}).Where("user_id = ?", complexUserID).Count(&finalLikeCount)
		db.Model(&models.Follow{}).Where("follower_id = ? OR followee_id = ?", complexUserID, complexUserID).Count(&finalFollowCount)

		if finalPostCount != 0 {
			t.Errorf("Expected posts to be cascade deleted, but %d posts remain", finalPostCount)
		}
		if finalLikeCount != 0 {
			t.Errorf("Expected likes to be cascade deleted, but %d likes remain", finalLikeCount)
		}
		if finalFollowCount != 0 {
			t.Errorf("Expected follows to be cascade deleted, but %d follows remain", finalFollowCount)
		}
	})
}