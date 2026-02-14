# 東方Project セーブデータ同期ツール（Windows CLI版）

`thlocalsync` - 東方Projectのセーブデータを複数のPCで同期するためのCLIツール

> **⚠️ 重要な注意**: このプロジェクトは初期リリース版です。重大な問題が含まれている可能性があります。**必ずセーブデータのバックアップを取った上で試用してください。** データ損失について開発者は一切責任を負いません。

## 概要

Windows PC上に存在する東方Project原作STGのセーブデータ（スコア、リプレイ、スナップショット）を、半自動で保存・上書き配布することができるツールです。

複数デバイスを認識し、それぞれ異なる保存場所を記録することができます。USBメモリなどのポータブルストレージ上で運用することで、複数端末間でオフライン同期を行うことが可能になります。

### 特徴

- 完全オフライン、ポータブルストレージ常駐、単一実行ファイル
- タイトル別の保存パスを半自動認識＋対話的に登録/編集
- mtime・ハッシュ・サイズの三点で新旧/正誤判定
- 履歴バックアップと安全なアトミック書き換え

### 前提条件

このツールは以下の環境を想定しています：
- **Windows 10/11の一般的なファイル構造**
- **ゲーム本体の実行ファイルが単一のフォルダ以下に配置されていること**
  - まとまって配置されていなくてもバックアップは可能ですが、最初に保存場所を認識させる際に自動認識が上手く動作しない可能性があります

例：
```
D:\Games\Touhou\
  東方紅魔郷\
    th06.exe
  東方妖々夢\
    th07.exe
  東方永夜抄\
    th08.exe
  ...
```

## 対象環境

- **OS**: Windows 10/11 (x64)
- **実行形態**: 単一 exe（ポータブルストレージ直置き）
- **権限**: 標準ユーザ（管理者不要）
- **開発言語**: Go 1.25+

## 使い方

### 初回セットアップ

1. ポータブルストレージを接続し、保存先としたいディレクトリにthlocalsync.exeを配置
2. セーブデータを半自動認識して登録:

```bash
thlocalsync detect
```

対話型ダイアログが不要な場合は以下のオプションでパスを指定できます：

```bash
thlocalsync detect --gamedir "D:\Games\Touhou"
```

### 基本的な使用フロー

1. ゲームプレイ後、ローカルからポータブルストレージへ保存（Pull）:

```bash
thlocalsync pull all
```

2. 別PCで、ポータブルストレージからローカルへ配布（Push）:

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
| `detect` | 半自動認識 + 対話登録 | `thlocalsync detect` |
| `status [title\|all]` | ポータブルストレージとローカルの差分一覧 | `thlocalsync status all` |
| `pull [title\|all]` | ローカル → ポータブルストレージ（正本へ吸い上げ） | `thlocalsync pull th08` |
| `push [title\|all]` | ポータブルストレージ → ローカル（配布） | `thlocalsync push all` |
| `backup [title] [--list\|--restore <name>]` | 履歴表示/復元 | `thlocalsync backup th08 --list` |

## 対応タイトル

東方紅魔郷から東方錦上京まで、小数点作品を含めた全22タイトルの原作STGに対応しています。

### セーブデータの保存場所

- **th06-th09**: ゲームディレクトリまたはVirtualStore
- **th095, th10**: ゲームディレクトリまたはVirtualStore（`scorethXX.dat`形式）
- **th11-th12**: ゲームディレクトリ（`scorethXX.dat`形式）
- **th125以降**: `%APPDATA%\ShanghaiAlice\thXXX\scorethXXX.dat`

## 開発

### プロジェクト構造

```
.
├── cmd/
│   └── thlocalsync/    # エントリーポイント
├── pkg/
│   ├── config/         # JSON設定ファイルI/O
│   ├── device/         # デバイスID生成
│   ├── pathdetect/     # パス半自動認識＋対話登録
│   ├── sync/           # Pull/Push・判定ロジック
│   ├── backup/         # 履歴保存/復元
│   ├── process/        # プロセス/ロック検知
│   ├── logger/         # 構造化ログ
│   └── utils/          # ハッシュ/アトミックコピー
├── internal/
│   └── models/         # 内部データモデル
```

### ビルド

```bash
# 開発用（現在のプラットフォーム）
go build ./cmd/thlocalsync

# Windows向けクロスコンパイル（Mac/Linuxから）
GOOS=windows GOARCH=amd64 go build -o thlocalsync.exe ./cmd/thlocalsync
```

### テスト

```bash
go test ./...
```


### 依存関係の自動更新

このプロジェクトは [Dependabot](https://docs.github.com/en/code-security/dependabot) による依存関係の自動更新を有効化しています。

- **通常更新**: 毎週月曜に自動PR作成
- **セキュリティ更新**: 毎日チェック＋即座にPR作成
- **自動マージ**: patch/minor アップデートは自動承認（ライセンス検証通過後）

開発者は NOTICE ファイルの手動更新のみ必要です。

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 注意事項

- このツールは東方Projectの二次創作物です
- セーブデータのバックアップは自己責任で行ってください
- ゲーム実行中はPull/Push操作を行わないでください
