# 悩ましい

名前の管理（パッケージ名の管理）
- rakudamiddlewareは参照実装なのでrakuda-prefixになっている
  - 実際のアプリを作る作者のために名前を空けておきたい
- bindingはそのまま
  - parserをデフォルトで作成したいがこれも参照実装？

Loggerの管理
- RespoderはDefaultLoggerをもたなくてよいのではないか？
- 先頭のミドルウェアでLoggerを設定したいかも？
  - Recoveryのミドルウェアでもログに依存してる
  - ginのDefaultを真似したいが…
    - Buildのタイミングで埋め込めば良いかも？
    - levelの調整ができてない

エラーログ
- 現状のsourceをpcのもので上書きするものはAPIErrorが必須になる
  - unexpected error（500）ではどこのエラーか判別できない
    - RespoderのErrorを陽に使えば指定可能。liftの時が無理。
  - stack traceが欲しいのでは無いか？

エラーレスポンス
-  rfc-7876直書きは辛い
    - https://github.com/podhmo/rakuda/pull/15#issue-3603403730
- 
真面目に作ったapi errorを都度ハンドリングしたくない
  - https://leapcell.io/blog/building-a-robust-error-handling-system-for-go-apis
  - https://techblog.hacomono.jp/entry/2025/08/19/110000

SSEが機能するか試せてない
- httplogでResponseWriterをwrapしてしまうとFlusherを取得しできないのでは？

Markdown出力が欲しい
- それっぽいhtmlが手に入るもの
- .mdをつけるとダウンロードできる
- https://github.com/yusukebe/gh-markdown-preview

まだ楽（raku）になってないかも?
- init的なscaffoldが欲しい？（？）
- encoding/json/v2を利用するようにしたい
- 機能の拡張はnet/httpだけの知識で済ませるようにしたい
- 細々とした調整ができてない
  - graceful shutdown
  - json dump時のオプションの設定（responder?）
- 手軽なdeploy
- web uiを付けるのに楽をしたい（ローカルで利用）
- bindingで利用するparserの自作は面倒なのでは？
- OpenAPI,,,
  - symgoを使ってコードを読む案 https://github.com/podhmo/go-scan/tree/main/examples/docgen
