# SNS Server Makefile
# Goサーバーの開発・テスト・デプロイを簡単にするためのMakefile

.PHONY: help dev build test test-models test-integration test-coverage clean db-up db-down db-reset lint format vet deps check-deps server-start server-stop

# デフォルトターゲット
.DEFAULT_GOAL := help

# 色付きヘルプメッセージ
BOLD := \033[1m
RESET := \033[0m
GREEN := \033[32m
BLUE := \033[34m
YELLOW := \033[33m

## ヘルプ表示
help:
	@echo "$(BOLD)SNS Server Development Commands$(RESET)"
	@echo ""
	@echo "$(GREEN)📦 環境・依存関係$(RESET)"
	@echo "  $(BLUE)deps$(RESET)          - Go依存関係をインストール・更新"
	@echo "  $(BLUE)check-deps$(RESET)    - 依存関係の整合性確認"
	@echo ""
	@echo "$(GREEN)🗄️  データベース$(RESET)"
	@echo "  $(BLUE)db-up$(RESET)         - 開発用データベース起動"
	@echo "  $(BLUE)db-test$(RESET)       - テスト用データベース起動"
	@echo "  $(BLUE)db-all$(RESET)        - 全データベース起動（開発・テスト・pgAdmin）"
	@echo "  $(BLUE)db-down$(RESET)       - 全データベース停止"
	@echo "  $(BLUE)db-reset$(RESET)      - データベースリセット（データ削除）"
	@echo "  $(BLUE)db-logs$(RESET)       - データベースログ表示"
	@echo ""
	@echo "$(GREEN)🚀 開発・実行$(RESET)"
	@echo "  $(BLUE)dev$(RESET)           - 開発サーバー起動（ホットリロード）"
	@echo "  $(BLUE)build$(RESET)         - プロダクションビルド"
	@echo "  $(BLUE)run$(RESET)           - ビルド済みサーバー実行"
	@echo ""
	@echo "$(GREEN)🧪 テスト（TDD）$(RESET)"
	@echo "  $(BLUE)test$(RESET)          - 全テスト実行"
	@echo "  $(BLUE)test-models$(RESET)   - モデルテストのみ実行"
	@echo "  $(BLUE)test-integration$(RESET) - 統合テストのみ実行"
	@echo "  $(BLUE)test-watch$(RESET)    - テスト監視モード（ファイル変更時に自動実行）"
	@echo "  $(BLUE)test-coverage$(RESET) - カバレッジ付きテスト実行"
	@echo ""
	@echo "$(GREEN)🔍 コード品質$(RESET)"
	@echo "  $(BLUE)lint$(RESET)          - 静的解析（go vet）"
	@echo "  $(BLUE)format$(RESET)        - コードフォーマット（go fmt）"
	@echo "  $(BLUE)check$(RESET)         - 全品質チェック（format + lint + test）"
	@echo ""
	@echo "$(GREEN)🛠️  ユーティリティ$(RESET)"
	@echo "  $(BLUE)clean$(RESET)         - ビルドファイル削除"
	@echo "  $(BLUE)logs$(RESET)          - サーバーログ表示"
	@echo "  $(BLUE)ps$(RESET)            - 実行中のプロセス確認"
	@echo ""
	@echo "$(YELLOW)📖 TDDワークフロー例:$(RESET)"
	@echo "  1. make db-up           # データベース起動"
	@echo "  2. make test-models     # Red: テスト失敗確認"
	@echo "  3. [実装]               # Green: 実装してテスト成功"
	@echo "  4. make check           # Refactor: 品質チェック"
	@echo "  5. make test            # 全テスト確認"

## 環境・依存関係
deps:
	@echo "$(GREEN)📦 Go依存関係をインストール・更新中...$(RESET)"
	go mod tidy
	go mod download
	@echo "$(GREEN)✅ 依存関係の更新完了$(RESET)"

check-deps:
	@echo "$(GREEN)📦 依存関係の整合性確認中...$(RESET)"
	go mod verify
	@echo "$(GREEN)✅ 依存関係の確認完了$(RESET)"

## データベース管理
db-up:
	@echo "$(GREEN)🗄️  開発用データベース起動中...$(RESET)"
	docker-compose up -d postgres
	@echo "$(GREEN)✅ 開発用データベース起動完了 (localhost:5432)$(RESET)"

db-test:
	@echo "$(GREEN)🗄️  テスト用データベース起動中...$(RESET)"
	docker-compose --profile test up -d postgres_test
	@echo "$(GREEN)✅ テスト用データベース起動完了 (localhost:5433)$(RESET)"

db-all:
	@echo "$(GREEN)🗄️  全データベース（開発・テスト・pgAdmin）起動中...$(RESET)"
	docker-compose --profile test up -d
	@echo "$(GREEN)✅ 全データベース起動完了$(RESET)"
	@echo "  - 開発用DB: localhost:5432"
	@echo "  - テスト用DB: localhost:5433"
	@echo "  - pgAdmin: http://localhost:5050"

db-down:
	@echo "$(GREEN)🗄️  全データベース停止中...$(RESET)"
	docker-compose --profile test down
	@echo "$(GREEN)✅ 全データベース停止完了$(RESET)"

db-reset:
	@echo "$(YELLOW)⚠️  データベースリセット（全データ削除）中...$(RESET)"
	docker-compose --profile test down -v
	docker-compose --profile test up -d
	@echo "$(GREEN)✅ データベースリセット完了$(RESET)"

db-logs:
	@echo "$(GREEN)📋 データベースログ表示中...$(RESET)"
	docker-compose logs -f postgres postgres_test

## 開発・実行
dev:
	@echo "$(GREEN)🚀 開発サーバー起動中...$(RESET)"
	@echo "$(YELLOW)💡 GraphQL API: http://localhost:8080$(RESET)"
	@echo "$(YELLOW)💡 終了: Ctrl+C$(RESET)"
	go run cmd/server/main.go

build:
	@echo "$(GREEN)🔨 プロダクションビルド中...$(RESET)"
	go build -o bin/sns-server cmd/server/main.go
	@echo "$(GREEN)✅ ビルド完了: bin/sns-server$(RESET)"

run: build
	@echo "$(GREEN)🚀 ビルド済みサーバー実行中...$(RESET)"
	@echo "$(YELLOW)💡 GraphQL API: http://localhost:8080$(RESET)"
	./bin/sns-server

## テスト（TDD）
test:
	@echo "$(GREEN)🧪 全テスト実行中（TDD）...$(RESET)"
	go test ./... -v
	@echo "$(GREEN)✅ 全テスト完了$(RESET)"

test-models:
	@echo "$(GREEN)🧪 モデルテスト実行中...$(RESET)"
	go test ./internal/models -v
	@echo "$(GREEN)✅ モデルテスト完了$(RESET)"

test-integration:
	@echo "$(GREEN)🧪 統合テスト実行中...$(RESET)"
	go test ./internal/server -v
	@echo "$(GREEN)✅ 統合テスト完了$(RESET)"

test-watch:
	@echo "$(GREEN)👀 テスト監視モード開始...$(RESET)"
	@echo "$(YELLOW)💡 ファイル変更時に自動テスト実行$(RESET)"
	@echo "$(YELLOW)💡 終了: Ctrl+C$(RESET)"
	# Go用のファイル監視ツールが必要な場合は追加でインストール
	# 例: go install github.com/cosmtrek/air@latest
	# air -c .air.toml
	while true; do \
		go test ./... -v; \
		echo "$(YELLOW)📝 ファイル変更を監視中... (手動再実行: Enterキー)$(RESET)"; \
		read; \
	done

test-coverage:
	@echo "$(GREEN)📊 カバレッジ付きテスト実行中...$(RESET)"
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ テストカバレッジ完了$(RESET)"
	@echo "$(YELLOW)📊 カバレッジレポート: coverage.html$(RESET)"

## コード品質
format:
	@echo "$(GREEN)🎨 コードフォーマット実行中...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✅ フォーマット完了$(RESET)"

lint: vet

vet:
	@echo "$(GREEN)🔍 静的解析実行中...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✅ 静的解析完了$(RESET)"

check: format vet test
	@echo "$(GREEN)✨ 全品質チェック完了$(RESET)"
	@echo "  ✅ コードフォーマット"
	@echo "  ✅ 静的解析"
	@echo "  ✅ 全テスト"

## ユーティリティ
clean:
	@echo "$(GREEN)🧹 ビルドファイル削除中...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache
	@echo "$(GREEN)✅ クリーンアップ完了$(RESET)"

logs:
	@echo "$(GREEN)📋 サーバーログ表示中...$(RESET)"
	# サーバーがバックグラウンドで実行されている場合のログ表示
	# 実装方法はサーバーの起動方法により異なる
	@echo "$(YELLOW)💡 開発サーバーは 'make dev' で起動してください$(RESET)"

ps:
	@echo "$(GREEN)📋 実行中のプロセス確認中...$(RESET)"
	@echo "$(YELLOW)Go関連プロセス:$(RESET)"
	@-ps aux | grep -E "(go|sns-server)" | grep -v grep || echo "Go関連プロセスが見つかりません"
	@echo ""
	@echo "$(YELLOW)Docker関連プロセス:$(RESET)"
	@-docker-compose ps || echo "Docker Composeプロセスが見つかりません"

## 開発ワークフロー用ショートカット
setup: deps db-up
	@echo "$(GREEN)🎉 開発環境セットアップ完了$(RESET)"
	@echo "$(YELLOW)次のステップ:$(RESET)"
	@echo "  1. make dev    # サーバー起動"
	@echo "  2. make test   # テスト実行"

teardown: db-down clean
	@echo "$(GREEN)🧹 環境クリーンアップ完了$(RESET)"

# TDD用クイックコマンド
red: test-models
green: test-models
refactor: check

tdd-cycle: red
	@echo "$(YELLOW)🔴 RED: テストが失敗することを確認しました$(RESET)"
	@echo "$(YELLOW)次: 実装してからmake greenを実行してください$(RESET)"

## 本番環境用（追加）
prod-build:
	@echo "$(GREEN)🏭 本番用ビルド中...$(RESET)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/sns-server cmd/server/main.go
	@echo "$(GREEN)✅ 本番用ビルド完了$(RESET)"