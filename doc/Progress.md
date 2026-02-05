# Progress

## 2026-02-04
- CLI: `translate` を実装（text/markdown/pdf 対応）
- OpenAI 互換 API クライアント（gpt-oss-20b）を実装
- Markdown: AST 解析でコード/HTMLを除外しテキストのみ翻訳
- PDF: UniPDF の content stream を翻訳して置換（レイアウト維持優先）
- README を追加（使い方、UNIDOC ライセンス要件）
- UNIDOC キー暗号化保存（パスフレーズ、~/.config/translate/unidoc.key）
- 設定ファイル保存（~/.config/translate/config.json）
- verbose モードで翻訳結果を逐次出力
- 翻訳入力のチャンク分割（max-chars, 句読点優先）
- completion エンドポイント対応（auto 判定）
- パスフレーズの一時キャッシュ対応
- エンドポイント既定を completion に変更
- completion プロンプトを Harmony 形式に変更
- PDF 抽出テキストのダンプ機能を追加

## Notes
- PDF 翻訳は元フォントに翻訳先の文字が含まれない場合、文字化け/欠落の可能性あり
- UNIDOC キーは `UNIDOC_LICENSE_API_KEY` または暗号化ストアから取得
