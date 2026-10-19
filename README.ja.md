# kojo-ext-notion

Notion 公式のローカル MCP サーバー [`@notionhq/notion-mcp-server`](https://github.com/makenotion/notion-mcp-server) を、internal integration token 付きで kojo の agent に渡す拡張。

中身は小さな launcher。kojo がこれを stdio MCP サーバー `notion` として起動すると、kojo の extension secrets API から agent の token を取得し、`NOTION_TOKEN` を設定して `npx -y @notionhq/notion-mcp-server@2.5.2` を実行する。token は backend CLI の MCP 設定やコマンドラインに出ない。kojo の拡張用認証情報 (`KOJO_EXT_*`) は Notion サーバーに渡さない。

## 必要なもの

- Node.js 18+ と `npx` (`PATH`、よくあるインストール先、または設定 `npxPath` から探す)。初回起動時に npm からサーバーを取得する。

## 設定

1. <https://www.notion.so/profile/integrations> で internal integration を作って secret (`ntn_…`) をコピー。agent に与えたい capability (読み取り / 更新 / 挿入) を選ぶ。
2. agent に使わせるページやデータベースを integration に共有する (ページの `···` → Connections)。
3. kojo にこの package を入れて agent で有効にし、**Internal integration token** に secret を貼る。

公式サーバーのツール (検索、ページ、ブロック、データベース / データソース、コメント、ユーザー) がそのまま使える。触れる範囲は integration の capability と共有したページで決まる。
