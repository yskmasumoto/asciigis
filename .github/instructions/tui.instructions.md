---
applyTo: "**/internal/tui/**/*.go"
---

# tui.instruction.md

このファイルは、Go言語のフレームワーク [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) を使用して、モダンでリッチな TUI アプリケーションを構築する際のガイドラインとベストプラクティスを定義します。

AI はコードを生成する際、以下の原則とパターンに従ってください。

## 1. 基本アーキテクチャ (The Elm Architecture)

Bubble Tea は Elm アーキテクチャに基づいています。すべてのアプリケーションは以下の3つの要素で構成されなければなりません。

* **Model**: アプリケーションの状態を保持する構造体。
* **Update**: メッセージ (`tea.Msg`) を受け取り、新しいモデルとコマンド (`tea.Cmd`) を返す関数。
* **View**: モデルの状態に基づいて UI 文字列をレンダリングする関数。

### コーディング規約
* 状態管理は単一の `model` 構造体で行うか、複雑な場合はサブモデル（`bubbles` コンポーネントなど）を埋め込むこと。
* `Init` 関数は初期コマンド（タイマーの開始、初期データのロードなど）を返すために使用する。何もしない場合は `nil` を返す。

## 2. スタイリングとレイアウト (Lip Gloss の活用)

おしゃれな UI を作るために、必ず [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) を使用してください。生の ANSI エスケープシーケンスは避け、スタイル定義を使用します。

### スタイルの定義
* スタイルはパッケージレベルの変数、または専用のスタイル構造体として定義する。
* **色**: `lipgloss.Color` またはライト/ダークモード対応の `lipgloss.AdaptiveColor` を使用する。
* **ボーダー**: モダンな見た目のために `lipgloss.RoundedBorder()` を積極的に採用する。
* **パディング/マージン**: 余白を適切に設定し、窮屈な UI を避ける。

```go
var (
    subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
    titleStyle  = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFFDF5")).
        Background(lipgloss.Color("#25A065")).
        Padding(0, 1)
)
```
