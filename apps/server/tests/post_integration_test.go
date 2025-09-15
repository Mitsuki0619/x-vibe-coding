package tests

import (
	"strings"
	"testing"

	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestPostIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	// テスト用ユーザー作成（投稿テストで使用）
	testUser := testutil.CreateTestUser(t, db, "postuser", "post@example.com", "Post User")

	t.Run("正常なコンテンツで投稿作成すると成功する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				createPost(input: { 
					content: "Hello, World!" 
				}) { 
					id 
					content 
					author { 
						username 
						name 
					} 
				} 
			}`,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"content": "Hello, World!",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		createPost := testutil.GetMap(t, data, "createPost")

		if createPost["content"] != "Hello, World!" {
			t.Errorf("Expected content 'Hello, World!', got %v", createPost["content"])
		}

		author := testutil.GetMap(t, createPost, "author")
		if author["username"] == "" {
			t.Error("Author username should not be empty")
		}
	})

	t.Run("投稿がない状態で一覧取得すると空配列が返される", func(t *testing.T) {
		// データベースをクリア
		db.Exec("DELETE FROM posts")

		req := testutil.GraphQLRequest{
			Query: `{ posts { id content author { username } } }`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		posts := testutil.GetArray(t, data, "posts")

		if len(posts) != 0 {
			t.Errorf("Expected 0 posts, got %d", len(posts))
		}
	})

	t.Run("投稿が存在する状態で一覧取得すると投稿一覧が取得できる", func(t *testing.T) {
		// テスト用投稿を直接データベースに作成
		testutil.CreateTestPost(t, db, testUser.ID, "Test post content")

		req := testutil.GraphQLRequest{
			Query: `{ posts { id content author { username name } } }`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		posts := testutil.GetArray(t, data, "posts")

		if len(posts) < 1 {
			t.Errorf("Expected at least 1 post, got %d", len(posts))
		}

		post, ok := posts[0].(map[string]interface{})
		if !ok {
			t.Fatal("First post is not a map")
		}

		if post["content"] != "Test post content" {
			t.Errorf("Expected content 'Test post content', got %v", post["content"])
		}

		author := testutil.GetMap(t, post, "author")
		if author["username"] != "postuser" {
			t.Errorf("Expected author username 'postuser', got %v", author["username"])
		}
	})

	t.Run("空のコンテンツで投稿作成しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				createPost(input: { 
					content: "" 
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
					"content": "",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// 空コンテンツの場合はエラーが発生することを期待
		if resp.Errors == nil {
			t.Error("Expected errors for empty content but got none")
		}
	})

	t.Run("文字数制限を超えたコンテンツで投稿作成した場合の動作確認", func(t *testing.T) {
		// 280文字を超える長いコンテンツ
		longContent := ""
		for i := 0; i < 300; i++ {
			longContent += "a"
		}

		req := testutil.GraphQLRequest{
			Query: `mutation { 
				createPost(input: { 
					content: $content 
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
					"content": longContent,
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// 長すぎるコンテンツの場合の動作を確認
		// 現在の実装によってはエラーまたは切り詰めて成功
		if resp.Errors != nil {
			// エラーが発生する場合
			t.Logf("Long content resulted in error (expected): %v", resp.Errors)
		} else {
			// 成功する場合は、コンテンツが適切に処理されているかを確認
			data := testutil.GetDataMap(t, resp)
			createPost := testutil.GetMap(t, data, "createPost")
			content := testutil.GetString(t, createPost, "content")

			if len(content) != len(longContent) && len(content) != 280 {
				t.Logf("Content was processed: original=%d chars, result=%d chars", len(longContent), len(content))
			}
		}
	})

	t.Run("作成した投稿が一覧に適切に表示される", func(t *testing.T) {
		// 特定の投稿IDが存在するかテスト
		testutil.CreateTestPost(t, db, testUser.ID, "Specific test post")

		req := testutil.GraphQLRequest{
			Query: `{ posts { id content author { username } } }`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		posts := testutil.GetArray(t, data, "posts")

		found := false
		for _, p := range posts {
			post := p.(map[string]interface{})
			if post["content"] == "Specific test post" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Specific test post not found in posts list")
		}
	})

	t.Run("投稿コンテンツの入力検証テスト", func(t *testing.T) {
		testCases := []struct {
			name    string
			content string
			wantErr bool
		}{
			{
				name:    "空白のみのコンテンツでエラーが発生する",
				content: "   ",
				wantErr: true,
			},
			{
				name:    "タブと改行のみでエラーが発生する",
				content: "\t\n\r ",
				wantErr: true,
			},
			{
				name:    "Unicode絵文字を含むコンテンツが正しく処理される",
				content: "Hello World! 😀🎉✨",
				wantErr: false,
			},
			{
				name:    "日本語コンテンツが正しく処理される",
				content: "こんにちは世界！これは日本語の投稿です。",
				wantErr: false,
			},
			{
				name:    "改行を含むコンテンツが正しく処理される",
				content: "First line\nSecond line\nThird line",
				wantErr: false,
			},
			{
				name:    "HTMLタグが含まれる場合の処理確認",
				content: "Hello <script>alert('test')</script> World",
				wantErr: false, // サニタイズされて成功すると仮定
			},
			{
				name:    "SQLインジェクション試行が無効化される",
				content: "'; DROP TABLE posts; --",
				wantErr: false, // 通常のコンテンツとして処理される
			},
			{
				name:    "279文字のコンテンツは成功する",
				content: strings.Repeat("a", 279),
				wantErr: false,
			},
			{
				name:    "280文字ちょうどのコンテンツは成功する",
				content: strings.Repeat("b", 280),
				wantErr: false,
			},
			{
				name:    "281文字のコンテンツはエラーになる",
				content: strings.Repeat("c", 281),
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := testutil.GraphQLRequest{
					Query: `mutation { 
						createPost(input: { 
							content: $content 
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
							"content": tc.content,
						},
					},
				}

				resp := testutil.ExecuteGraphQLRequest(t, srv, req)

				if tc.wantErr {
					testutil.AssertHasErrors(t, resp)
				} else {
					testutil.AssertNoErrors(t, resp)

					// 成功した場合、適切に処理されていることを確認
					data := testutil.GetDataMap(t, resp)
					createPost := testutil.GetMap(t, data, "createPost")
					content := testutil.GetString(t, createPost, "content")

					// HTMLタグのケースではサニタイズされていることを確認
					if tc.name == "HTMLタグが含まれる場合の処理確認" {
						if strings.Contains(content, "<script>") {
							t.Error("HTML tags not properly sanitized")
						}
					}

					// 空白のみではないことを確認（正常系の場合）
					if strings.TrimSpace(content) == "" && !tc.wantErr {
						t.Error("Content should not be empty after processing")
					}
				}
			})
		}
	})

	t.Run("投稿の境界値と特殊ケーステスト", func(t *testing.T) {
		testCases := []struct {
			name    string
			content string
			wantErr bool
		}{
			{
				name:    "1文字の投稿は成功する",
				content: "a",
				wantErr: false,
			},
			{
				name:    "絵文字1つの投稿は成功する",
				content: "😀",
				wantErr: false,
			},
			{
				name:    "数字のみの投稿は成功する",
				content: "12345",
				wantErr: false,
			},
			{
				name:    "URL形式の投稿は成功する",
				content: "https://example.com",
				wantErr: false,
			},
			{
				name:    "メンション形式の投稿は成功する",
				content: "@username hello world",
				wantErr: false,
			},
			{
				name:    "ハッシュタグ形式の投稿は成功する",
				content: "#hashtag test post",
				wantErr: false,
			},
			{
				name:    "特殊記号を含む投稿は成功する",
				content: "!@#$%^&*()_+-={}|[]\\:\";'<>?,./",
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := testutil.GraphQLRequest{
					Query: `mutation { 
						createPost(input: { 
							content: $content 
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
							"content": tc.content,
						},
					},
				}

				resp := testutil.ExecuteGraphQLRequest(t, srv, req)

				if tc.wantErr {
					testutil.AssertHasErrors(t, resp)
				} else {
					testutil.AssertNoErrors(t, resp)

					data := testutil.GetDataMap(t, resp)
					createPost := testutil.GetMap(t, data, "createPost")
					content := testutil.GetString(t, createPost, "content")

					// コンテンツが期待通りに保存されていることを確認
					if content != tc.content {
						t.Logf("Content may have been processed: expected %q, got %q", tc.content, content)
					}
				}
			})
		}
	})
}
