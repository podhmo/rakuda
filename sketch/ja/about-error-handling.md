# 雑にエラーハンドリングなどについて考える

ハンドラーの定義をそのまま使うとだるいかも？

liftみたいな関数が欲しい

- https://github.com/podhmo/quickapi/blob/main/lift.go

あとerrorをそのまま渡したいのだけど500のときにはloggerにだけ詳細を出力して外部には単にinternal server errorとかにしたい

rfc-7876直書きは辛い

-   https://github.com/podhmo/rakuda/pull/15#issue-3603403730

真面目に作ったapi errorを都度ハンドリングしたくない

- https://leapcell.io/blog/building-a-robust-error-handling-system-for-go-apis
- https://techblog.hacomono.jp/entry/2025/08/19/110000

せっかくだしencoding/json/v2専用にする？
