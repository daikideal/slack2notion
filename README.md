# slack2notion

Slack の reaction_added イベントを受信し、対象メッセージを Notion DB に追加する Go アプリです。

## 必要な環境変数

このアプリを実行・デプロイする際は、以下の環境変数を必ずセットしてください。

- `SLACK_BOT_TOKEN` : Slack Bot のトークン
- `NOTION_API_TOKEN` : Notion Integration の API トークン
- `NOTION_DB_ID` : 追加先 Notion データベースの ID

`.env`ファイル例:

```
SLACK_BOT_TOKEN=xxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
NOTION_API_TOKEN=secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
NOTION_DB_ID=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## デプロイについて

このアプリは Cloud Run Functions や Vercel Functions など、さまざまな PaaS の関数型環境にもデプロイ可能です。

- 環境変数の設定方法は各 PaaS のドキュメントを参照してください。
- HTTP サーバーは`/slack/events`エンドポイントでリクエストを受け付けます。

## ローカル実行

```
export SLACK_BOT_TOKEN=...
export NOTION_API_TOKEN=...
export NOTION_DB_ID=...
go run main.go
```

または、`.env`ファイルを使う場合は`direnv`や`dotenv`などのツールを利用してください。
