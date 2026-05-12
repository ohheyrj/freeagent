# freeagent

A command-line interface for the [FreeAgent](https://www.freeagent.com) API.
Built around the resources I use day-to-day — listing projects, tasks and
users, plus full CRUD for timeslips. Not affiliated with FreeAgent.

## Install

```bash
git clone https://github.com/ohheyrj/go-freeagent.git
cd go-freeagent
go install
```

The binary lands in `$(go env GOPATH)/bin`. Make sure that's on your `PATH`.

## One-time setup

1. **Register an OAuth app** in the FreeAgent dev dashboard
   ([production](https://dev.freeagent.com) or
   [sandbox](https://api.sandbox.freeagent.com)) and add
   `http://localhost:8080/callback` as a redirect URI.

2. **Export your credentials.** The CLI reads these from the environment:

   ```bash
   export FREEAGENT_CLIENT_ID="your-client-id"
   export FREEAGENT_CLIENT_SECRET="your-client-secret"
   ```

   Optionally point at sandbox (defaults to production):

   ```bash
   export FREEAGENT_BASE_URL="https://api.sandbox.freeagent.com/v2"
   ```

   `direnv` works well for per-directory environments.

3. **Authenticate.** One command opens your browser, runs the OAuth flow, and
   caches the tokens locally:

   ```bash
   freeagent auth login
   ```

   Tokens are cached per environment in your OS cache dir (`~/Library/Caches/freeagent/`
   on macOS, `~/.cache/freeagent/` on Linux). You won't need to do this again
   until the refresh token actually expires.

## Commands

```
freeagent <resource> <verb> [flags]
```

Every list command supports `--json` for machine-readable output.

### Projects

```bash
freeagent projects list
freeagent projects list --view=active   # active | completed | cancelled | hidden
freeagent projects list --json | jq '.[].name'
```

### Tasks

```bash
freeagent tasks list
freeagent tasks list --project=42       # filter to a specific project
```

### Users

```bash
freeagent users list
```

### Timeslips

```bash
# Read
freeagent timeslips list
freeagent timeslips list --from=2026-01-01 --to=2026-01-31
freeagent timeslips list --user=12 --project=42

# Write
freeagent timeslips add --user=12 --project=42 --task=3 --hours=2
freeagent timeslips add --user=12 --project=42 --task=3 --hours=1.5 \
                       --date=2026-05-09 --comment="planning meeting"

# Update
freeagent timeslips update 12345 --hours=3
freeagent timeslips update 12345 --comment="moved to next week" --date=2026-05-12

# Delete (prompts for confirmation)
freeagent timeslips delete 12345
freeagent timeslips delete 12345 --yes   # skip prompt, for scripts
```

### Shell completion

Cobra generates completion scripts for all commands and flags:

```bash
freeagent completion zsh > ~/.zsh/completions/_freeagent
freeagent completion bash > /etc/bash_completion.d/freeagent
```

## Environment variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `FREEAGENT_CLIENT_ID` | yes | OAuth app client ID |
| `FREEAGENT_CLIENT_SECRET` | yes | OAuth app client secret |
| `FREEAGENT_BASE_URL` | no | API base URL (defaults to production) |
| `FREEAGENT_ACCESS_TOKEN` | no | Seed access token (cache takes precedence) |
| `FREEAGENT_REFRESH_TOKEN` | no | Seed refresh token (cache takes precedence) |

`FREEAGENT_ACCESS_TOKEN` and `FREEAGENT_REFRESH_TOKEN` are only used as a seed
on first run — after `auth login`, the cache supersedes them.

## Notes

- **Pagination** is automatic. List commands fetch every page transparently using
  `X-Total-Count` from the API; you don't need to ask.
- **Token rotation** is handled automatically. FreeAgent rotates refresh tokens;
  the cache keeps up so you don't lose access silently.
- **Environment safety.** When `FREEAGENT_BASE_URL` is set to anything other
  than production, the CLI prints `Using FreeAgent API: <url>` to stderr on every
  command. Production stays silent.
- **Resetting auth.** Delete the cache file (e.g. `rm ~/Library/Caches/freeagent/tokens-*.json`)
  and run `freeagent auth login` again.
