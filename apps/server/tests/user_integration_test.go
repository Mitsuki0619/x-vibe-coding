package tests

import (
	"strings"
	"testing"

	"sns-server/internal/server"
	"sns-server/internal/testutil"
)

func TestUserIntegration(t *testing.T) {
	// テスト用データベースセットアップ
	db := testutil.SetupTestDB(t)

	// サーバーインスタンス作成
	srv := server.NewServer(db, nil)

	t.Run("初期状態でユーザー一覧を取得すると空配列が返される", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `{ users { id username name } }`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		users := testutil.GetArray(t, data, "users")

		if len(users) != 0 {
			t.Errorf("Expected 0 users, got %d", len(users))
		}
	})

	t.Run("正常なデータでユーザー登録するとトークンとユーザー情報が返される", func(t *testing.T) {
		req := testutil.GraphQLRequest{
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
						name 
						email 
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

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		register := testutil.GetMap(t, data, "register")

		// トークンの確認
		token := testutil.GetString(t, register, "token")
		if token == "" {
			t.Error("Token is empty")
		}

		// ユーザー情報の確認
		user := testutil.GetMap(t, register, "user")

		if user["username"] != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", user["username"])
		}

		if user["name"] != "Test User" {
			t.Errorf("Expected name 'Test User', got %v", user["name"])
		}

		if user["email"] != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %v", user["email"])
		}
	})

	t.Run("ユーザー登録後に一覧取得すると登録したユーザーが含まれる", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `{ users { id username name email } }`,
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)
		testutil.AssertNoErrors(t, resp)

		data := testutil.GetDataMap(t, resp)
		users := testutil.GetArray(t, data, "users")

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		user, ok := users[0].(map[string]interface{})
		if !ok {
			t.Fatal("First user is not a map")
		}

		if user["username"] != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", user["username"])
		}
	})

	t.Run("既存のユーザー名で登録しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				register(input: { 
					username: "testuser", 
					email: "another@example.com", 
					password: "password456", 
					name: "Another User" 
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
					"username": "testuser", // 既に存在するユーザー名
					"email":    "another@example.com",
					"password": "password456",
					"name":     "Another User",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// エラーが発生することを期待
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("既存のメールアドレスで登録しようとするとエラーが発生する", func(t *testing.T) {
		req := testutil.GraphQLRequest{
			Query: `mutation { 
				register(input: { 
					username: "newuser", 
					email: "test@example.com", 
					password: "password789", 
					name: "New User" 
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
					"username": "newuser",
					"email":    "test@example.com", // 既に存在するメールアドレス
					"password": "password789",
					"name":     "New User",
				},
			},
		}

		resp := testutil.ExecuteGraphQLRequest(t, srv, req)

		// エラーが発生することを期待
		testutil.AssertHasErrors(t, resp)
	})

	t.Run("必須フィールドが不足している場合登録でエラーが発生する", func(t *testing.T) {
		testCases := []struct {
			name  string
			input map[string]interface{}
		}{
			{
				name: "ユーザー名が空の場合",
				input: map[string]interface{}{
					"email":    "noname@example.com",
					"password": "password123",
					"name":     "No Name User",
				},
			},
			{
				name: "メールアドレスが空の場合",
				input: map[string]interface{}{
					"username": "nomail",
					"password": "password123",
					"name":     "No Mail User",
				},
			},
			{
				name: "パスワードが空の場合",
				input: map[string]interface{}{
					"username": "nopass",
					"email":    "nopass@example.com",
					"name":     "No Password User",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := testutil.GraphQLRequest{
					Query: `mutation { 
						register(input: $input) { 
							token 
							user { 
								id 
								username 
							} 
						} 
					}`,
					Variables: map[string]interface{}{
						"input": tc.input,
					},
				}

				resp := testutil.ExecuteGraphQLRequest(t, srv, req)

				// エラーが発生することを期待
				testutil.AssertHasErrors(t, resp)
			})
		}
	})

	t.Run("入力検証テスト", func(t *testing.T) {
		testCases := []struct {
			name  string
			input map[string]interface{}
		}{
			{
				name: "無効なメールアドレス形式の場合エラーが発生する",
				input: map[string]interface{}{
					"username": "validuser",
					"email":    "invalid-email",
					"password": "password123",
					"name":     "Valid User",
				},
			},
			{
				name: "メールアドレスに@がない場合エラーが発生する",
				input: map[string]interface{}{
					"username": "validuser2",
					"email":    "invalidemail.com",
					"password": "password123",
					"name":     "Valid User",
				},
			},
			{
				name: "ユーザー名が長すぎる場合エラーが発生する",
				input: map[string]interface{}{
					"username": "verylongusernamethatexceedslimitverylongusernamethatexceedslimit",
					"email":    "long@example.com",
					"password": "password123",
					"name":     "Long User",
				},
			},
			{
				name: "ユーザー名に特殊文字が含まれる場合エラーが発生する",
				input: map[string]interface{}{
					"username": "user@#$%",
					"email":    "special@example.com",
					"password": "password123",
					"name":     "Special User",
				},
			},
			{
				name: "名前が異常に長い場合エラーが発生する",
				input: map[string]interface{}{
					"username": "longname",
					"email":    "longname@example.com",
					"password": "password123",
					"name":     strings.Repeat("非常に長い名前", 50), // 異常に長い名前
				},
			},
			{
				name: "パスワードが短すぎる場合エラーが発生する",
				input: map[string]interface{}{
					"username": "shortpass",
					"email":    "shortpass@example.com",
					"password": "12", // 2文字のパスワード
					"name":     "Short Pass User",
				},
			},
			{
				name: "SQLインジェクション試行が無効化される",
				input: map[string]interface{}{
					"username": "'; DROP TABLE users; --",
					"email":    "sql@injection.com",
					"password": "password123",
					"name":     "SQL Injection",
				},
			},
			{
				name: "XSS試行が無効化される",
				input: map[string]interface{}{
					"username": "xssuser",
					"email":    "xss@example.com",
					"password": "password123",
					"name":     "<script>alert('XSS')</script>",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := testutil.GraphQLRequest{
					Query: `mutation { 
						register(input: $input) { 
							token 
							user { 
								id 
								username 
								name
							} 
						} 
					}`,
					Variables: map[string]interface{}{
						"input": tc.input,
					},
				}

				resp := testutil.ExecuteGraphQLRequest(t, srv, req)

				// XSS試行以外はエラーが発生することを期待
				// XSSの場合はサニタイズされて成功する可能性もある
				if tc.name == "XSS試行が無効化される" {
					if resp.Errors == nil {
						// 成功した場合、名前がサニタイズされていることを確認
						data := testutil.GetDataMap(t, resp)
						register := testutil.GetMap(t, data, "register")
						user := testutil.GetMap(t, register, "user")
						name := user["name"].(string)

						// HTMLタグが除去またはエスケープされていることを確認
						if strings.Contains(name, "<script>") {
							t.Error("XSS attack not properly sanitized")
						}
					}
				} else {
					// その他のケースはエラーが発生することを期待
					testutil.AssertHasErrors(t, resp)
				}
			})
		}
	})

	t.Run("境界値テスト", func(t *testing.T) {
		testCases := []struct {
			name    string
			input   map[string]interface{}
			wantErr bool
		}{
			{
				name: "ユーザー名が最小長度ちょうどの場合成功する",
				input: map[string]interface{}{
					"username": "ab", // 2文字（最小長度と仮定）
					"email":    "min@example.com",
					"password": "password123",
					"name":     "Min User",
				},
				wantErr: false,
			},
			{
				name: "ユーザー名が1文字の場合エラーが発生する",
				input: map[string]interface{}{
					"username": "a", // 1文字
					"email":    "single@example.com",
					"password": "password123",
					"name":     "Single User",
				},
				wantErr: true,
			},
			{
				name: "Unicode文字を含むユーザー名が正しく処理される",
				input: map[string]interface{}{
					"username": "ユーザー123",
					"email":    "unicode@example.com",
					"password": "password123",
					"name":     "Unicode User 😀",
				},
				wantErr: false, // Unicode対応していると仮定
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := testutil.GraphQLRequest{
					Query: `mutation { 
						register(input: $input) { 
							token 
							user { 
								id 
								username 
								name
							} 
						} 
					}`,
					Variables: map[string]interface{}{
						"input": tc.input,
					},
				}

				resp := testutil.ExecuteGraphQLRequest(t, srv, req)

				if tc.wantErr {
					testutil.AssertHasErrors(t, resp)
				} else {
					testutil.AssertNoErrors(t, resp)
					// 成功した場合、データが正しく保存されていることを確認
					data := testutil.GetDataMap(t, resp)
					register := testutil.GetMap(t, data, "register")
					user := testutil.GetMap(t, register, "user")

					if user["username"] != tc.input["username"] {
						t.Errorf("Expected username %s, got %s", tc.input["username"], user["username"])
					}
				}
			})
		}
	})
}
