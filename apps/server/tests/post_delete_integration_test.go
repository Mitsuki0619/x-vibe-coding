package tests

import (
	"testing"

	"sns-server/internal/models"
	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestPostDeleteIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	// テスト開始前にデータをクリア
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
				name: "Test User" 
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
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
			},
		},
	}
	userResp := testutil.ExecuteGraphQLRequest(t, srv, userReq)
	testutil.AssertNoErrors(t, userResp)

	// テスト用投稿作成
	postReq := testutil.GraphQLRequest{
		Query: `mutation { 
			createPost(input: { 
				content: "削除テスト用投稿" 
			}) { 
				id 
				content 
				author { 
					username 
				} 
			} 
		}`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"content": "削除テスト用投稿",
			},
		},
	}
	postResp := testutil.ExecuteGraphQLRequest(t, srv, postReq)
	testutil.AssertNoErrors(t, postResp)

	postData := testutil.GetDataMap(t, postResp)
	post := testutil.GetMap(t, postData, "createPost")
	postID := uint(post["id"].(float64))

	t.Run("正常な投稿IDで削除すると成功する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				deletePost(input: { 
					postId: 1
				}) { 
					success
					message
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": postID,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		deleteResult := testutil.GetMap(t, data, "deletePost")

		if deleteResult["success"] != true {
			t.Error("Delete should succeed")
		}

		message := testutil.GetString(t, deleteResult, "message")
		if message == "" {
			t.Error("Delete message should not be empty")
		}

		// 投稿が削除されたことを確認
		var count int64
		db.Model(&models.Post{}).Where("id = ?", postID).Count(&count)
		if count != 0 {
			t.Errorf("Expected post to be deleted, but it still exists")
		}
	})

	t.Run("存在しない投稿IDで削除しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				deletePost(input: { 
					postId: 999999
				}) { 
					success
					message
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("投稿IDが0の場合エラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				deletePost(input: { 
					postId: 0
				}) { 
					success
					message
				} 
			}`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("いいねがついている投稿を削除するとカスケード削除される", func(t *testing.T) {
		// 新しい投稿作成
		postReq := testutil.GraphQLRequest{
			Query: `mutation { 
				createPost(input: { 
					content: "いいね付き投稿" 
				}) { 
					id 
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"content": "いいね付き投稿",
				},
			},
		}
		postResp := testutil.ExecuteGraphQLRequest(t, srv, postReq)
		testutil.AssertNoErrors(t, postResp)

		postData := testutil.GetDataMap(t, postResp)
		post := testutil.GetMap(t, postData, "createPost")
		newPostID := uint(post["id"].(float64))

		// いいねを追加
		likeReq := testutil.GraphQLRequest{
			Query: `mutation { 
				likePost(input: { 
					postId: 2
				}) { 
					id
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": newPostID,
				},
			},
		}
		testutil.ExecuteGraphQLRequest(t, srv, likeReq)

		// いいね数確認
		var initialLikeCount int64
		db.Model(&models.Like{}).Where("post_id = ?", newPostID).Count(&initialLikeCount)
		if initialLikeCount == 0 {
			t.Fatal("Like should be created before delete test")
		}

		// 投稿削除
		deleteReq := testutil.GraphQLRequest{
			Query: `mutation { 
				deletePost(input: { 
					postId: 2
				}) { 
					success
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": newPostID,
				},
			},
		}

		deleteResp := testutil.ExecuteGraphQLRequest(t, srv, deleteReq)
		testutil.AssertNoErrors(t, deleteResp)

		// カスケード削除確認
		var finalLikeCount int64
		db.Model(&models.Like{}).Where("post_id = ?", newPostID).Count(&finalLikeCount)
		if finalLikeCount != 0 {
			t.Errorf("Expected likes to be cascade deleted, but %d likes remain", finalLikeCount)
		}
	})
}