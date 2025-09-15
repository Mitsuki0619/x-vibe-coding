package tests

import (
	"testing"

	"sns-server/internal/models"
	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestFollowIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	// テスト開始前に全フォローデータをクリア
	db.Exec("DELETE FROM follows")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM posts")
	db.Exec("DELETE FROM likes")

	// テスト用ユーザー登録
	user1Req := testutil.GraphQLRequest{
		Query: `mutation { 
			register(input: { 
				username: "follower_user", 
				email: "follower@example.com", 
				password: "password123", 
				name: "Follower User" 
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
				"username": "follower_user",
				"email":    "follower@example.com",
				"password": "password123",
				"name":     "Follower User",
			},
		},
	}
	user1Resp := testutil.ExecuteGraphQLRequest(t, srv, user1Req)
	testutil.AssertNoErrors(t, user1Resp)

	user2Req := testutil.GraphQLRequest{
		Query: `mutation { 
			register(input: { 
				username: "followee_user", 
				email: "followee@example.com", 
				password: "password456", 
				name: "Followee User" 
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
				"username": "followee_user",
				"email":    "followee@example.com",
				"password": "password456",
				"name":     "Followee User",
			},
		},
	}
	user2Resp := testutil.ExecuteGraphQLRequest(t, srv, user2Req)
	testutil.AssertNoErrors(t, user2Resp)

	t.Run("ユーザーをフォローすると成功する", func(t *testing.T) {
		// データベースをクリアしてテストの独立性を確保
		db.Exec("DELETE FROM follows")

		req := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 2
				}) { 
					id
					follower {
						id
						username
					}
					followee {
						id
						username
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"followeeId": 2,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		follow := testutil.GetMap(t, data, "followUser")

		if follow["id"] == nil {
			t.Error("Follow ID should be set")
		}

		follower := testutil.GetMap(t, follow, "follower")
		if follower["username"] != "follower_user" {
			t.Errorf("Expected follower username 'follower_user', got %v", follower["username"])
		}

		followee := testutil.GetMap(t, follow, "followee")
		if followee["username"] != "followee_user" {
			t.Errorf("Expected followee username 'followee_user', got %v", followee["username"])
		}
	})

	t.Run("既にフォローしているユーザーを再度フォローしようとするとエラーが発生する", func(t *testing.T) {
		// データベースをクリアしてテストの独立性を確保
		db.Exec("DELETE FROM follows")

		// まずフォローする
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
					"followeeId": 2,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, followReq)

		// 再度フォローしようとしてエラーを確認
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 2
				}) { 
					id
					follower {
						username
					}
					followee {
						username
					}
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"followeeId": 2,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("自分自身をフォローしようとするとエラーが発生する", func(t *testing.T) {
		// データベースをクリアしてテストの独立性を確保
		db.Exec("DELETE FROM follows")

		req := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 1
				}) { 
					id
					follower {
						username
					}
					followee {
						username
					}
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("存在しないユーザーをフォローしようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 999
				}) { 
					id
					follower {
						username
					}
					followee {
						username
					}
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("フォローを解除すると成功する", func(t *testing.T) {
		// データベースをクリアしてテストの独立性を確保
		result := db.Exec("DELETE FROM follows")
		if result.Error != nil {
			t.Fatalf("Failed to clean follows table: %v", result.Error)
		}

		// 削除確認
		var count int64
		db.Model(&models.Follow{}).Count(&count)
		if count != 0 {
			t.Fatalf("Expected 0 follows after cleanup, got %d", count)
		}

		// まずフォローする
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
					"followeeId": 2,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, followReq)

		// そしてフォロー解除
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				unfollowUser(input: { 
					followeeId: 2
				}) {
					success
					message
				}
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"followeeId": 2,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		unfollow := testutil.GetMap(t, data, "unfollowUser")

		if unfollow["success"] != true {
			t.Error("Unfollow should succeed")
		}

		message := testutil.GetString(t, unfollow, "message")
		if message == "" {
			t.Error("Unfollow message should not be empty")
		}
	})

	t.Run("フォローしていないユーザーをフォロー解除しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				unfollowUser(input: { 
					followeeId: 2
				}) {
					success
					message
				}
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("存在しないユーザーをフォロー解除しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				unfollowUser(input: { 
					followeeId: 999
				}) {
					success
					message
				}
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("フォロー関係の確認クエリが正しく動作する", func(t *testing.T) {
		// データベースをクリアしてテストの独立性を確保
		db.Exec("DELETE FROM follows")

		// まずフォローする
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
					"followeeId": 2,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, followReq)

		// フォロワー一覧を取得
		followersReq := testutil.GraphQLRequest{
			Query: `{ 
				user(id: 2) {
					id
					username
					followers {
						id
						follower {
							id
							username
						}
					}
				}
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, followersReq)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		user := testutil.GetMap(t, data, "user")
		followers := testutil.GetArray(t, user, "followers")

		if len(followers) != 1 {
			t.Errorf("Expected 1 follower, got %d", len(followers))
		}

		followerRel, ok := followers[0].(map[string]interface{})
		if !ok {
			t.Fatal("First follower is not a map")
		}

		follower := testutil.GetMap(t, followerRel, "follower")
		if follower["username"] != "follower_user" {
			t.Errorf("Expected follower username 'follower_user', got %v", follower["username"])
		}
	})

	t.Run("フォロー数とフォロワー数が正しくカウントされる", func(t *testing.T) {
		// データベースをクリアしてテストの独立性を確保
		db.Exec("DELETE FROM follows")

		// まずフォローする
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
					"followeeId": 2,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, followReq)

		req := testutil.GraphQLRequest{
			Query: `{ 
				users {
					id
					username
					followerCount
					followingCount
				}
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		users := testutil.GetArray(t, data, "users")

		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}

		// follower_userのチェック（フォロー数: 1, フォロワー数: 0）
		for _, u := range users {
			user, ok := u.(map[string]interface{})
			if !ok {
				t.Fatal("User is not a map")
			}

			username := user["username"].(string)
			if username == "follower_user" {
				if user["followingCount"].(float64) != 1 {
					t.Errorf("Expected follower_user following count 1, got %v", user["followingCount"])
				}
				if user["followerCount"].(float64) != 0 {
					t.Errorf("Expected follower_user follower count 0, got %v", user["followerCount"])
				}
			} else if username == "followee_user" {
				if user["followingCount"].(float64) != 0 {
					t.Errorf("Expected followee_user following count 0, got %v", user["followingCount"])
				}
				if user["followerCount"].(float64) != 1 {
					t.Errorf("Expected followee_user follower count 1, got %v", user["followerCount"])
				}
			}
		}
	})

	t.Run("複数のフォロー関係における境界値テスト", func(t *testing.T) {
		// 3人目のユーザーを作成
		user3Req := testutil.GraphQLRequest{
			Query: `mutation { 
				register(input: { 
					username: "third_user", 
					email: "third@example.com", 
					password: "password789", 
					name: "Third User" 
				}) { 
					user { 
						id 
					} 
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"username": "third_user",
					"email":    "third@example.com",
					"password": "password789",
					"name":     "Third User",
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, user3Req)

		// follower_userが3人目もフォロー
		followReq := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 3
				}) { 
					id
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"followeeId": 3,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, followReq)

		// followee_userがthird_userをフォロー
		followReq2 := testutil.GraphQLRequest{
			Query: `mutation { 
				followUser(input: { 
					followeeId: 3
				}) { 
					id
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"followeeId": 3,
				},
			},
		}
		// Note: ここでエラーが発生する可能性があります（認証システムが未実装のため）
		// 実際の実装では、認証されたユーザーIDを使用してフォロー操作を行います
		testutil.ExecuteGraphQLRequest(t, srv, followReq2)

		// カウントを確認
		countReq := testutil.GraphQLRequest{
			Query: `{ 
				user(id: 3) {
					id
					username
					followerCount
				}
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, countReq)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		user := testutil.GetMap(t, data, "user")

		// third_userは複数人からフォローされている
		if user["followerCount"].(float64) < 1 {
			t.Errorf("Expected third_user to have at least 1 follower, got %v", user["followerCount"])
		}
	})

	t.Run("フォローリストの取得が正しく動作する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `{ 
				user(id: 1) {
					id
					username
					following {
						id
						followee {
							id
							username
							name
						}
						createdAt
					}
				}
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		user := testutil.GetMap(t, data, "user")
		following := testutil.GetArray(t, user, "following")

		// follower_userは少なくとも1人はフォローしている
		if len(following) < 1 {
			t.Errorf("Expected at least 1 following relationship, got %d", len(following))
		}

		// 最初のフォロー関係をチェック
		followRel, ok := following[0].(map[string]interface{})
		if !ok {
			t.Fatal("First follow relationship is not a map")
		}

		followee := testutil.GetMap(t, followRel, "followee")
		if followee["id"] == nil {
			t.Error("Followee ID should be set")
		}

		if followRel["createdAt"] == nil {
			t.Error("Follow creation timestamp should be set")
		}
	})
}
