# Portainer CLI (`portainer-cli`)

A Go-based command line interface and SDK client for managing Portainer instances, based on the [Portainer CE/EE 2.39.1 API specification](https://api-docs.portainer.io/?edition=ce&version=2.39.1#tag/stacks).

## Features

- **Stack Module**:
  - `list` / `ls`: List stacks across all environments with optional endpoint or swarm filtering.
  - `inspect` / `get`: View detailed configuration, environment variables, git config, and status of a stack.
  - `file` / `cat`: View or export a stack's docker-compose content.
  - `start`: Start a stopped stack.
  - `stop`: Stop an active stack.
  - `restart`: Restart a stack (shorthand that executes stop & start sequentially).
  - `deploy`: Deploy a new standalone compose or swarm stack from file or stdin.
  - `deploy-git`: Deploy a stack from a remote Git repository (GitOps).
  - `update`: Update stack docker-compose file content or environment variables.
  - `redeploy`: Trigger a git pull and redeploy for git-backed stacks.
  - `migrate`: Migrate a stack to another Portainer environment.
  - `delete` / `rm`: Remove a stack.
- **Authentication & Config Management**:
  - `login`: Authenticate and save credentials permanently to `~/.config/portainer-cli/config.json`.
  - `logout`: Clear saved credentials.
  - `config show / set`: Inspect active configuration or update settings (URL, API key, endpoint, format).
- **Multiple Output Formats**: Table (human-readable), JSON, and YAML (`-o json`, `-o yaml`, `-o table`).
- **Flexible Authentication**: Automatically prioritizes CLI flags > Environment Variables > Local `.env` > Persistent `config.json`.
- **SDK & CLI**: `pkg/client` and `pkg/stacks` can be imported as a Go library.

---

## Quick Start: One-time Authentication (`login`)

You only need to authenticate once:
```bash
portainer-cli login --url https://localhost:9443 --api-key "ptr_..."
```
This validates your key with Portainer and writes to `~/.config/portainer-cli/config.json`. Afterwards, you can run any `portainer-cli` command from any directory without passing `--api-key` or having a `.env` file!

To update or view configuration:
```bash
# View active config
portainer-cli config show

# Update API key or URL anytime
portainer-cli config set api-key "new-api-key"
portainer-cli config set url "https://portainer.example.com"
```

---

## Installation & Build

### Using Go
```bash
cd portainer-cli
make build
make install # installs to ~/.local/bin/portainer-cli
```

### Using Nix
```bash
nix develop
make build
```

---

## Configuration

`portainer-cli` looks for credentials in:
1. CLI flags (`--url`, `--api-key`, `--insecure`, `--endpoint-id`)
2. Environment variables:
   - `PORTAINER_URL` (default: `https://localhost:9443`)
   - `PORTAINER_API_KEY` or `X_API_KEY`
   - `PORTAINER_JWT`
   - `PORTAINER_INSECURE` (default: `true`)
   - `PORTAINER_ENDPOINT_ID`
   - `PORTAINER_OUTPUT` (default: `table`)
3. `.env` file in current, parent, or `~/.config/portainer-cli/` directory.

Example `.env`:
```env
PORTAINER_URL=https://localhost:9443
PORTAINER_API_KEY=ptr_EVpCkqjJO3i4jWEYdaY/7jlpH+nUVV5SVKp3OXGKuXU=
PORTAINER_INSECURE=true
PORTAINER_ENDPOINT_ID=2
```

---

## Usage Examples

### 1. List stacks
```bash
# List all stacks in table format
portainer-cli stack list

# Filter by environment ID
portainer-cli stack list --endpoint 2

# Output as JSON or YAML
portainer-cli stack list -o json
portainer-cli stack list -o yaml
```

### 2. Inspect a stack
```bash
portainer-cli stack inspect 7
portainer-cli stack inspect 7 -o yaml
```

### 3. Retrieve stack compose file
```bash
# Print to stdout
portainer-cli stack file 7

# Save to file
portainer-cli stack file 7 -w docker-compose.yml
```

### 4. Start, Stop & Restart stacks
```bash
portainer-cli stack stop 7
portainer-cli stack start 7
portainer-cli stack restart 7
```

### 5. Deploy a stack
```bash
# From a local docker-compose file
portainer-cli stack deploy --name my-app --file docker-compose.yml --endpoint 2

# With custom environment variables
portainer-cli stack deploy --name my-app --file docker-compose.yml --endpoint 2 --env PORT=8080 --env APP_ENV=production

# From stdin
cat docker-compose.yml | portainer-cli stack deploy --name my-app --file - --endpoint 2
```

### 6. Deploy from Git (GitOps)
```bash
portainer-cli stack deploy-git \
  --name my-git-app \
  --repo https://github.com/my-org/my-repo \
  --ref refs/heads/main \
  --compose-file docker-compose.yml \
  --endpoint 2
```

### 7. Update a stack
```bash
portainer-cli stack update 7 --file docker-compose.updated.yml --pull-image
```

### 8. Redeploy a git-backed stack
```bash
portainer-cli stack redeploy 7 --pull-image
```

### 9. Delete a stack
```bash
portainer-cli stack delete 7 --endpoint 2
```
