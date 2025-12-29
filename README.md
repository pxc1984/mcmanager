# Minecraft Plugins Manager

Small HTTP service that watches a plugins repo and updates your Minecraft server data directory. On each webhook hit it pulls the repo, copies `plugins/` and `bedwars_worlds/` into the server data dir, then triggers an RCON restart countdown (60s notice, 10→1 second countdown, then restart).

## Requirements

- Docker / docker-compose
- RCON enabled on your Minecraft server
- A Git repo containing `plugins/` and `bedwars_worlds/` directories

## Configuration

Create a `.env` file (see `.env.example`) with:

- `REPO_URL` (required): Git URL to clone.
- `REPO_BRANCH` (default `main`): Branch to pull.
- `REPO_PATH` (default `/tmp/plugin-repo`): Clone location inside the container.
- `DATA_DIR` (default `./data`): Path to the Minecraft data dir inside the container. For Docker, mount your server data volume here.
- `PORT` (default `8080`): HTTP listen port.
- `RCON_HOST`, `RCON_PORT`, `RCON_PASSWORD` (all required): RCON connection to the running server.
- `RCON_RESTART_COMMAND` (default `restart`): Command sent after countdown.
- `PLUGINS_UID` (optional): Numeric group ID to apply (recursive chown) to synced `plugins/` and `bedwars_worlds/` on Linux.

## Running with docker-compose

1. Copy `.env.example` to `.env` and fill in values.
2. Adjust `docker-compose.yml` if needed (e.g., swap the sample `minecraft` image for your own).
3. Start the stack:

   ```sh
   docker-compose up --build
   ```

   The manager listens on `PORT` (default 8080).

## Webhook usage

- Endpoint: `POST /update`
- Health: `GET /healthz`
- Expected body: none (the manager just pulls the configured repo).
- On `/update`, the service:
  1. Pulls the repo/branch (clones if missing).
  2. Copies `plugins/` → `${DATA_DIR}/plugins` and `bedwars_worlds/` → `${DATA_DIR}/bedwars_worlds` (destinations are fully replaced).
  3. Announces `Restarting in 60 seconds` via RCON, waits ~50s, counts down 10…1, then sends `RCON_RESTART_COMMAND`.

## Directory layout in the repo

```
plugins/
bedwars_worlds/
```

Anything else is ignored by the sync logic.

## Local development (optional)

```sh
go run ./cmd/manager
```

Use a `.env` or export variables before running.

## Notes and safety

- Destination directories are deleted/recreated on each sync to ensure a clean copy.
- RCON credentials are loaded from env; avoid checking them into git.
- If multiple webhooks hit quickly, requests are serialized to avoid overlapping syncs/restarts.
