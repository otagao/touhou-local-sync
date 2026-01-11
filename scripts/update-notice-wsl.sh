#!/bin/bash
# WSL環境でNOTICEファイルを更新するスクリプト
# 使用方法: WSL Ubuntuターミナルから以下を実行
#   cd /mnt/c/Users/smelt00/gits/touhou-local-sync
#   bash scripts/update-notice-wsl.sh

set -e

# Go 1.23がインストールされているか確認
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed in WSL"
    echo "Please install Go 1.23.4 first:"
    echo "  wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz"
    echo "  sudo rm -rf /usr/local/go"
    echo "  sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz"
    echo "  export PATH=\$PATH:/usr/local/go/bin"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "Using Go version: $GO_VERSION"

# Go 1.25だと警告
if [[ "$GO_VERSION" == *"go1.25"* ]]; then
    echo "Warning: Go 1.25 detected. go-licenses may not work correctly."
    echo "Consider downgrading to Go 1.23.4 for NOTICE generation."
fi

# PATHにGOPATH/binを追加
export PATH=$PATH:$(go env GOPATH)/bin

# go-licensesをインストール
echo "Installing go-licenses..."
go install github.com/google/go-licenses@v1.6.0

# NOTICEファイルを生成
echo "Generating NOTICE file..."
go-licenses report ./cmd/thlocalsync \
    --template=scripts/notice.tmpl \
    --ignore=github.com/otagao/touhou-local-sync \
    --ignore=std > NOTICE 2>/dev/null || {
        echo "go-licenses reported errors, but checking output..."
    }

if [ -s NOTICE ] && grep -q "github.com" NOTICE; then
    echo "✓ NOTICE file generated successfully"
    echo "File size: $(wc -c < NOTICE) bytes"
    echo "Dependencies found: $(grep -c "^## " NOTICE || echo "0")"
else
    echo "✗ Error: NOTICE generation failed"
    rm -f NOTICE
    exit 1
fi
