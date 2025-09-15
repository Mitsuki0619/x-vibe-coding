package tests

import (
	"testing"

	"sns-server/internal/models"
	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestFollowUnfollowSequence(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	// データベースをクリア
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

	t.Run("フォローして解除する一連の流れが正しく動作する", func(t *testing.T) {
		// 最初にフォロー数を確認
		var initialCount int64
		db.Model(&models.Follow{}).Count(&initialCount)
		if initialCount != 0 {
			t.Fatalf("Expected 0 initial follows, got %d", initialCount)
		}

		// 1. フォローする
		followReq := testutil.GraphQLRequest{
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

		followResp := testutil.ExecuteGraphQLRequest(t, srv, followReq)
		testutil.AssertNoErrors(t, followResp)

		followData := testutil.GetDataMap(t, followResp)
		follow := testutil.GetMap(t, followData, "followUser")

		if follow["id"] == nil {
			t.Error("Follow ID should be set")
		}

		// フォロー数が1になったことを確認
		var followCount int64
		db.Model(&models.Follow{}).Count(&followCount)
		if followCount != 1 {
			t.Fatalf("Expected 1 follow after following, got %d", followCount)
		}

		// 2. フォロー解除する
		unfollowReq := testutil.GraphQLRequest{
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

		unfollowResp := testutil.ExecuteGraphQLRequest(t, srv, unfollowReq)
		testutil.AssertNoErrors(t, unfollowResp)

		unfollowData := testutil.GetDataMap(t, unfollowResp)
		unfollow := testutil.GetMap(t, unfollowData, "unfollowUser")

		if unfollow["success"] != true {
			t.Error("Unfollow should succeed")
		}

		message := testutil.GetString(t, unfollow, "message")
		if message == "" {
			t.Error("Unfollow message should not be empty")
		}

		// フォロー数が0に戻ったことを確認
		var finalCount int64
		db.Model(&models.Follow{}).Count(&finalCount)
		if finalCount != 0 {
			t.Fatalf("Expected 0 follows after unfollowing, got %d", finalCount)
		}
	})
}