# 東方Project セーブデータ同期ツール（Windows CLI版）

`thlocalsync` - 東方Projectのセーブデータを複数のPCで同期するためのCLIツール

> **⚠️ 開発中**: このプロジェクトは現在開発初期段階です。実装はまだ完了しておらず、動作する実行ファイルはありません。

## 概要

複数のWindows PCで東方Project原作STGのセーブデータ（例：`score.dat`）を**USBメモリ上の正本（ハブ）**に集約し、手動でPull/Pushするオフライン同期ツールです。

### 特徴

- 完全オフライン、USB常駐、単一実行ファイル
- タイトル別の保存パスを自動検出＋対話的に登録/編集
- mtime・ハッシュ・サイズの三点で新旧/正誤判定
- 履歴バックアップと安全なアトミック書き換え

## 対象環境

- **OS**: Windows 10/11 (x64)
- **実行形態**: 単一 exe（USB直置き）
- **権限**: 標準ユーザ（管理者不要）
- **開発言語**: Go 1.23+

## インストール

### ビルド方法

```bash
# Windowsバイナリのビルド（クロスコンパイル）
go build -o thlocalsync.exe ./cmd/thlocalsync
```

### USB内ディレクトリ構成

```
/THLocalSync/
  thlocalsync.exe
  /data/
    devices.json
    paths.json
    rules.json
  /vault/
    th06/
      main/score.dat
      _history/2025-11-11T06-20Z-score.dat
    th08/
      main/score.dat
  /logs/
    2025-11-11.log
```

## 使い方

### 初回セットアップ

1. USBメモリを接続
2. セーブデータを自動検出して登録:

```bash
thlocalsync detect
```

### 基本的な使用フロー

1. ゲームプレイ後、ローカルからUSBへ保存（Pull）:

```bash
thlocalsync pull all
```

2. 別PCで、USBからローカルへ配布（Push）:

```bash
thlocalsync push all
```

3. 状態確認:

```bash
thlocalsync status all
```

### コマンド一覧

| コマンド | 機能 | 例 |
|---------|------|-----|
| `detect` | 自動検出 + 対話登録 | `thlocalsync detect` |
| `status [title\|all]` | USBとローカルの差分一覧 | `thlocalsync status all` |
| `pull [title\|all]` | ローカル → USB（正本へ吸い上げ） | `thlocalsync pull th08` |
| `push [title\|all]` | USB → ローカル（配布） | `thlocalsync push all` |
| `backup [title] [--list\|--restore <name>]` | 履歴表示/復元 | `thlocalsync backup th08 --list` |

## 対応タイトル

| タイトル | コード | セーブデータパス例 |
|---------|--------|------------------|
| 東方紅魔郷 | th06 | `<game_dir>\score.dat` |
| 東方妖々夢 | th07 | `<game_dir>\score.dat` |
| 東方永夜抄 | th08 | `<game_dir>\score.dat` |
| 東方花映塚 | th09 | `<game_dir>\score.dat` |
| 東方風神録 | th10 | `<game_dir>\score.dat` |
| 東方地霊殿 | th11 | `%APPDATA%\ShanghaiAlice\th11\scoreth11.dat` |
| 東方星蓮船 | th12 | `%APPDATA%\ShanghaiAlice\th12\scoreth12.dat` |
| 東方神霊廟 | th13 | `%APPDATA%\ShanghaiAlice\th13\scoreth13.dat` |
| ... | ... | ... |

## 開発

### プロジェクト構造

```
.
├── cmd/
│   └── thlocalsync/    # エントリーポイント
├── pkg/
│   ├── config/         # JSON設定ファイルI/O
│   ├── device/         # デバイスID生成
│   ├── pathdetect/     # パス検出＋対話登録
│   ├── sync/           # Pull/Push・判定ロジック
│   ├── backup/         # 履歴保存/復元
│   ├── process/        # プロセス/ロック検知
│   ├── logger/         # 構造化ログ
│   └── utils/          # ハッシュ/アトミックコピー
├── internal/
│   └── models/         # 内部データモデル
└── docs/
    └── specs_windows_cli.md  # 仕様書
```

### ビルド

```bash
# 開発用（現在のプラットフォーム）
go build ./cmd/thlocalsync

# Windows向けクロスコンパイル（Macから）
GOOS=windows GOARCH=amd64 go build -o thlocalsync.exe ./cmd/thlocalsync
```

### テスト

```bash
go test ./...
```

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 注意事項

- このツールは東方Projectの二次創作物です
- セーブデータのバックアップは自己責任で行ってください
- ゲーム実行中はPull/Push操作を行わないでください
