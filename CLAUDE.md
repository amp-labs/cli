# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`amp` — the Ampersand CLI (Go, Cobra-based). It talks to the Ampersand API to manage
integrations, installations, connections, destinations, and projects, and provides local
webhook development tooling. User-facing docs: https://docs.withampersand.com/cli/overview

## Build / lint / test

Building requires [Task](https://taskfile.dev). The build bakes stage-specific config
(API URL, Clerk URL, login URL, version metadata) into the binary via `-ldflags -X`, so
you build *per stage* rather than with a plain `go build`.

```sh
task build          # alias for build-dev → produces bin/amp
task build-local    # points API_URL at http://127.0.0.1:8080
task build-staging
task build-prod
```

For local dev the README suggests symlinking `bin/amp` to `lamp` ("local amp") so it
coexists with an installed production `amp`.

```sh
task lint                              # golangci-lint (config .golangci.yml)
task fix                               # wsl + gci + golangci-lint --fix (needs gci, wsl installed)
make fix-files FILES="cmd/a.go cmd/b.go"   # format/vet specific files only

go test ./...                          # run tests (only files/ has tests currently)
go test ./files -run TestGetRemovedReadObjects # run a single test
```

There is no `task test`; use `go test` directly. Linting is strict (`enable-all` minus a
curated disable list in `.golangci.yml`); the `openapi/` directory is excluded from lint.
Note revive is configured to accept `Id`/`Api` (not just `ID`/`API`).

## Code generation

`openapi/*.gen.go` are generated types (via `oapi-codegen`) pulled from remote specs in the
`amp-labs/openapi` repo — the manifest, problem, and catalog schemas. Regenerate with:

```sh
cd openapi && make gen
```

Do not hand-edit the `.gen.go` files. The `Manifest`, `Integration`, `CatalogType`, etc.
types used throughout the code live in this generated `openapi` package.

## Architecture

**Command wiring (Cobra + Viper).** `main.go` → `cmd.Execute()` → `rootCmd` (in
`cmd/root.go`). Each command lives in its own file under `cmd/` and self-registers via an
`init()` that calls `rootCmd.AddCommand(...)`. Global persistent flags (`--debug/-d`,
`--project/-p`, `--key/-k`) are set up in `flags/config.go` (`flags.Init`) and bound to
Viper; `--key` also reads the `AMP_API_KEY` env var. Use the `flags` package helpers
(`GetProjectOrFail`, `GetAPIKey`, `GetOutputFormat`, `GetDebugMode`) rather than reading
Viper directly.

**Stage configuration is build-time.** The `vars` package holds `ApiURL`, `ClerkRootURL`,
`LoginURL`, `Stage`, `Version`, etc., all defaulting to `"unset"` and overwritten at build
time by ldflags (see `Taskfile.yaml`). At runtime these can be overridden by env vars for
testing: `AMP_API_URL`, `AMP_CLERK_URL_OVERRIDE`, `AMP_STAGE_OVERRIDE`, `AMP_API_KEY`.

**Authentication (two modes).** `request.APIClient.getAuthHeader` decides:
1. If an API key is present (`--key` / `AMP_API_KEY`) → sends `X-Api-Key`.
2. Otherwise → uses a Clerk JWT browser-login session. `amp login` performs the Clerk OAuth
   flow and writes session data to the XDG config dir (`amp/jwt.json`, or
   `amp/jwt-<stage>.json` for non-prod). The `clerk` package exchanges that stored session
   for a fresh JWT on each request (`FetchJwt`).

**API layer (`request/`).** `request.go` is the low-level HTTP client (`Client`) that adds
`X-Amp-Client*` headers, marshals/unmarshals JSON, and parses error responses. Non-2xx
responses with `application/problem+json` are decoded into `ProblemError` (RFC 7807).
`api.go` defines `APIClient` — one method per API endpoint (`ListIntegrations`,
`BatchUpsertIntegrations`, `GetPreSignedUploadURL`, etc.), each building a URL under
`{ApiURL}/v1/projects/{projectId}/...` and attaching the auth header. Response DTOs are in
`request/types.go`.

**Deploy flow (`amp deploy`).** The most involved path, worth understanding end to end:
`files.Zip` locates and validates an `amp.yaml`/`amp.yml` manifest (parsed into the
generated `openapi.Manifest`, validated by `files/manifest.go` against `specVersion 1.0.0`),
zips it in memory → compute MD5 → `GetPreSignedUploadURL` → `storage.Upload` PUTs the zip to
GCS → `BatchUpsertIntegrations` is called with the resulting `gs://` URL. Before deploying,
`confirmReadObjectRemoval` compares the new manifest against existing integrations and, if
read objects were removed *and* there are live installations, interactively prompts (via
`promptui`) whether to pause reads (the `destructive` flag).

**Local webhook development.** `amp listen` (hidden command) runs a local HTTP server that
pretty-prints and forwards incoming webhooks to your app, writing its chosen port to the OS
cache dir (`ampersand/webhook-port`). `amp trigger` sends fixture events (see
`internal/fixtures/`, e.g. hubspot/stripe payloads) to that listener; `sync-ngrok` bridges
to a public ngrok URL. Webhook helpers live in `internal/webhook/`.

**Logging.** Use the `logger` package (`Info`, `Infof`, `Debugf`, `Fatal`, `FatalErr`).
`Debugf` output is gated on the `--debug` flag. Note: `logger` imports `flags`, so `flags`
must not import `logger` (there is an explicit circular-dependency workaround in
`flags/config.go`'s `GetProjectOrFail`).

## Conventions

- Command failures typically call `logger.Fatal`/`logger.FatalErr` (which `os.Exit(1)`),
  rather than returning errors up the stack — follow the pattern of neighboring commands.
- Naming (`amp <verb>:<noun>`): list commands use `list:integrations`,
  `list:installations`, etc.; deletes use `delete:integration`. Keep this colon style.
- Output format for list-style commands is controlled by a `--format/-f` flag (json|yaml);
  wire it with `flags.InitAndBindFormatFlag` and render with `utils.WriteStruct`.
