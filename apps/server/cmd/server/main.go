package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"sns-server/internal/config"
	"sns-server/internal/models"
	"sns-server/internal/server"
)

func main() {
	// 設定読み込み
	cfg := config.Load()

	log.Printf("Starting server in %s mode on port %s", cfg.Env, cfg.Port)

	// データベース接続
	db := connectDB(cfg)

	// マイグレーション実行
	err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Like{},
		&models.Follow{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// サーバー作成
	srv := &server.Server{
		DB:     db,
		Config: cfg,
	}

	// ルーター設定
	router := chi.NewRouter()

	// ミドルウェア
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware(cfg))

	// GraphQLエンドポイント
	router.Post("/query", srv.HandleGraphQL)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>SNS GraphQL API</title>
</head>
<body>
    <h1>SNS GraphQL API Server</h1>
    <p>GraphQLエンドポイント: <code>POST /query</code></p>
    <h2>サンプルクエリ:</h2>
    <pre>
# ユーザー一覧
{
  users {
    id
    username
    name
  }
}

# 投稿一覧
{
  posts {
    id
    content
    author {
      username
    }
  }
}
    </pre>
    <h2>サンプルミューテーション:</h2>
    <pre>
# ユーザー登録
mutation {
  register(input: {
    username: "test_user"
    email: "test@example.com"
    password: "password123"
    name: "テストユーザー"
  }) {
    token
    user {
      id
      username
    }
  }
}

# 投稿作成
mutation {
  createPost(input: {
    content: "Hello, SNS!"
  }) {
    id
    content
    author {
      username
    }
  }
}
    </pre>
</body>
</html>
		`))
	})

	log.Printf("GraphQL server ready at http://localhost:%s/", cfg.Port)
	log.Printf("GraphQL endpoint: http://localhost:%s/query", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func connectDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func corsMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// CORS Origins設定をチェック
			if len(cfg.CORSOrigins) == 1 && cfg.CORSOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, allowedOrigin := range cfg.CORSOrigins {
					if origin == allowedOrigin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
