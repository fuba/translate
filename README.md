# translate

OpenAI 互換 API (llama.cpp) を使って、テキスト/Markdown/PDF を翻訳する CLI です。

## 使い方

```sh
translate --from en --to ja --in input.txt --out output.txt
cat input.md | translate --format md --to ja > output.md
translate --format pdf --in input.pdf --out output.pdf
```

### 主なオプション

- `--format` : `text|md|pdf|auto`（デフォルト `auto`）
- `--in` / `--out` : 入出力パス。省略時は stdin/stdout
- `--from` : 翻訳元言語（デフォルト `auto`）
- `--to` : 翻訳先言語（デフォルト `LANG` から推定）
- `--model` : 既定 `gpt-oss-20b`
- `--base-url` : 既定 `http://kirgizu:8080`
- `--api-key` : API キー（省略時は `OPENAI_API_KEY`）
- `--timeout` : HTTP タイムアウト（既定 120s）
- `--max-chars` : 翻訳 API への最大文字数（既定 2000、0 で無効）
- `--verbose` : 翻訳途中のテキストを stderr に逐次出力
- `--endpoint` : `chat|completion|auto`（既定 `completion`）
- `--passphrase-ttl` : パスフレーズキャッシュ（既定 10m、0 で無効）
- `--dump-extracted` : PDF の生テキスト抽出を出力（パス指定、`-` で stdout）

`--base-url` は `http://kirgizu:8080` または `http://kirgizu:8080/v1` を指定できます。内部で `/v1/*` を付与します。

## 設定ファイル

設定ファイルは `~/.config/translate/config.json`（`XDG_CONFIG_HOME` があればそちら）に保存されます。

```sh
translate config set --base-url http://kirgizu:8080 --model gpt-oss-20b --max-chars 2000 --endpoint completion
```

## PDF について

- UniPDF (unidoc/unipdf) v4 を使用します。
- `UNIDOC_LICENSE_API_KEY` を暗号化保存できます。

### UNIDOC キーの保存

```sh
translate auth set-unidoc
```

- パスフレーズで暗号化して `~/.config/translate/unidoc.key` に保存します。
- 復号には `TRANSLATE_PASSPHRASE` を使うか、対話入力します。
- `--passphrase-ttl` を指定すると、一定時間は再入力を省略できます。

### PDF 翻訳

- PDF のテキストオブジェクトを翻訳して置換します（レイアウト維持を優先）。
- 元のフォントに翻訳先の文字が含まれない場合、文字化け/欠落の可能性があります。

### PDF 抽出テキストの確認

```sh
translate --format pdf --in input.pdf --dump-extracted -
```

## 必要環境

- Go 1.23 以上（UniPDF v4 要件）
- `UNIDOC_LICENSE_API_KEY`（PDF 翻訳時）

## ライセンス

CC0
