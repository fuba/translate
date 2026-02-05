# translate

OpenAI 互換 API (llama.cpp) を使って、テキスト/Markdown/PDF を翻訳する CLI です。

## 使い方

```sh
translate --from en --to ja --in input.txt --out output.txt
cat input.md | translate --format md --to ja > output.md
translate --format pdf --in input.pdf --out output.pdf
```

## インストール（make install）

```sh
make install
```

`~/.local/bin/translate` にインストールされます（`PREFIX`/`BINDIR` で変更可）。
`config.json` の雛形を `~/.config/translate/config.json` に作成します。
`base_url` を自分の API エンドポイントに設定してください。

### 主なオプション

- `--format` : `text|md|pdf|auto`（デフォルト `auto`）
- `--in` / `--out` : 入出力パス。省略時は stdin/stdout
- `--from` : 翻訳元言語（デフォルト `auto`）
- `--to` : 翻訳先言語（デフォルト `LANG` から推定）
- `--model` : 既定 `gpt-oss-20b`
- `--base-url` : OpenAI 互換 API の base URL（必須）
- `--api-key` : API キー（省略時は `OPENAI_API_KEY`）
- `--timeout` : HTTP タイムアウト（既定 120s）
- `--max-chars` : 翻訳 API への最大文字数（既定 2000、0 で無効）
- `--verbose` : 翻訳途中のテキストを stderr に逐次出力
- `--verbose-prompt` : 送信するプロンプトを stderr に出力
- `--silent` : 進捗表示を抑制
- 端末実行時は、ファイル出力かつ verbose ではない場合に簡易プログレス表示を stderr に出します（総数が計算できる場合は割合を表示）
- `--endpoint` : `chat|completion|auto`（既定 `completion`）
- `--passphrase-ttl` : パスフレーズキャッシュ（既定 10m、0 で無効）
- `--dump-extracted` : PDF の生テキスト抽出を出力（パス指定、`-` で stdout）
- `--pdf-font` : PDF オーバーレイ用の TTF フォント（既定: `~/.config/translate/fonts/LINESeedJP-Regular.ttf`）

`--base-url` は `http://your-host:8080` または `http://your-host:8080/v1` を指定できます。内部で `/v1/*` を付与します。

## 設定ファイル

設定ファイルは `~/.config/translate/config.json`（`XDG_CONFIG_HOME` があればそちら）に保存されます。

```sh
translate config set --base-url http://your-host:8080 --model gpt-oss-20b --max-chars 2000 --endpoint completion
```

PDF 用フォントを固定したい場合は `--pdf-font` を保存できます。

## PDF について

- UniPDF (unidoc/unipdf) v4 を使用します。
- `UNIDOC_LICENSE_API_KEY` を暗号化保存できます。
- PDF は **抽出した行ごとに翻訳し、白背景でオーバーレイ描画**します。

### UNIDOC キーの保存

```sh
translate auth set-unidoc
```

- パスフレーズで暗号化して `~/.config/translate/unidoc.key` に保存します。
- 復号には `TRANSLATE_PASSPHRASE` を使うか、対話入力します。
- `--passphrase-ttl` を指定すると、一定時間は再入力を省略できます。

### PDF 翻訳

- PDF の行単位で翻訳しオーバーレイ描画します（レイアウト維持を優先）。
- 日本語を描画する場合は `--pdf-font` で日本語対応 TTF を指定してください。

### フォントのインストール（LINE Seed JP）

```sh
./scripts/install-fonts.sh
```

インストール先: `~/.config/translate/fonts/LINESeedJP-Regular.ttf`（デフォルトで使用）

### PDF 抽出テキストの確認

```sh
translate --format pdf --in input.pdf --dump-extracted -
```

## 必要環境

- Go 1.23 以上（UniPDF v4 要件）
- `UNIDOC_LICENSE_API_KEY`（PDF 翻訳時）

## ライセンス

CC0
