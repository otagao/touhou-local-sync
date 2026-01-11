#!/bin/bash
set -euo pipefail

# go-licensesで依存関係のライセンス情報を取得
go-licenses report ./cmd/thlocalsync \
  --template=scripts/license-table.tmpl \
  --ignore=github.com/otagao/touhou-local-sync \
  --ignore=std

# 注: go-licensesは自動的にgo.modを読み込み、
# 直接・間接依存をすべて収集します
