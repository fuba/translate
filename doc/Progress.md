# Progress

## 2026-02-04
- CLI: `translate` を実装（text/markdown/pdf 対応）
- OpenAI 互換 API クライアント（gpt-oss-20b）を実装
- Markdown: AST 解析でコード/HTMLを除外しテキストのみ翻訳
- PDF: UniPDF の content stream を翻訳して置換（レイアウト維持優先）
- README を追加（使い方、UNIDOC ライセンス要件）

## Notes
- PDF 翻訳は元フォントに翻訳先の文字が含まれない場合、文字化け/欠落の可能性あり
- `UNIDOC_LICENSE_API_KEY` が必須
