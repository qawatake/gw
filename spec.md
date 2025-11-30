# git worktree wrapper - gw

git worktreeを便利にラップするコマンドラインツール

## 概要

複数のブランチで作業する際に、git worktreeの作成・管理・移動を効率化するためのツール。

## サブコマンド

### `gw add`

新しいブランチを作成し、対応するgit worktreeを作成する。

#### 動作

1. ブランチ名をインタラクティブに入力させる（vcatと同じ仕組み）
   - 標準入力を一時ファイルに保存
   - エディタ（vim）で編集
   - 編集内容をブランチ名として使用
2. ブランチ名のプレフィックスは `{user}/YYYY/MM/DD/` 形式
   - `{user}` は `git config user.name` から取得（小文字化・スペースをハイフンに変換）
   - user.nameが設定されていない場合はエラー
   - 環境変数 `GW_BRANCH_PREFIX` で上書き可能
3. worktreeのディレクトリ名はブランチ名から生成
   - スラッシュをハイフンに置換するなど、ファイルシステム用に正規化
4. worktreeを作成
   - リポジトリのdirname（例: `gw`）を親ディレクトリとする
   - パス例: `~/.worktrees/gw/sample-user-2025-11-24-feature-login`
   - これにより複数のリポジトリのworktreeを整理して管理できる

#### 実装の参考

- `~/bin/git-c`: ブランチ名のプレフィックス形式
- `~/bin/vcat`: インタラクティブな入力の受け取り方

### `gw list` (alias: `gw ls`)

worktreeの一覧を表示する。

#### 動作

1. `git worktree list` の結果を取得
2. 日付順（新しい順）にソート
3. 見やすい形式で表示

#### 並び順

- `~/bin/git-swt` と同じく、`--sort=-authordate` 相当の順序
- 最近作業したworktreeが上に来るようにする

#### エイリアス

- `gw ls` でも同じ動作をする（短縮形）

### `gw cd`

worktreeを選択して移動する。

#### 動作

1. worktreeの一覧を表示（listと同じ並び順）
2. インタラクティブな選択UI（peco等）で選択
3. 選択したworktreeに移動

#### 実装の仕組み

**シェルラッパー関数方式**（`https://github.com/tobi/try` と同じ仕組み）

1. `gw` 本体はコマンドを出力するだけ
2. シェル初期化スクリプトでラッパー関数を定義
3. `gw cd` の結果をシェルが `eval` することで実際のディレクトリ移動を実現

例：
```bash
# シェル初期化（.bashrc/.zshrc等）
eval "$(gw init)"

# 使用時
# gw cd は "cd /path/to/worktree" という文字列を出力
# ラッパー関数が eval して実際に移動
```

#### 各シェルのサポート

- Bash
- Zsh
- Fish（構文が異なるため個別対応）

### `gw rm` (旧: `gw clean`)

不要なworktreeを削除する。

#### 動作

1. worktreeの一覧を表示（listと同じ並び順）
   - ただし、main worktreeは表示しない（削除対象外）
2. インタラクティブな複数選択UI
   - fzf: スペースキーでチェックボックスのトグル、エンターで確定
   - peco（フォールバック）: 1つずつ選択、"Done"で確定
3. 確認プロンプトを表示してから削除実行
4. 削除処理
   - `git worktree remove` を実行
   - 紐づくブランチも自動的に削除

## 技術スタック

- 言語: Go（クロスプラットフォーム対応、シングルバイナリ配布のため）
- インタラクティブUI:
  - 単一選択（`gw cd`）: peco（必須）
  - 複数選択（`gw clean`）: fzf優先、なければpecoにフォールバック
  - テキスト入力: エディタ起動（vimまたは$EDITOR）

## ディレクトリ構成

```
.
├── cmd/
│   └── gw/
│       └── main.go
├── internal/
│   ├── worktree/     # worktree操作
│   ├── branch/       # ブランチ操作
│   ├── ui/           # インタラクティブUI
│   └── shell/        # シェルラッパー生成
├── go.mod
├── go.sum
└── README.md
```

## 設定

環境変数またはコンフィグファイルで以下を設定可能にする：

- `GW_WORKTREE_ROOT`: worktreeを作成する親ディレクトリ（デフォルト: `~/.worktrees`）
- `GW_EDITOR`: 使用するエディタ（デフォルト: `$EDITOR` または `vim`）
- `GW_BRANCH_PREFIX`: ブランチ名のプレフィックス形式（デフォルト: `{git-user-name}/{date}/`）
  - `{date}` は自動的に `YYYY/MM/DD` に置換される
  - デフォルトでは `git config user.name` を小文字化してスペースをハイフンに変換したもの

## インストール

```bash
# シェル初期化
echo 'eval "$(gw init)"' >> ~/.bashrc  # or ~/.zshrc
```

## 使用例

```bash
# 新しいworktreeを作成
$ gw add
# エディタが開く → "feature-login" と入力
# → ブランチ "sample-user/2025/11/24/feature-login" を作成（user.nameに基づく）
# → worktree を ~/.worktrees/gw/sample-user-2025-11-24-feature-login に作成

# worktree一覧
$ gw list  # または gw ls
sample-user/2025/11/24/feature-login    ~/.worktrees/gw/sample-user-2025-11-24-feature-login
sample-user/2025/11/23/bugfix-auth      ~/.worktrees/gw/sample-user-2025-11-23-bugfix-auth
main                                          ~/src/myproject

# worktreeに移動
$ gw cd
# pecoで選択 → 選択したディレクトリに移動

# worktreeを削除
$ gw rm
# fzf: スペースキーで複数選択 → エンターで削除確認 → 削除実行
# peco: 1つずつ選択、"Done"で確定 → 削除確認 → 削除実行
```

### `gw ln`

gitignoreされているファイルをworktree間で共有するためのコマンド。

#### 概要

- 共有ファイルの実体は `~/.worktrees/{repo}/.gw-links/` に保存
- 各worktreeからはシンボリックリンクでアクセス
- シンボリックリンクは絶対パスで作成

#### サブコマンド

##### `gw ln add <path>`

ファイルまたはディレクトリを共有対象として登録する。

###### 動作

1. 指定されたパスが存在するか確認（存在しない場合はエラー）
2. `~/.worktrees/{repo}/.gw-links/` ディレクトリがなければ作成
3. 対象ファイル/ディレクトリを `.gw-links/` に移動
4. 元の場所にシンボリックリンクを作成（絶対パス）

###### 例

```bash
$ gw ln add .env
# .env を ~/.worktrees/gw/.gw-links/.env に移動
# 現在のworktreeに .env → ~/.worktrees/gw/.gw-links/.env のシンボリックリンクを作成

$ gw ln add tmp/cache
# tmp/cache ディレクトリを ~/.worktrees/gw/.gw-links/tmp/cache に移動
# 現在のworktreeに tmp/cache → ~/.worktrees/gw/.gw-links/tmp/cache のシンボリックリンクを作成
```

###### 注意

- glob パターンは非対応（ファイル・ディレクトリのパスのみ）
- 既存のworktreeには適用されない（新規worktree作成時のみ自動リンク）

##### `gw ln ls`

共有ファイルの一覧を表示する。

###### 動作

1. `~/.worktrees/{repo}/.gw-links/` 内のファイル・ディレクトリを再帰的にリスト
2. 相対パス形式で表示（例: `.env`、`tmp/cache/data.json`）

###### 例

```bash
$ gw ln ls
.env
tmp/cache
node_modules
```

##### `gw ln rm`

共有ファイルの登録を解除する。

###### 動作

1. `.gw-links/` 内のファイル一覧を表示
2. インタラクティブな選択UI（peco/fzf）で削除対象を選択
3. 選択されたファイル/ディレクトリをメインworktreeに移動
   - メインworktree = `git worktree list` で最初に表示されるworktree
   - メインworktreeのシンボリックリンクを実体ファイルで置き換え
4. `.gw-links/` から削除

###### 注意

- 他のworktreeのシンボリックリンクは壊れる（リンク切れになる）
- 各worktreeからの削除は行わない

#### worktree作成時の動作

`gw add` または `gw pr checkout` でworktreeを作成する際：

1. `.gw-links/` 内の各ファイル/ディレクトリに対してシンボリックリンクを作成
2. リンク対象のパスに既にファイルが存在する場合は警告を出してスキップ
   - worktree作成自体は続行する

#### ディレクトリ構成

```
~/.worktrees/
└── {repo}/
    ├── .gw-links/              # 共有ファイルの実体
    │   ├── .env
    │   ├── tmp/
    │   │   └── cache/
    │   └── node_modules/
    ├── user-2025-11-30-feature-a/   # worktree
    │   ├── .env -> ~/.worktrees/{repo}/.gw-links/.env
    │   ├── tmp/
    │   │   └── cache -> ~/.worktrees/{repo}/.gw-links/tmp/cache
    │   └── ...
    └── user-2025-11-30-feature-b/   # worktree
        ├── .env -> ~/.worktrees/{repo}/.gw-links/.env
        └── ...
