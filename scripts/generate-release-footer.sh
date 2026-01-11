#!/bin/bash
set -euo pipefail

# ライセンステーブル部分を生成
LICENSE_TABLE=$(./scripts/generate-license-table.sh)

# Go言語のバージョンを取得
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

# フッター全体を構築
cat <<EOF
---

## 📦 インストール方法 / Installation

1. \`thlocalsync-vX.X.X-windows-XXX.zip\` をダウンロード
2. ZIPファイルを解凍
3. \`thlocalsync.exe\` をUSBメモリの適切な場所に配置
4. コマンドプロンプトまたはPowerShellから実行

詳細な使用方法は [README.md](https://github.com/otagao/touhou-local-sync/blob/main/README.md) をご確認ください。

---

## ⚖️ ライセンス情報 / License Information

**本ソフトウェアのライセンス:**
- MIT License（詳細は [LICENSE](https://github.com/otagao/touhou-local-sync/blob/main/LICENSE) を参照）
- Copyright (c) 2025 otagao

**含まれるオープンソースコンポーネント:**

このリリースには以下のオープンソースライブラリが含まれています：

${LICENSE_TABLE}
| [Go Language](https://go.dev/) | ${GO_VERSION} | BSD-3-Clause |

完全なライセンステキストは [NOTICE](https://github.com/otagao/touhou-local-sync/blob/main/NOTICE) ファイルをご確認ください。

---

## 🔐 セキュリティ / Security

セキュリティ上の問題を発見した場合は、公開イシューではなく[セキュリティアドバイザリ](https://github.com/otagao/touhou-local-sync/security/advisories/new)で報告してください。

---

## 🙏 謝辞 / Acknowledgments

このプロジェクトは上記のオープンソースコミュニティの素晴らしい成果物の上に成り立っています。
各プロジェクトのメンテナーと貢献者の皆様に感謝いたします。

---

**Full Changelog**: https://github.com/otagao/touhou-local-sync/compare/\$PREVIOUS_TAG...\$CURRENT_TAG
EOF
