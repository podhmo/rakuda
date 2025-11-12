# 悩ましい

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

SSEが機能するか試せてない
- httplogでResponseWriterをwrapしてしまうとFlusherを取得しできないのでは？

まだ楽になってないかも?
- init的なscaffoldが欲しい？（？）
- encoding/json/v2を利用するようにしたい
- 機能の拡張はnet/httpだけの知識で済ませるようにしたい
- 細々とした調整ができてない
  - graceful shutdown
  - json dump時のオプションの設定（responder?）
- 手軽なdeploy
- web uiを付けるのに楽をしたい（ローカルで利用）
