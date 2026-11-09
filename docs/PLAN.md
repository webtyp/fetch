---
PLAN: "feat: fetch.Patch — the verb REST uses for a partial update"
EXECUTOR: antigravity
REVIEWER: none
STATUS: running
SESSION: 16452781565504915696
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit after every stage and stop. Never ask whether to
> proceed. This plan runs **locally**: do not tag, push or publish anything; the owner reviews first.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## Goal

`rpc.Caller` (a later plan) sends an update as `PATCH /api/<module>/<entity>` and a delete as `DELETE` with the ids in the body. `fetch` has `Get`, `Post`,
`Put`, `Delete` but no `Patch`. This plan adds it, and proves that `DELETE` carries a body on both clients (stdlib and WASM).

## Ecosystem search

Searched `Patch`, `PATCH` over `webtyp/*`: `fetch.go` builds every verb as `&Request{method: "<VERB>", endpoint: endpoint}` and both clients (`client_stdlib.go:40`,
`client_wasm.go:37`) pass `r.method` through unchanged. **Extend** `fetch` with one more constructor; no new library.

## Design gate

1. **Prior art** (unverified in this session): the Fetch standard and `net/http` both treat `PATCH` as an ordinary method with a body.
2. **Novice-name test**: `fetch.Patch(url)` — the same word HTTP uses.
3. **Ledger**: concepts +1 (`Patch`); lines at the `rpc` call site: the verb is chosen with `fetch.Patch` instead of building the method by hand; ways to send a partial update 0 → 1.
4. **Where**: `fetch` already owns the verbs.
5. **Deletes**: nothing.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## APIs used (inside this repo — read first)

| symbol | exact signature | file |
|---|---|---|
| `fetch.Put` | `func Put(endpoint any) *Request` | `fetch.go:56` |
| `(*Request).Body` | `func (r *Request) Body(b []byte) *Request` (used as `req.Body(body).Send(…)` by `rpc.Caller`) — read `fetch.go` for the exact signature | `fetch.go` |
| `(*Request).Send` | `func (r *Request) Send(cb func(*Response, error))` | `fetch.go` |

If a signature differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — tests first (red)

`fetch_shared_test.go` (shared by the stdlib and the WASM runs) gets:

```go
func SendRequest_PatchShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	var body string

	fetch.Patch(baseURL + "/patch").ContentTypeJSON().Body([]byte(`{"name":"b"}`)).Send(func(resp *fetch.Response, err error) {
		if err == nil {
			body = resp.Text()
		}
		done <- true
	})
	<-done
	if body != `patched {"name":"b"}` {
		t.Errorf("Patch failed: %s", body)
	}
}

func SendRequest_DeleteWithBodyShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	var body string

	fetch.Delete(baseURL + "/delete-body").ContentTypeJSON().Body([]byte(`{"ids":["a"]}`)).Send(func(resp *fetch.Response, err error) {
		if err == nil {
			body = resp.Text()
		}
		done <- true
	})
	<-done
	if body != `deleted {"ids":["a"]}` {
		t.Errorf("Delete with body failed: %s", body)
	}
}
```

`server_test.go`: two handlers next to `/put` and `/delete`:

```go
	mux.HandleFunc("/patch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		b, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("patched " + string(b)))
	})

	mux.HandleFunc("/delete-body", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		b, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deleted " + string(b)))
	})
```

Register both in `fetch_stlib_test.go` (next to the `PutDelete` line) and in `fetch_wasm_test.go:32`:

```go
	t.Run("Patch", func(t *testing.T) { SendRequest_PatchShared(t, server.URL) })
	t.Run("DeleteWithBody", func(t *testing.T) { SendRequest_DeleteWithBodyShared(t, server.URL) })
```

(in `fetch_wasm_test.go` the variable is `serverURL`, as the neighbouring line shows). Run `gotest`: it must fail to compile (`fetch.Patch` missing). Commit: `test: Patch and Delete with a body (red)`.

## Stage 2 — the code

In `fetch.go`, after `Put`:

```go
// Patch creates a new PATCH request: a partial update.
//
//	fetch.Patch(url).ContentTypeJSON().Body(body).Send(done)
func Patch(endpoint any) *Request {
	return &Request{method: "PATCH", endpoint: endpoint}
}
```

If `TestSendRequest_DeleteWithBody` fails on one client, the client drops the body of a `DELETE`: fix that client (`client_stdlib.go` around line 40, `client_wasm.go` around line 37) so the body is sent for every method except `GET` and `HEAD`;
do not special-case `rpc`. `gotest` passes. Commit: `feat: fetch.Patch`.

## Stage 3 — documentation

- `README.md`: the line "**Tiny API**: `Get`, `Post`, `Put`, `Delete`" gains `Patch`; keep ≤ 150 lines and the four ordered sections.
- `docs/API.md`: add `### func Patch(url string) *Request` with "Creates a new PATCH request: a partial update." after `Put`.
- `docs/CORS.md:26`: the example `Access-Control-Allow-Methods` becomes `"GET, POST, PUT, PATCH, DELETE, OPTIONS"`; add one sentence: `PATCH` and `DELETE` from another origin trigger a preflight.
- Do not link `docs/PLAN.md` from permanent docs.

Commit: `docs: Patch`.

## Delivery check

1. `gotest` (builds, vets, runs the stdlib and the WASM tests; a PR that does not compile, or whose tests could not run, is a failed delivery).
2. `grep -rn '"PATCH"' --include=*.go .` lists `fetch.go` and the test files only.
3. `git diff --stat main` lists only `fetch.go`, the client file if stage 2 needed it, the three test files, `README.md`, `docs/API.md`, `docs/CORS.md`.

| Stage | Files | Commit |
|---|---|---|
| 1 | `fetch_shared_test.go`, `server_test.go`, `fetch_stlib_test.go`, `fetch_wasm_test.go` | `test: … (red)` |
| 2 | `fetch.go` (+ a client file only if needed) | `feat: fetch.Patch` |
| 3 | `README.md`, `docs/API.md`, `docs/CORS.md` | `docs: Patch` |
