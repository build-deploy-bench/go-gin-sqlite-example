# Build-deploy benchmark fixture: Go (Gin + GORM + SQLite)

## Source

- Upstream: https://github.com/realworld-apps/golang-gin-realworld-example-app at `a41289b2a4951e7cc09066f86742a27f85121c9b` (2026-11-02). MIT license, kept in `LICENSE`.
- This repo: https://github.com/build-deploy-bench/go-gin-sqlite-example (private, independent copy). The first commit is the unmodified upstream.
- Smoke-tested commit: `4690fcacf70c7abce218685a6e578786b106033f`.

## Patches

1. `gorm.io/driver/sqlite` (cgo) replaced with `github.com/glebarez/sqlite` v1.11.0 (pure Go, same GORM API). One test regex was loosened because the new driver words a SQLite error differently.
2. Go pinned to 1.27.3 in `go.mod` and in the Docker `golang:1.27.3-bookworm` tag. Render's native Go runtime cannot be pinned: it uses its latest stable release with `GOTOOLCHAIN=local`. A `go.mod` version newer than Render's fails the build, so the pin follows Render. When Render moves to a newer patch, Render builds with it and the other providers stay on 1.27.3. Check `go version` in the install step output.
3. Default `DB_PATH` is `/tmp/gorm.db`. The database is created and migrated at startup on every provider; there is no external database. Data is lost on restart.
4. `GET /version` returns `{"commit":"<sha>","run":"<id>"}`. `run` is the runtime env var `BENCH_RUN_ID` (else `unknown`). The harness sets it before each deploy, so it identifies the deployment even when the commit is the same. `bench/build.sh` embeds the SHA with `-ldflags -X main.commit=`. It uses the first value it finds in `BENCH_COMMIT`, `RENDER_GIT_COMMIT`, `RAILWAY_GIT_COMMIT_SHA` or `VERCEL_GIT_COMMIT_SHA`, then `git rev-parse HEAD`, and then `unknown`.
5. Upstream GitHub Actions workflows removed (a scheduled job ran every 6 hours).
6. Added `bench/build.sh`, `Dockerfile.fly`, `Dockerfile.vercel` (identical below the header comment), `fly.toml` and `vercel.json`. There is no root `Dockerfile`.

The app already binds all interfaces (`r.Run(":" + port)`) and reads `PORT`. The JWT secret is a hard-coded constant upstream, so no secrets are set on any provider.

## Commands

| Step | Command | Markers |
|---|---|---|
| Install | `./bench/build.sh install` = `go version && go mod download` | `::bench::install_start` / `::bench::install_done` |
| Build | `./bench/build.sh build` = `CGO_ENABLED=0 go build -ldflags "-X main.commit=<sha>" -o app .` | `::bench::build_start` / `::bench::build_done` |
| Start | `./app` | |
| Health | `GET /api/tags` returns 200 `{"tags":[]}` | |

Go has no separate install step. `go mod download` is the install, and compile plus link is the build. `go mod download` downloads all modules in `go.sum`, test-only ones too.

Env: `GIN_MODE=release` on every provider (`[env]` in `fly.toml`, `ENV` in the Dockerfiles). Docker images listen on `PORT=80`.

Docker image: build stage `golang:1.27.3-bookworm`, runtime `gcr.io/distroless/static-debian12` (a moving tag, not pinned by digest). About 39 MB. The binary is static.

## Providers

### Render

- Workspace `tea-dbmbhbg8v5cc73e4m0jg` (workspace plan: Pro).
- Service `srv-dbpn9he1egvs738u31fg` `bdb-go-gin-sqlite`, https://bdb-go-gin-sqlite.onrender.com
- Native runtime `go`, region `virginia`. Instance plan `starter` (paid, always on) and build pipeline `starter`, the same as the Node service. Recreated on 2026-11-10 because the API cannot change a free service to a paid plan; setup deploy `dep-dbpn9hu1egvs738u32cg`. Auto-deploy `no`. Not in a Render project or environment.
- Build command `./bench/build.sh all` (Render has one build-command field). Start command `./app`. Health check path `/api/tags`.
- Go: Render's latest stable (1.27.3 on 2026-11-10); see patch 2.
- Deploys: `dep-dbpmohgu01pc73dr4mkg` (the creation deploy of `50e121b`) failed because `go.mod` asked for 1.27.4. Smoke deploy `dep-dbpmr26tlv9c73dplui0` of `4690fca` is live: health 200, and `/version` returns the right SHA.

### Railway

- Workspace `d535606c-49a7-4a6c-9e29-ce73b5f01428` (paid plan).
- New project `bdb-go-gin-sqlite` `cdd69b1d-d658-4df0-84e1-6b5348ccb42f`, environment `production` `b77edf7d-b9e0-4051-aeff-46f1619ea7eb`.
- Service `bdb-go-gin-sqlite` `ec91dcc2-4f92-43df-8fde-83c270f4b814`, source `build-deploy-bench/go-gin-sqlite-example` branch `main`. https://bdb-go-gin-sqlite-production.up.railway.app
- Builder Railpack (0.43.0; it resolves Go from `go.mod` through mise). One replica in `iad`, the default region (the Node service also leaves the region unset).
- Variables: `GIN_MODE=release`, `RAILPACK_INSTALL_CMD=./bench/build.sh install`. Build command `./bench/build.sh build`, start command `./app`, health check `/api/tags`, all set with `serviceInstanceUpdate`.
- Auto-deploy is disabled and confirmed with `serviceInstanceAutoDeployStatus`. **`serviceConnect` turns auto-deploy back on**, so disable it after you connect the repo. Because of this, the push of `4690fca` started deploy `551cca5a-3d7d-4ae9-8fc8-97b596b2ddf8`, which was replaced.
- Deploys: `536a36eb-caf9-4513-a262-78ee930a4f50` (the connect deploy of `50e121b`, which succeeded and showed all four markers). Smoke deploy `2eecabe8-a17b-4fe9-bf52-4276909aedfc` of `4690fca` (`railway redeploy --commit`) succeeded: health 200, and `/version` returns the right SHA. `RAILWAY_GIT_COMMIT_SHA` is available at build time.

### Vercel

- Team `build-deploy-bench` `team_oVjpdXNb0OhSfAzc6FSNjen8` (paid plan).
- Project `bdb-go-gin-sqlite` `prj_zbY0uRpBCe3Se5YFg1jLFHp7D1jU`, https://bdb-go-gin-sqlite.vercel.app. The project has no Git link, the same as the Node project; deploys upload a clean context directory with `vercel deploy --prod`.
- Container Runtime (`vercel.json` `framework: container`, `Dockerfile.vercel`). Function region `iad1`. Build machine `basic`, `fixed`. Deployment protection: the default `all_except_custom_domains`.
- Deploys of `4690fca`: `dpl_Df8gjGCnFFB935QzXDkG67DKxNR6` (first) and `dpl_31Y5wdLk9h9NLHXzPYsS8ptfh87R`. The second one was an accidental repeat. The production alias returns health 200.
- **`/version` returns `unknown` on Vercel.** `vercel deploy --build-env BENCH_COMMIT=…` does not reach the Dockerfile as a build arg, and a CLI upload has no `VERCEL_GIT_COMMIT_SHA`. Follow-up: deploy from Git (`gitSource`, which sets `VERCEL_GIT_COMMIT_SHA`), or set the project env var `BENCH_COMMIT` before each deploy. The probe now identifies a deployment by `run` (`BENCH_RUN_ID`, set per deployment with `vercel deploy --env BENCH_RUN_ID=<id>`), so the commit is no longer needed for that.
- The Buildah store cache is now on by default (builder CLI 65.11.0 logs `Restoring the Buildah store cache`). The first build was cold (`miss (no_compatible_snapshot)`), and its image push took 344 s. The second build hit the cache (545 MiB) and pushed in 0.4 s.

### Fly.io

- Org `personal` (paid plan). App `bdb-go-gin-sqlite`, https://bdb-go-gin-sqlite.fly.dev
- One Machine `8d99e47ae42d98`, `iad`, shared-cpu-1x with 512 MB (the same as the Node app), always on. The health check `/api/tags` passes.
- Deploy: `flyctl deploy --config fly.toml -a bdb-go-gin-sqlite --remote-only --depot=false --ha=false --yes --build-arg BENCH_COMMIT=<sha>` from a `git archive` of the commit. `/version` returns the right SHA.
- **Builder difference:** two attempts with the default Depot builder failed before the build started (`connecting to Depot builder: ... authentication handshake failed: EOF`). The smoke deploy used Fly's own remote builder (`--depot=false`, `fly-builder-rosy-headland-8386`). Decide which builder the harness uses and keep it the same for every run.

## Run ID check (2026-11-10)

`d0964a1` deployed with two different `BENCH_RUN_ID` values on each provider. `/version` returned each ID.

- Render: `PUT /v1/services/<id>/env-vars/BENCH_RUN_ID` does not start a deploy. The service uses cache profile `no-cache`, so the ID has no cache effect. Deploy `dep-dbpof5rm8hqs73crh320`. For about 20 s after `live`, the old instance still answered.
- Railway: `railway variable set BENCH_RUN_ID=<id> --skip-deploys`, then `railway redeploy --commit`. Railpack hashes all variables into its build secrets, so a new ID also reruns the mise Go install (about 3 s). The install and build steps run again on every deploy, also with the same ID. Deploys `bc6855b6-…`, `6582be91-…`, control `24e41071-…`.
- Vercel: `vercel deploy --prod --env BENCH_RUN_ID=<id>` (runtime env for that deployment only). A new ID kept the build cache and every Buildah layer (`dpl_F5Huph1AK1BPVt7tZqjFLZRgmzoS`).
- Fly: `flyctl deploy --env BENCH_RUN_ID=<id>`. A new ID kept every BuildKit layer.
