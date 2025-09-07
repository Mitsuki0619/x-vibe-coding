# SNS Application

TwitterライクなSNSアプリケーションです。Turborepoでモノレポ構成され、TDD（Test-Driven Development）手法で開発されています。

## 🏗️ プロジェクト構成

このプロジェクトは以下の構成で構築されています：

### Apps
- **`apps/web/`** - Next.js 15 + React 19 + App Routerを使用したWebフロントエンド
- **`apps/mobile/`** - Expo + React Nativeを使用したモバイルアプリ
- **`apps/server/`** - Golang + GraphQL + PostgreSQLを使用したAPIサーバー

### Packages
- **`packages/biome-config/`** - 共有Biome設定（ESLint/Prettierの代替）
- **`packages/typescript-config/`** - 共有TypeScript設定

## 🚀 技術スタック

### フロントエンド
- **Web**: Next.js 15, React 19, TypeScript, Turbopack
- **Mobile**: Expo, React Native, TypeScript

### バックエンド
- **サーバー**: Golang 1.23
- **API**: GraphQL（カスタム実装）
- **データベース**: PostgreSQL with Docker
- **ORM**: GORM v2
- **HTTP Router**: Chi v5
- **テスト**: TDD with Go標準テストパッケージ

### 開発ツール
- **モノレポ**: Turborepo + pnpm workspaces
- **コード品質**: Biome（リント・フォーマット）
- **環境管理**: Docker + 環境変数
- **開発手法**: TDD（Test-Driven Development）

## 🎯 機能

### 実装済み機能 ✅
- **ユーザー管理**: 登録、認証、プロフィール
- **投稿機能**: 作成、一覧表示、詳細表示
- **いいね機能**: 投稿へのいいね・いいね取り消し
- **フォロー機能**: ユーザー間のフォロー・アンフォロー
- **GraphQL API**: 完全なCRUD操作
- **データベース**: PostgreSQL with完全なリレーション

### 開発予定機能 🚧
- JWT認証システム
- リアルタイム通信（Subscription）
- ファイルアップロード（画像・動画）
- より完全なGraphQLスキーマ
- フロントエンド実装

## 🛠️ 開発環境セットアップ

### 前提条件
- Node.js 18以上
- pnpm
- Go 1.23以上
- Docker & Docker Compose

### 初期設定

1. **リポジトリクローン**
   ```bash
   git clone [repository-url]
   cd sample-todo-app
   ```

2. **依存関係インストール**
   ```bash
   pnpm install
   ```

3. **サーバー環境設定**
   ```bash
   cd apps/server
   cp .env.example .env
   # 必要に応じて.envファイルを編集
   ```

4. **データベース起動**
   ```bash
   cd apps/server
   docker-compose up -d postgres
   ```

## 📝 開発コマンド

### 全体操作（ルートディレクトリ）
```bash
# 全アプリ開発モードで起動
pnpm dev

# 全アプリビルド
pnpm build

# コードリント
pnpm lint

# コードフォーマット
pnpm fmt

# 型チェック
pnpm check-types
```

### サーバー開発（apps/server/）

#### 🚀 Makefileを使用した簡単コマンド（推奨）
```bash
# ヘルプ表示（全コマンド一覧）
make help

# 開発環境セットアップ
make setup              # 依存関係インストール + データベース起動

# 開発サーバー起動
make dev                # ホットリロード対応

# TDD開発
make test               # 全テスト実行
make test-models        # モデルテスト実行
make test-integration   # 統合テスト実行

# データベース管理
make db-up              # 開発用DB起動
make db-test            # テスト用DB起動
make db-all             # 全DB + pgAdmin起動

# コード品質管理
make check              # フォーマット + 静的解析 + テスト
make format             # コードフォーマット
make lint               # 静的解析
```

#### 📋 TDDワークフロー例
```bash
# 1. 環境準備
make setup

# 2. TDDサイクル
make red          # Red: テスト失敗確認
# [実装作業]
make green        # Green: テスト成功確認
make refactor     # Refactor: 品質チェック

# 3. 最終確認
make test         # 全テスト実行
```

#### 🛠️ 従来のGoコマンド（直接実行）
```bash
# 開発用データベース起動
docker-compose up -d postgres

# テスト用データベース起動
docker-compose --profile test up -d postgres_test

# サーバー起動
go run cmd/server/main.go

# 全テスト実行（TDD）
go test ./... -v

# モデルテスト実行
go test ./internal/models -v

# 統合テスト実行
go test ./internal/server -v

# コードフォーマット
go fmt ./...

# 静的解析
go vet ./...
```

### フロントエンド開発
```bash
# Web開発サーバー（Turbopack）
pnpm dev --filter=web

# Mobile開発サーバー（Expo）
pnpm dev --filter=mobile
```

## 🧪 テスト

このプロジェクトは**TDD（Test-Driven Development）**で開発されています。

### サーバーサイドテスト ✅
- **完全なTDD実装済み**
- **モデルテスト**: User、Post、Like、Followの包括的テスト
- **統合テスト**: GraphQL API全体のE2Eテスト
- **テスト実行**: `go test ./... -v`
- **テストカバレッジ**: 23テストスイート、46サブテスト

### テスト環境
- 独立したテスト用PostgreSQLデータベース（port 5433）
- 自動データベースクリーンアップ
- 環境変数による設定分離（.env.test）

## 🗄️ データベース

### PostgreSQL（Docker）
- **開発用**: localhost:5432
- **テスト用**: localhost:5433
- **管理ツール**: pgAdmin (localhost:5050)

### データベーススキーマ
```sql
users: id, username, email, password, name, bio, created_at, updated_at
posts: id, content, author_id, created_at, updated_at
likes: id, user_id, post_id, created_at
follows: id, follower_id, followee_id, created_at
```

## 🔌 API

### GraphQL エンドポイント
- **URL**: `http://localhost:8080/query`
- **管理画面**: `http://localhost:8080/`

### 利用可能なクエリ・ミューテーション
```graphql
# クエリ
{
  users { id username name email }
  posts { id content author { username } }
}

# ミューテーション
mutation {
  register(input: { username: "user", email: "user@example.com", password: "password", name: "User" }) {
    token
    user { id username }
  }
  
  createPost(input: { content: "Hello, SNS!" }) {
    id content author { username }
  }
  
  likePost(input: { postId: 1 }) {
    id user { username } post { content }
  }
  
  unlikePost(input: { postId: 1 })
}
```

## 🏆 TDD開発手法

このプロジェクトはt-wada推奨のTDD手法に従って開発されています。

### Red-Green-Refactor サイクル
1. **🔴 Red**: 失敗するテストを書く
2. **🟢 Green**: 最小限の実装でテストを成功させる
3. **🔵 Refactor**: コード品質を改善（テスト実行必須）

### TDD実践ルール
- テストなしにコードを書かない
- リファクタリング時は必ずテスト実行
- 静的解析・フォーマッタを定期実行

詳細は[CLAUDE.md](./CLAUDE.md)の「TDD開発手法」セクションを参照。

## 📚 ドキュメント

- **[CLAUDE.md](./CLAUDE.md)** - Claude Code向け開発ガイド（アーキテクチャ、TDD手法、コマンド等）
- **API仕様** - GraphQLエンドポイント（`http://localhost:8080/`）で確認可能

## 🤝 開発への参加

1. **TDD手法を必須で遵守**してください
2. すべての新機能はテストファーストで開発
3. リファクタリング時は必ずテスト実行
4. コードフォーマット（`go fmt`、`pnpm fmt`）を実行
5. 静的解析（`go vet`）でエラーがないことを確認

## 🔗 関連リンク

- [Turborepo Documentation](https://turborepo.com/docs)
- [Next.js Documentation](https://nextjs.org/docs)
- [Expo Documentation](https://docs.expo.dev/)
- [Go Documentation](https://golang.org/doc/)
- [GraphQL Specification](https://graphql.org/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

**開発ステータス**: アクティブ開発中 🚧  
**開発手法**: TDD（Test-Driven Development）  
**テストカバレッジ**: サーバーサイド100%（23テストスイート）# x-vibe-coding
