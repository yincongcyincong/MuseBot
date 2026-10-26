## Configuring the `MCP_CONF_PATH` Environment Variable for MCP Servers (Go Binary)

This document guides you on configuring the **`MCP_CONF_PATH`** environment variable when using the Go binary of the **`MuseBot`** project. This allows you to use a custom MCP server configuration.

### 1. Create the MCP Configuration File

First, you'll need to create a JSON file that defines your MCP servers. Here's an example configuration, including settings for GitHub, Playwright, Amap (amap-mcp-server), and Amap Maps (amap-maps) MCP servers. You can modify this file to suit your specific needs.

```json
{
    "mcpServers": {
       "github": {
          "command": "docker",
          "args": [
             "run",
             "-i",
             "--rm",
             "-e",
             "GITHUB_PERSONAL_ACCESS_TOKEN",
             "ghcr.io/github/github-mcp-server"
          ],
          "env": {
             "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>"
          }
       },
       "playwright": {
          "url": "http://localhost:8931/sse"
       },
       "amap-mcp-server": {
          "url": "http://localhost:8000/mcp"
       },
       "amap-maps": {
          "command": "npx",
          "args": [
             "-y",
             "@amap/amap-maps-mcp-server"
          ],
          "env": {
             "AMAP_MAPS_API_KEY": "<YOUR_TOKEN>"
          }
       }
    }
}
```

**Please note:**

* Replace **<YOUR\_TOKEN>** with your actual GitHub personal access token and Amap API key.
* The **`url`** fields for **`amap-mcp-server`** and **`playwright`** MCP servers point to services running locally. Ensure these services are running and accessible.
* You can add or remove MCP server configurations as needed.

Save this file to a directory of your choice. For instance, you might name it **`mcp_config.json`** and place it in the project's root directory.

### 2. Set the `MCP_CONF_PATH` Environment Variable

For Go binaries, setting the **`MCP_CONF_PATH`** environment variable is straightforward. The main goal is to ensure the environment variable is correctly set *before* running the binary.

#### Method 1: Set in the Command Line (Temporary)

If you just need to run the binary temporarily for testing, you can set the environment variable directly in your command line before execution:

**Linux/macOS:**

```bash
export MCP_CONF_PATH=/path/to/your/mcp_config.json
./MuseBot # Assuming this is your Go binary
```

#### Method 2: Set in Docker Compose or Dockerfile (If Using Docker)

If you're deploying **`MuseBot`** using Docker, you can set the environment variable in `Dockerfile`.


**Setting in `Dockerfile`:**

```dockerfile
# ... other Dockerfile instructions ...

COPY mcp_config.json /app/mcp_config.json

ENV MCP_CONF_PATH /app/mcp_config.json

# ... other Dockerfile instructions ...
```

### 3. Run the `MuseBot` Go Binary

After setting the **`MCP_CONF_PATH`** environment variable, you can run the **`MuseBot`** Go binary as usual. The project will load your specified MCP configuration file and will be able to use the MCP servers defined within it.

For example:

```bash
./MuseBot
```

Your **`MuseBot`** should now be able to interact with your configured MCP servers.

---

### 4. Example: a keyless remote MCP server (deed.page)

Remote servers only need a `url`. MuseBot detects Streamable HTTP or SSE from the
endpoint. For example, [deed.page](https://deed.page) is a free static hosting
service with a keyless MCP endpoint. With it, `/task` or `/mcp` runs can publish
an HTML page and reply with its URL:

```json
{
    "mcpServers": {
       "deed-page": {
          "url": "https://deed.page/mcp"
       }
    }
}
```

You can also add it at runtime through the MCP management in `/task`, for example
"add an MCP server named deed-page with url https://deed.page/mcp".

**Please note:**

* **Published pages are public.** Anyone with the URL can open them, so don't ask the bot to publish secrets or private chat content.
* MuseBot exposes every tool the server lists (about two dozen, including `publish_site`, `get_site` and `delete_site`). Remove the entry or set `"disabled": true` when you don't need it.
* `publish_site` returns the page URL and a management token, which shows up in the tool result. Keep the token if you want to update or delete the page. Ask for `ttl_seconds` to make a page expire on its own.
* See the [deed.page docs](https://deed.page/docs) for limits and the full tool list.
