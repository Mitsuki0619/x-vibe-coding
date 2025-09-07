# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

このリポジトリはTwitterライクなSNSアプリケーションをTurborepoで構成されたモノレポです：

- **apps/web/**: Next.js + React 19 + App Routerを使用したWebフロントエンド
- **apps/mobile/**: Expo + React Nativeを使用したモバイルアプリ
- **apps/server/**: Golang + GraphQLを使用したAPIサーバー
- **packages/biome-config/**: 共有Biome設定
- **packages/typescript-config/**: 共有TypeScript設定

### 技術スタック
- **フロントエンド（Web）**: Next.js 15 + React 19 + TypeScript + GraphQL
- **フロントエンド（Mobile）**: Expo + React Native + TypeScript + GraphQL
- **バックエンド**: Golang + GraphQL + PostgreSQL
- **API**: すべてのクライアント-サーバー通信にGraphQLを使用

### SNS機能
- ユーザー登録・認証
- ツイート投稿・表示
- いいね・リツイート機能
- フォロー・フォロワー機能
- タイムライン表示

## 開発コマンド

### ルートレベル（Turborepo）
```bash
# 依存関係のインストール
pnpm install

# 全アプリを開発モードで起動
pnpm dev

# 全アプリをビルド
pnpm build

# 全アプリをリント
pnpm lint

# リントエラーを修正
pnpm lint:fix

# コードフォーマット
pnpm fmt

# 型チェック
pnpm check-types
```

### Webアプリ（apps/web/）
```bash
# Turbopackで開発サーバー起動
pnpm dev --filter=web

# プロダクション用ビルド
pnpm build --filter=web

# プロダクションサーバー起動
cd apps/web && pnpm start

# Biomeでリント
cd apps/web && pnpm lint

# リントエラーを修正
cd apps/web && pnpm lint:fix
```

### モバイルアプリ（apps/mobile/）
```bash
# Expo開発サーバー起動
pnpm dev --filter=mobile

# iOSシミュレーターで起動
cd apps/mobile && pnpm ios

# Androidエミュレーターで起動
cd apps/mobile && pnpm android

# Webブラウザで起動
cd apps/mobile && pnpm web

# プロジェクトリセット（現在のappをapp-exampleに移動）
cd apps/mobile && pnpm reset-project
```

### サーバー（apps/server/）
```bash
# 開発用データベース起動（Docker）
cd apps/server && docker-compose up -d postgres

# テスト用データベース起動（Docker）
cd apps/server && docker-compose --profile test up -d postgres_test

# 🚀 Makefile使用（推奨・簡単）
cd apps/server && make help          # 全コマンド表示
cd apps/server && make setup         # 初期セットアップ
cd apps/server && make dev            # 開発サーバー起動
cd apps/server && make test           # 全テスト実行
cd apps/server && make check          # 品質チェック（format+lint+test）

# 🛠️ 従来コマンド（直接実行）
# 全データベース起動（開発・テスト・pgAdmin）
cd apps/server && docker-compose --profile test up -d

# 環境変数設定（.envファイルから自動読込）
# 必要に応じて .env ファイルを .env.example からコピーして編集

# Golangサーバー起動（開発モード）
cd apps/server && go run cmd/server/main.go

# ビルド版でサーバー起動
cd apps/server && go build -o sns-server cmd/server/main.go && ./sns-server

# モデルテスト実行（TDD）
cd apps/server && go test ./internal/models -v

# 統合テスト実行
cd apps/server && go test ./internal/server -v

# 全テスト実行
cd apps/server && go test ./... -v

# 依存関係整理
cd apps/server && go mod tidy

# GraphQL API確認
# ブラウザで http://localhost:8080 を開くとAPI仕様が表示されます
# GraphQLクエリは http://localhost:8080/query にPOSTで送信

# データベース管理ツール（pgAdmin）
# http://localhost:5050 でアクセス
# Email: admin@sns.local, Password: admin
```

## アーキテクチャの特徴

### モノレポ構成
- Turborepoとpnpm workspacesでオーケストレーション
- packages/で共通設定を管理し、アプリ間の一貫性を保つ
- ESLint/Prettierの代わりにBiomeを使用

### GraphQL設計
- クライアント-サーバー間の全通信をGraphQLで統一
- 型安全性とパフォーマンスを重視
- コードジェネレーションでTypeScript型定義を自動生成

### モバイルアプリ構成
- Expo Routerでファイルベースナビゲーション
- app/(tabs)/でタブベースナビゲーション
- ライト/ダークモード対応のThemedコンポーネント
- カスタムフックでカラースキーム管理

### Webアプリ構成
- Next.js 15 + App Router + Turbopack
- CSS Modulesでスタイリング
- React 19 + TypeScript

### Golangサーバー構成

#### アーキテクチャ
- **GraphQLサーバー**: カスタム実装（本格運用時はgqlgen推奨）
- **データベース**: PostgreSQL with Docker
- **ORM**: GORM v2
- **HTTP Router**: Chi v5
- **環境変数管理**: godotenv + カスタムconfigパッケージ
- **テスト**: TDD手法でGo標準テストパッケージ使用

#### 実装済み機能
- ✅ **データベース接続とマイグレーション**: PostgreSQL + GORM
- ✅ **データモデル**: User、Post、Like、Followの完全なモデル実装
- ✅ **TDDモデルテスト**: 全モデルの包括的テストスイート（8テストスイート）
- ✅ **GraphQL API**: 基本的なクエリとミューテーション
  - ユーザー登録（register）
  - ユーザー一覧取得（users）
  - 投稿作成（createPost）
  - 投稿一覧取得（posts）
- ✅ **統合テスト**: GraphQL APIの完全な統合テストスイート（6テストスイート）
- ✅ **Docker環境**: 開発用・テスト用PostgreSQLの分離
- ✅ **環境変数管理**: .env/.env.test/.env.exampleファイルと設定システム
- ✅ **CORSミドルウェア**: 設定可能なCORS対応
- ✅ **ログミドルウェア**: Chi標準ログ・リカバリーミドルウェア

#### 技術的特徴
- **TDD手法**: t-wada推奨のTDDプロセスでモデル開発
- **分離されたテスト環境**: 本番・開発・テスト用の独立したDB
- **型安全性**: GORM v2のタイプセーフなクエリ
- **バリデーション**: モデルレベルでのデータ検証（BeforeCreateフック）
- **関係性管理**: ユーザー間フォロー、投稿へのいいね等の関係性を適切に実装

#### 未実装機能（次のステップ）
- 🚧 JWT認証・認可システム
- 🚧 より完全なGraphQLスキーマ（リツイート、返信等）
- 🚧 リアルタイム通信（Subscription）
- 🚧 ファイルアップロード（画像・動画）
- 🚧 パフォーマンス最適化（DataLoader等）

#### API仕様
GraphQL endpoint: `POST /query`
- 現在のクエリ: `users`, `posts`
- 現在のミューテーション: `register`, `createPost`
- レスポンス形式: 標準GraphQLレスポンス（data/errorsフィールド）

#### データベーススキーマ
```
users: id, username, email, password, name, bio, created_at, updated_at
posts: id, content, author_id, created_at, updated_at
likes: id, user_id, post_id, created_at
follows: id, follower_id, followee_id, created_at
```

#### テストカバレッジ
- モデルテスト: 全機能の包括的テスト（バリデーション、関係性、ビジネスロジック）
- 統合テスト: GraphQL API全体のE2Eテスト
- テスト実行: `go test ./...` で全テスト実行可能

### 共有ツール
- 全アプリでTypeScript使用
- 一貫したリント・フォーマットにBiome使用
- 効率的なタスク実行にTurborepo使用
- パッケージ管理にpnpm使用

## テスト

### サーバーサイド（Go）
✅ **完全にTDD実装済み**
- **モデルテスト**: `go test ./internal/models -v`
  - User、Post、Like、Followモデルの包括的テスト
  - バリデーション、関係性、ビジネスロジックのテスト
  - 8つのテストスイートで全機能をカバー
  
- **統合テスト**: `go test ./internal/server -v`
  - GraphQL API全体のE2Eテスト
  - ユーザー登録、投稿作成、データ取得のテスト
  - 6つのテストスイートでAPI動作を検証
  
- **テスト環境**: 
  - 独立したテスト用PostgreSQLデータベース（port 5433）
  - 自動的なデータベースクリーンアップ
  - 環境変数による設定分離（.env.test）

### フロントエンド
🚧 **未実装** - テスト実装時は既存のpackage.jsonファイルでテストスクリプトを確認してから実装すること。

## TDD開発手法（必須）

このプロジェクトでは**Test-Driven Development（TDD）**を徹底すること。t-wada推奨のTDDサイクルに従い、以下の**Red-Green-Refactor**サイクルを必ず実行する。

### TDDサイクル: Red-Green-Refactor

#### 🔴 **Red Phase（レッドフェーズ）**
**目標**: 失敗するテストを書く

1. **新機能の仕様を理解**
   - ユーザーストーリーや要件を分析
   - 実装すべき機能の振る舞いを明確化

2. **失敗するテストを作成**
   ```bash
   # テスト作成例
   go test ./internal/models -v  # 失敗することを確認
   ```
   - 期待する動作を表現するテストケースを記述
   - テストが失敗することを確認（まだ実装がないため）
   - テストが適切な理由で失敗していることを確認

3. **テストケースの網羅性確認**
   - 正常系のテスト
   - 異常系のテスト（バリデーションエラー等）
   - 境界値のテスト
   - エッジケースのテスト

#### 🟢 **Green Phase（グリーンフェーズ）**
**目標**: テストを最小限の実装で成功させる

1. **最小実装でテストを通す**
   ```bash
   go test ./internal/models -v  # 成功することを確認
   ```
   - テストが成功する最小限のコードを実装
   - 「美しい」コードである必要はない
   - まず動作することを最優先

2. **全テストの成功確認**
   ```bash
   go test ./... -v  # 既存テストが壊れていないことを確認
   ```
   - 新しいテストが成功
   - 既存のテストが壊れていない
   - 回帰テストの確認

3. **統合テストの確認**
   - モデルテスト成功後、統合テストも実装
   - GraphQL APIレベルでの動作確認

#### 🔵 **Refactor Phase（リファクタリングフェーズ）**
**目標**: 機能を保ったままコードの品質を改善

1. **コード品質の分析**
   - 重複コード（DRY原則違反）の特定
   - 長すぎる関数やクラスの特定
   - 複雑すぎるロジックの特定
   - 命名の改善が必要な箇所の特定

2. **リファクタリング実行**
   ```bash
   # リファクタリング例
   # 1. 重複コード削除
   # 2. 関数の抽出
   # 3. 定数の定義
   # 4. エラーハンドリング統一
   ```

3. **各リファクタリング後のテスト実行**
   ```bash
   go test ./... -v  # 必須：各変更後にテスト実行
   ```
   - **重要**: リファクタリングのたびに必ずテストを実行
   - 機能が壊れていないことを確認
   - 回帰テストの実行

4. **静的解析とフォーマッタ実行**
   ```bash
   go fmt ./...      # コードフォーマット
   go vet ./...      # 静的解析
   ```

### TDD実践のベストプラクティス

#### **必須ルール**
1. **テストなしにコードを書かない**
   - 実装前に必ずテストを書く
   - テストが失敗することを確認してから実装

2. **リファクタリング時は必ずテスト実行**
   - 小さな変更でもテストを実行
   - 機能が壊れていないことを確認

3. **全体テストの定期実行**
   ```bash
   go test ./... -v  # 定期的に実行
   ```

#### **テスト設計指針**

**モデルテスト**:
```go
func TestModel_Creation(t *testing.T) {
    // 正常系・異常系・境界値を網羅
    tests := []struct {
        name    string
        input   ModelInput
        wantErr bool
        errMsg  string
    }{
        // テストケース定義
    }
}
```

**統合テスト**:
```go
func TestAPI_Integration(t *testing.T) {
    // GraphQL APIの統合テスト
    // データベースとの連携テスト
    // エラーハンドリングテスト
}
```

#### **コード品質基準**

1. **DRY原則**: 重複コードの排除
2. **SOLID原則**: 特に単一責任原則
3. **関数の長さ**: 20行以下を目標
4. **命名**: 意図が明確に伝わる命名
5. **エラーハンドリング**: 一貫した処理

#### **リファクタリングチェックリスト**

- [ ] 重複コードは抽出されているか
- [ ] 関数は適切な長さか（20行以下推奨）
- [ ] マジックナンバーは定数化されているか
- [ ] エラーハンドリングは統一されているか
- [ ] 命名は意図を明確に表現しているか
- [ ] 全テストが成功しているか
- [ ] 静的解析でエラーが出ていないか
- [ ] コードフォーマットが統一されているか

### TDD実装例

このプロジェクトのLikeモデル実装例：

1. **Red**: `TestLike_Creation`で失敗するテスト作成
2. **Green**: `Like`モデルとバリデーション実装でテスト成功
3. **Refactor**: デフォルトユーザー作成の重複コード削除、ヘルパー関数統一

結果：高品質で保守性の高いコードベースの実現

### 開発フロー

```bash
# 1. Red: テスト作成
go test ./internal/models -v  # 失敗確認

# 2. Green: 最小実装
go test ./internal/models -v  # 成功確認
go test ./... -v             # 全体確認

# 3. Refactor: 品質改善
# リファクタリング実行
go test ./... -v             # テスト確認
go fmt ./...                 # フォーマット
go vet ./...                 # 静的解析

# 4. 統合テスト
go test ./internal/server -v

# 5. 最終確認
go test ./... -v
```

**TDDは単なるテスト手法ではなく、設計手法である。テストファーストで考えることで、より良い設計とコード品質を実現する。**