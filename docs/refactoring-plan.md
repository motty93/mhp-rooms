# リファクタリング計画

## 概要

現在のアプリケーションは機能開発とリリースを優先しており、一部の設計改善は次フェーズで実装予定です。

## 現在の設計状況

### アーキテクチャ
```
Handler → Repository → Database
```

### Service層の現状
- `internal/services`ディレクトリは存在
- `ActivityService`のみ実装済み（ユーザーアクティビティ記録）
- 他のハンドラーは直接Repository層を呼び出し

### 問題点
1. Service層の一貫性不足
2. ビジネスロジックがHandlerに散在
3. 再利用性の低下

## リファクタリング計画

### Phase 1: リリース優先（現在）
**期間**: リリースまで
**方針**: 現在の設計を維持し、機能開発に集中

#### 対応内容
- [ ] 現在のパターン（Handler → Repository）で新機能を実装
- [ ] ビジネスロジックの複雑化を避ける
- [ ] 技術的負債として記録
- [ ] リリース完了

#### 新機能実装時の注意点
```go
// 現在のパターンを維持
func (h *Handler) SomeAction(w http.ResponseWriter, r *http.Request) {
    // 簡単なバリデーション
    if err := validateInput(input); err != nil {
        // エラー処理
    }
    
    // Repository呼び出し
    err := h.repo.SomeEntity.Create(data)
    if err != nil {
        // エラー処理
    }
}
```

### Phase 2: Service層導入（リリース後）
**期間**: リリース後3-6ヶ月
**方針**: 段階的なService層導入

#### Step 1: Service層設計
- [ ] 各ドメインのService定義
- [ ] インターフェース設計
- [ ] 依存関係の整理

#### Step 2: 優先度の高いServiceから実装
1. **UserService** (最優先)
   - ユーザー管理
   - プロフィール更新
   - 認証関連ロジック

2. **ReportService**
   - 通報処理
   - バリデーション
   - 通知ロジック

3. **RoomService**
   - ルーム管理
   - 参加制御
   - メッセージ管理

#### Step 3: 段階的移行
```go
// 移行後の理想的なアーキテクチャ
Handler → Service → Repository → Database

// 実装例
type UserService interface {
    CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error)
    UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error
    GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
}
```

#### Step 4: 既存コードの移行
- [ ] Handler層からビジネスロジックを抽出
- [ ] Service層への移行
- [ ] テストの更新
- [ ] パフォーマンステスト

### Phase 3: 最適化とクリーンアップ
**期間**: Service層導入完了後
**方針**: パフォーマンス最適化と設計改善

#### 対応内容
- [ ] 不要なコードの削除
- [ ] パフォーマンス最適化
- [ ] ドキュメント更新
- [ ] テストカバレッジ向上

## 判断基準

### Service層導入の判断基準
以下の条件が揃った時点でPhase 2に移行：
1. ✅ 初回リリース完了
2. ✅ ユーザーフィードバック収集完了
3. ✅ 重大なバグ修正完了
4. ✅ 開発リソースに余裕がある

### 移行時の注意点
1. **後方互換性**: 既存APIの互換性を維持
2. **段階的移行**: 一度に全てを変更しない
3. **テスト**: 各段階で十分なテストを実施
4. **ドキュメント**: 変更内容を適切に文書化

## リスク管理

### 主要リスク
1. **既存機能の破綻**: 大幅な変更によるバグ混入
2. **開発期間の延長**: リファクタリングによる機能開発の遅延
3. **学習コスト**: 新しいアーキテクチャの習得

### 対策
1. **段階的実装**: 小さな単位での変更
2. **十分なテスト**: 自動テストとマニュアルテストの実施
3. **ドキュメント整備**: 変更内容の詳細な記録
4. **ロールバック計画**: 問題発生時の復旧手順

## 成功指標

### Phase 1 (リリース優先)
- [x] 機能開発の継続
- [ ] 予定通りのリリース完了
- [ ] 重大なバグの発生なし

### Phase 2 (Service層導入)
- [ ] コードの可読性向上
- [ ] ビジネスロジックの再利用性向上
- [ ] テスト容易性の向上
- [ ] 開発速度の向上（長期的）

### Phase 3 (最適化)
- [ ] パフォーマンス改善
- [ ] 保守性向上
- [ ] 技術的負債の解消

## 関連ドキュメント

- [プロジェクト概要](../CLAUDE.md)
- [開発環境設定](development-environment.md)
- [実装ログ](implement_logs/)