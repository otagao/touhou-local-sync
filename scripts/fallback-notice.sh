#!/bin/bash
# フォールバック用のNOTICE生成スクリプト
# go-licenses reportが失敗した場合の代替手段

set -e

CMD_PATH="${1:-./cmd/thlocalsync}"
OUTPUT_FILE="${2:-NOTICE}"

echo "Attempting fallback NOTICE generation..."

# Step 1: go list で依存関係を取得
echo "Step 1: Fetching dependencies using go list..."

# jqが使えるか確認
if command -v jq &> /dev/null; then
    echo "Using jq for JSON parsing..."
    DEPS=$(go list -m -json all | jq -r 'select(.Path != "github.com/otagao/touhou-local-sync" and .Path != "std") | .Path + "@" + .Version')
else
    echo "jq not found, using go list -m all..."
    # jqがない場合はシンプルなフォーマットを使用
    DEPS=$(go list -m all | grep -v "^github.com/otagao/touhou-local-sync" | tail -n +2)
fi

if [ -z "$DEPS" ]; then
    echo "Error: No dependencies found"
    exit 1
fi

# Step 2: NOTICEファイルを生成
cat > "$OUTPUT_FILE" <<'HEADER'
NOTICE - Third-Party Software Licenses
=======================================

This software includes components distributed under the following licenses:

HEADER

INDEX=1
echo "$DEPS" | while IFS= read -r dep; do
    if [ -z "$dep" ] || [ "$dep" = "std" ]; then
        continue
    fi

    # '@'で分割（jq使用時）または空白で分割（go list -m all使用時）
    if [[ "$dep" == *"@"* ]]; then
        MODULE_PATH=$(echo "$dep" | cut -d'@' -f1)
        VERSION=$(echo "$dep" | cut -d'@' -f2)
    else
        MODULE_PATH=$(echo "$dep" | awk '{print $1}')
        VERSION=$(echo "$dep" | awk '{print $2}')
        [ -z "$VERSION" ] && VERSION="(unknown)"
    fi

    echo "-------------------------------------------------------------------------------" >> "$OUTPUT_FILE"
    echo "$INDEX. $MODULE_PATH" >> "$OUTPUT_FILE"
    echo "   Version: $VERSION" >> "$OUTPUT_FILE"
    echo "   License: (See module repository for details)" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"

    INDEX=$((INDEX + 1))
done

cat >> "$OUTPUT_FILE" <<'FOOTER'
===============================================================================

For the complete list of dependencies, please refer to the go.mod file in the
source repository.

Full license texts can be found at:
- Apache License 2.0: https://www.apache.org/licenses/LICENSE-2.0
- MIT License: https://opensource.org/licenses/MIT
- BSD Licenses: https://opensource.org/licenses/BSD-3-Clause
FOOTER

echo "✓ Fallback NOTICE file generated successfully"
echo "File size: $(wc -c < "$OUTPUT_FILE") bytes"
