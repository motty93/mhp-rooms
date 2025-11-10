# Cloud Buildのビルド時間最適化

## 実装日時
- 開始: 2025-11-10
- 完了: 2025-11-10
- 所要時間: 約30分

## 概要
Cloud Buildのビルド時間が14~15分かかり、課金が発生していた問題を解決するため、コスト重視でDockerfileとCloud Buildの設定を最適化しました。

## 問題の詳細
- `cloudbuild.stg.yml`の実行時間が14~15分
- 課金が発生している
- キャッシュが効いていない可能性

## 実装内容

### 1. メインDockerfileの最適化（`Dockerfile`）

#### 変更点
- **BuildKitのキャッシュマウント機能を活用**
  - `--mount=type=cache,target=/go/pkg/mod`: Go modulesのキャッシュ
  - `--mount=type=cache,target=/root/.cache/go-build`: Go buildのキャッシュ

- **レイヤーの最適化**
  - 依存パッケージのインストールを独立したレイヤーに分離
  - `go mod download`を独立したステップに分離してキャッシュ効率向上
  - 静的アセット生成（`go run ./cmd/generate_info/main.go`）にもキャッシュマウント適用

#### 期待される効果
- 2回目以降のビルドで`go mod download`がスキップされる
- Go buildのコンパイルキャッシュが再利用される
- コード変更時でも依存関係が変わらなければキャッシュが効く

### 2. OGP Renderer Dockerfileの最適化（`cmd/ogp-renderer/Dockerfile`）

#### 変更点
- メインDockerfileと同じキャッシュマウント戦略を適用
- `go mod download`と`go build`にキャッシュマウントを追加

#### 期待される効果
- Jobイメージのビルド時間も短縮
- App と Job の並列ビルドの効率向上

### 3. Cloud Build設定の最適化（`cloudbuild.stg.yml`）

#### 変更点
- **volumesオプションの追加**
  ```yaml
  volumes:
    - name: go-modules
      path: /go/pkg/mod
    - name: go-build-cache
      path: /root/.cache/go-build
  ```

#### 期待される効果
- Cloud Build実行間でキャッシュが永続化される
- 連続してビルドを実行した場合に効果が大きい

## 技術的な詳細

### BuildKitのキャッシュマウント
```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build ...
```
- BuildKitの機能で、指定したディレクトリをビルド間で共有
- イメージには含まれないため、イメージサイズは増えない
- 同じDockerホスト上でのビルドでキャッシュが再利用される

### Cloud Buildのvolumes
- Cloud Build VM上で複数のステップ間でディレクトリを共有
- 同じビルド内での並列ステップ（App と Job）でキャッシュが共有される
- ビルド実行間での永続化は限定的（同じVMが再利用される場合のみ）

## 期待される改善効果

### ビルド時間
- **初回ビルド**: ほぼ変わらず（14~15分）
- **2回目以降（コード変更のみ）**: 5~8分に短縮（予想）
- **2回目以降（依存関係変更なし）**: 3~5分に短縮（予想）

### コスト削減
- ビルド時間が1/3になると仮定: 月間コストが約66%削減
- E2_MEDIUMのまま維持してコスト増加なし

## 注意事項

### BuildKit必須
- `DOCKER_BUILDKIT=1`が設定されていることが前提
- Cloud Buildでは既に設定済み（`cloudbuild.stg.yml`の37行目）

### キャッシュの効果
- Cloud BuildのVMが変わるとキャッシュが無効化される可能性あり
- 頻繁にビルドする環境ほど効果が大きい

### 互換性
- Dockerのバージョンが古い環境では動作しない可能性
- ローカル開発環境でもBuildKitを有効にする必要あり

## 検証方法

### 1. 次回のビルド時間を計測
```bash
# Cloud Buildのログで実行時間を確認
gcloud builds list --limit=5
```

### 2. キャッシュの効果確認
- ビルドログで「CACHED」と表示されるステップを確認
- `go mod download`がスキップされているか確認

### 3. コスト確認
- Google Cloud Console の課金ページで Cloud Build の料金推移を確認
- 1週間～1ヶ月のスパンで比較

## 今後の改善案（必要に応じて）

### さらなる高速化が必要な場合
1. **Kanikoビルダーの導入**
   - Cloud Build専用の最適化されたビルダー
   - キャッシュがより確実に効く
   - 設定変更のみで実装可能（コスト増加なし）

2. **マシンタイプのアップグレード**
   - `E2_MEDIUM` → `E2_HIGHCPU_8`
   - ビルド時間が1/2～1/3になる可能性
   - コストは上がるが、総合的には下がる可能性

3. **静的アセット生成の分離**
   - `go run ./cmd/generate_info/main.go`を別のジョブに分離
   - 変更がないときは再生成をスキップ
   - GCS/CDNへのアップロードに移行（TODO参照）

## 参考資料
- [Docker BuildKit Documentation](https://docs.docker.com/build/buildkit/)
- [Cloud Build - Caching builds](https://cloud.google.com/build/docs/optimize-builds/speeding-up-builds)
- [Go build cache](https://pkg.go.dev/cmd/go#hdr-Build_and_test_caching)
