# kojo-ext-notion

Gives a kojo agent Notion's official local MCP server, [`@notionhq/notion-mcp-server`](https://github.com/makenotion/notion-mcp-server), authenticated with a Notion internal integration token.

The package ships a small launcher. kojo starts it as the stdio MCP server `notion`; it reads the agent's token from kojo's extension secrets API and runs `npx -y @notionhq/notion-mcp-server@2.5.2` with `NOTION_TOKEN` set. The token never lands in the backend CLI's MCP config or on a command line, and kojo's own extension credentials (`KOJO_EXT_*`) are not passed to the Notion server.

## Requirements

- Node.js 18+ with `npx` (found in `PATH`, common install locations, or the `npxPath` setting). The first start downloads the server from npm.

## Setup

1. Create an internal integration at <https://www.notion.so/profile/integrations> and copy its secret (`ntn_…`). Pick the capabilities (read / update / insert content) you want the agent to have.
2. Share the pages and databases the agent may use with the integration (page `···` → Connections).
3. Install this package in kojo, enable it for the agent and paste the secret into **Internal integration token**.

The agent gets every tool of the official server (search, pages, blocks, databases / data sources, comments, users). What it can actually touch is limited by the integration's capabilities and the pages shared with it.
