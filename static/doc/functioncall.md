## Configuring the `MCP_CONF_PATH` Environment Variable for MCP Servers (Go Binary)

This document guides you on configuring the **`MCP_CONF_PATH`** environment variable when using the Go binary of the **`telegram-deepseek-bot`** project. This allows you to use a custom MCP server configuration.

### 1. Create the MCP Configuration File

First, you'll need to create a JSON file that defines your MCP servers. Here's an example configuration, including settings for GitHub, Playwright, Amap (amap-mcp-server), and Amap Maps (amap-maps) MCP servers. You can modify this file to suit your specific needs.

```json
{
    "mcpServers": {
       "github": {
          "command": "docker",
          "description": "Performs Git operations and integrates with GitHub to manage repositories, pull requests, issues, and workflows.",
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
          "description": "Simulates browser behavior for tasks like web navigation, data scraping, and automated interactions with web pages.",
          "url": "http://localhost:8931/sse"
       },
       "amap-mcp-server": {
          "description": "Provides geographic services such as location lookup, route planning, and map navigation.",
          "url": "http://localhost:8000/mcp"
       },
       "amap-maps": {
          "command": "npx",
          "description": "Provides geographic services such as location lookup, route planning, and map navigation.",
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
./telegram-deepseek-bot # Assuming this is your Go binary
```

#### Method 2: Set in Docker Compose or Dockerfile (If Using Docker)

If you're deploying **`telegram-deepseek-bot`** using Docker, you can set the environment variable in `Dockerfile`.


**Setting in `Dockerfile`:**

```dockerfile
# ... other Dockerfile instructions ...

COPY mcp_config.json /app/mcp_config.json

ENV MCP_CONF_PATH /app/mcp_config.json

# ... other Dockerfile instructions ...
```

### 3. Run the `telegram-deepseek-bot` Go Binary

After setting the **`MCP_CONF_PATH`** environment variable, you can run the **`telegram-deepseek-bot`** Go binary as usual. The project will load your specified MCP configuration file and will be able to use the MCP servers defined within it.

For example:

```bash
./telegram-deepseek-bot
```

Your **`telegram-deepseek-bot`** should now be able to interact with your configured MCP servers.

---
