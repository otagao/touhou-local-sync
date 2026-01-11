#!/bin/bash
set -euo pipefail

CMD_PATH="${1:-./cmd/thlocalsync}"

# GOTOOLCHAIN=localの影響を回避（Go 1.25要求エラーを防ぐ）
unset GOTOOLCHAIN

# go-licensesで依存関係のライセンス情報を取得（エラー時はフォールバック）
if ! LICENSE_TABLE=$(go-licenses report "$CMD_PATH" \
  --template=scripts/license-table.tmpl \
  --ignore=github.com/otagao/touhou-local-sync \
  --ignore=std 2>/dev/null); then

  # フォールバック: go listで依存関係を取得してシンプルなテーブルを生成
  echo "Warning: go-licenses failed, using fallback method" >&2

  # テーブルヘッダー
  echo "| Library | Version | License |"
  echo "|---------|---------|---------|"

  # 依存関係を取得（GOTOOLCHAIN制約を無視）
  GOTOOLCHAIN=auto go list -m all | grep -v "^github.com/otagao/touhou-local-sync" | tail -n +2 | while IFS= read -r dep; do
    MODULE_PATH=$(echo "$dep" | awk '{print $1}')
    VERSION=$(echo "$dep" | awk '{print $2}')

    # ライセンス種別は "See repository" とする（go-licensesなしでは取得不可）
    echo "| [$MODULE_PATH](https://pkg.go.dev/$MODULE_PATH) | $VERSION | See repository |"
  done
else
  # go-licensesが成功した場合はその結果を出力
  echo "$LICENSE_TABLE"
fi

# 注: go-licensesは自動的にgo.modを読み込み、
# 直接・間接依存をすべて収集します
