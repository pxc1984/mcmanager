# Migration: 625d5b3fc98cfff1df7d745be2f453f69da2729a -> current

This document explains how to migrate a deployment configured like commit
`625d5b3fc98cfff1df7d745be2f453f69da2729a` to the current version.

## Summary of changes

- New env vars:
  - `PLUGINS_DOWNLOAD` (default `true`): runs `plugins/download.sh` before syncing.
  - `SECRET_TOKEN` (optional): if set, `/update` requires `X-Secret-Token` header.
  - `COPY_DIRS` (default `plugins,bedwars_worlds,configs,worlds`): directories to sync.
  - `SKIP_DIRS` (optional): directories to exclude from `COPY_DIRS`.
- Sync now copies all directories from `COPY_DIRS` (minus `SKIP_DIRS`), not only
  `plugins/` and `bedwars_worlds/`.
- `PLUGINS_UID` now defaults to `0` (disabled). Non-zero applies recursive chown
  on Linux only.

## Example: old env -> new env

Old env example (from 625d5b3):

```
REPO_URL=https://github.com/pxc1984/mc-cu-server-configs.git
REPO_BRANCH=main
REPO_PATH=/tmp/plugin-repo
DATA_DIR=/data
PORT=8080

RCON_HOST=minecraft
RCON_PORT=25575
RCON_PASSWORD=changeme
RCON_RESTART_COMMAND=restart

PLUGINS_UID=1000
LOCALE=ru
```

New env (current) with the same intent:

```
REPO_URL=https://github.com/pxc1984/mc-cu-server-configs.git
REPO_BRANCH=main
REPO_PATH=/tmp/plugin-repo
DATA_DIR=/data
PORT=8080

RCON_HOST=minecraft
RCON_PORT=25575
RCON_PASSWORD=changeme
RCON_RESTART_COMMAND=restart

PLUGINS_UID=1000
LOCALE=ru

# New in current
PLUGINS_DOWNLOAD=true
SECRET_TOKEN=
COPY_DIRS=plugins,bedwars_worlds,configs,worlds
SKIP_DIRS=
```

## If you want to preserve old behavior exactly

Only sync `plugins/` and `bedwars_worlds/`:

```
COPY_DIRS=plugins,bedwars_worlds
```

Disable plugin download script:

```
PLUGINS_DOWNLOAD=false
```

Lock down `/update` with a shared secret:

```
SECRET_TOKEN=your-long-random-token
```

Then send header `X-Secret-Token: your-long-random-token` on `POST /update`.

## Notes

- `plugins/download.sh` must exist if `PLUGINS_DOWNLOAD=true`.
- If all directories from `COPY_DIRS` are missing, the sync will fail.
