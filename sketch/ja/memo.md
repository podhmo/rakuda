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

エラーレスポンス
-  rfc-7876直書きは辛い
    - https://github.com/podhmo/rakuda/pull/15#issue-3603403730
- 
真面目に作ったapi errorを都度ハンドリングしたくない
  - https://leapcell.io/blog/building-a-robust-error-handling-system-for-go-apis
  - https://techblog.hacomono.jp/entry/2025/08/19/110000

SSEが機能するか試せてない
- httplogでResponseWriterをwrapしてしまうとFlusherを取得しできないのでは？

まだ楽（raku）になってないかも?
- init的なscaffoldが欲しい？（？）
- encoding/json/v2を利用するようにしたい
- 機能の拡張はnet/httpだけの知識で済ませるようにしたい
- 細々とした調整ができてない
  - graceful shutdown
  - json dump時のオプションの設定（responder?）
- 手軽なdeploy
- web uiを付けるのに楽をしたい（ローカルで利用）
