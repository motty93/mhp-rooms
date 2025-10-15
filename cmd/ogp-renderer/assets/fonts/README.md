# フォントファイル

このディレクトリには、OGP画像生成に使用するフォントファイルが配置されています。

## 必要なフォント

- **NotoSansJP-Bold.ttf**: タイトル用
- **NotoSansJP-Regular.ttf**: サブ情報用

## セットアップ

フォントファイルが欠落している場合、プロジェクトルートから以下のスクリプトを実行してください：

```bash
./scripts/setup-fonts.sh
```

または、手動でダウンロード：

```bash
cd cmd/ogp-renderer/assets/fonts

# NotoSansJP Bold
curl -L "https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP-Bold.ttf" -o NotoSansJP-Bold.ttf

# NotoSansJP Regular
curl -L "https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP-Regular.ttf" -o NotoSansJP-Regular.ttf
```

ダウンロード後、これらのフォントファイルをコミットしてください。

## ライセンス

Noto Sans JPは、Open Font License (OFL)の下で配布されています。
商用・非商用問わず、自由に使用でき、再配布も可能です。

詳細: https://scripts.sil.org/OFL
