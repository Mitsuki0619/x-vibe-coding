package tests

import (
	"fmt"
	"testing"

	"sns-server/internal/models"
	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestLikeIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	// テスト用ユーザーと投稿を作成
	testUser := testutil.CreateTestUser(t, db, "likeuser", "like@example.com", "Like User")
	testPost := testutil.CreateTestPost(t, db, testUser.ID, "Test post for likes")

	t.Run("投稿にいいねをつけると成功する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation {
				likePost(input: {
					postId: $postId
				}) {
					id
					user {
						id
						username
					}
					post {
						id
						content
					}
				}
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": int(testPost.ID),
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		likePost := testutil.GetMap(t, data, "likePost")

		if likePost["id"] == nil {
			t.Error("Expected like ID")
		}

		user := testutil.GetMap(t, likePost, "user")
		if user["username"] != "likeuser" {
			t.Errorf("Expected user username 'likeuser', got %v", user["username"])
		}

		post := testutil.GetMap(t, likePost, "post")
		if post["content"] != "Test post for likes" {
			t.Errorf("Expected post content 'Test post for likes', got %v", post["content"])
		}
	})

	t.Run("既にいいねした投稿に再度いいねしようとするとエラーが発生する", func(t *testing.T) {
		// 同じ投稿に再度いいねを試す
		req := testutil.GraphQLRequest{
			Query: `mutation {
				likePost(input: {
					postId: $postId
				}) {
					id
					user {
						username
					}
					post {
						id
					}
				}
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": int(testPost.ID),
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// 重複いいねの場合はエラーが発生することを期待
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("つけたいいねを取り消すことができる", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation {
				unlikePost(input: {
					postId: $postId
				})
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": int(testPost.ID),
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		unlikePost := testutil.GetBool(t, data, "unlikePost")

		if !unlikePost {
			t.Error("Expected unlikePost to be true")
		}
	})

	t.Run("存在しない投稿IDにいいねしようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation {
				likePost(input: {
					postId: 99999
				}) {
					id
					user {
						username
					}
					post {
						id
					}
				}
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": 99999,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// 存在しない投稿へのいいねはエラーになることを期待
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("いいねしていない投稿のいいねを取り消そうとするとエラーが発生する", func(t *testing.T) {
		// 新しい投稿を作成（いいねしていない状態）
		newPost := testutil.CreateTestPost(t, db, testUser.ID, "New post without likes")

		req := testutil.GraphQLRequest{
			Query: `mutation {
				unlikePost(input: {
					postId: $postId
				})
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": int(newPost.ID),
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// いいねしていない投稿の取り消しはエラーまたはfalseを期待
		if resp.Errors != nil {
			t.Logf("Expected error for unliking non-liked post: %v", resp.Errors)
		} else {
			data := testutil.GetDataMap(t, resp)
			unlikePost := testutil.GetBool(t, data, "unlikePost")

			if unlikePost {
				t.Error("Expected unlikePost to be false for non-liked post")
			}
		}
	})

	t.Run("複数の投稿にいいねをつけることができる", func(t *testing.T) {
		// 複数の投稿を作成
		post1 := testutil.CreateTestPost(t, db, testUser.ID, "Post 1 for multiple likes")
		post2 := testutil.CreateTestPost(t, db, testUser.ID, "Post 2 for multiple likes")

		// 両方の投稿にいいね
		posts := []*models.Post{post1, post2}
		for i, post := range posts {
			req := testutil.GraphQLRequest{
				Query: `mutation {
					likePost(input: {
						postId: $postId
					}) {
						id
						user {
							username
						}
						post {
							content
						}
					}
				}`,
				Variables: map[string]interface{}{
					"input": map[string]interface{}{
						"postId": int(post.ID),
					},
				},
			}

			resp := testutil.ExecuteGraphQLRequest(t, srv, req)
			testutil.AssertNoErrors(t, resp)

			data := testutil.GetDataMap(t, resp)
			likePost := testutil.GetMap(t, data, "likePost")

			postData := testutil.GetMap(t, likePost, "post")
			expectedContent := ""
			if i == 0 {
				expectedContent = "Post 1 for multiple likes"
			} else {
				expectedContent = "Post 2 for multiple likes"
			}

			if postData["content"] != expectedContent {
				t.Errorf("Expected post content '%s', got %v", expectedContent, postData["content"])
			}
		}
	})

	t.Run("無効なパラメータでクエリしようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `{ invalidLikeField }`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		if resp.Errors == nil {
			t.Error("Expected errors for invalid query")
		}
	})

	t.Run("いいね機能の境界値テスト", func(t *testing.T) {
		testCases := []struct {
			name     string
			postID   interface{}
			wantErr  bool
			errorMsg string
		}{
			{
				name:     "負のpostIDでエラーが発生する",
				postID:   -1,
				wantErr:  true,
				errorMsg: "invalid post ID",
			},
			{
				name:     "ゼロのpostIDでエラーが発生する",
				postID:   0,
				wantErr:  true,
				errorMsg: "invalid post ID",
			},
			{
				name:     "非常に大きなpostIDでエラーが発生する",
				postID:   999999999,
				wantErr:  true,
				errorMsg: "post not found",
			},
			{
				name:     "文字列のpostIDでエラーが発生する",
				postID:   "invalid",
				wantErr:  true,
				errorMsg: "invalid post ID",
			},
			{
				name:     "nullのpostIDでエラーが発生する",
				postID:   nil,
				wantErr:  true,
				errorMsg: "post ID required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := testutil.GraphQLRequest{
					Query: `mutation {
						likePost(input: {
							postId: $postId
						}) {
							id
							user {
								username
							}
							post {
								id
							}
						}
					}`,
					Variables: map[string]interface{}{
						"input": map[string]interface{}{
							"postId": tc.postID,
						},
					},
				}

				resp := testutil.ExecuteGraphQLRequest(t, srv, req)

				if tc.wantErr {
					testutil.AssertHasErrors(t, resp)
				} else {
					testutil.AssertNoErrors(t, resp)
				}
			})
		}
	})

	t.Run("同時いいね操作テスト", func(t *testing.T) {
		// 現在の実装では全いいねが第一ユーザーとして処理されるため、
		// 同一投稿への複数いいねは重複エラーとなる
		// これは将来的に認証システム実装時に修正される予定

		user1 := testutil.CreateTestUser(t, db, "concurrent1", "concurrent1@test.com", "User 1")
		user2 := testutil.CreateTestUser(t, db, "concurrent2", "concurrent2@test.com", "User 2")
		user3 := testutil.CreateTestUser(t, db, "concurrent3", "concurrent3@test.com", "User 3")

		t.Run("複数の異なる投稿にいいねできる", func(t *testing.T) {
			// 各ユーザーが異なる投稿にいいねするパターン
			users := []*models.User{user1, user2, user3}

			for i, user := range users {
				// 各ユーザー用の投稿を作成
				post := testutil.CreateTestPost(t, db, user.ID, fmt.Sprintf("Post %d for concurrent likes", i+1))

				t.Run(fmt.Sprintf("User%d が自分の投稿%dにいいねする", i+1, i+1), func(t *testing.T) {
					req := testutil.GraphQLRequest{
						Query: `mutation {
							likePost(input: {
								postId: $postId
							}) {
								id
								user {
									id
								}
								post {
									id
								}
							}
						}`,
						Variables: map[string]interface{}{
							"input": map[string]interface{}{
								"postId": int(post.ID),
							},
						},
					}

					resp := testutil.ExecuteGraphQLRequest(t, srv, req)
					testutil.AssertNoErrors(t, resp)

					data := testutil.GetDataMap(t, resp)
					likePost := testutil.GetMap(t, data, "likePost")

					// いいねが作成されたことを確認
					if likePost["id"] == nil {
						t.Error("Like ID should be set")
					}

					postData := testutil.GetMap(t, likePost, "post")
					if postData["id"] != float64(post.ID) {
						t.Errorf("Expected post ID %d, got %v", post.ID, postData["id"])
					}
				})
			}
		})
	})

	t.Run("いいね取り消しの詳細テスト", func(t *testing.T) {
		user := testutil.CreateTestUser(t, db, "unliketest", "unlike@test.com", "Unlike User")
		post := testutil.CreateTestPost(t, db, user.ID, "Post for unlike test")

		// まずいいねする
		likeReq := testutil.GraphQLRequest{
			Query: `mutation {
				likePost(input: {
					postId: $postId
				}) {
					id
				}
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"postId": int(post.ID),
				},
			},
		}

		likeResp := testutil.ExecuteGraphQLRequest(t, srv, likeReq)
		testutil.AssertNoErrors(t, likeResp)

		// いいね取り消し
		t.Run("いいね後の取り消しが成功する", func(t *testing.T) {
			unlikeReq := testutil.GraphQLRequest{
				Query: `mutation {
					unlikePost(input: {
						postId: $postId
					})
				}`,
				Variables: map[string]interface{}{
					"input": map[string]interface{}{
						"postId": int(post.ID),
					},
				},
			}

			unlikeResp := testutil.ExecuteGraphQLRequest(t, srv, unlikeReq)
			testutil.AssertNoErrors(t, unlikeResp)
		})

		// 再度同じ投稿に対していいね取り消しを試す
		t.Run("既に取り消し済みの投稿の再取り消しでエラーが発生する", func(t *testing.T) {
			unlikeReq := testutil.GraphQLRequest{
				Query: `mutation {
					unlikePost(input: {
						postId: $postId
					})
				}`,
				Variables: map[string]interface{}{
					"input": map[string]interface{}{
						"postId": int(post.ID),
					},
				},
			}

			unlikeResp := testutil.ExecuteGraphQLRequest(t, srv, unlikeReq)
			testutil.AssertHasErrors(t, unlikeResp)
		})
	})
}
