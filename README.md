# Go Server Side Rendering Experiment

Using Golang with `templ` and `htmx` to serve an HDA with Tailwind layout.

## How to get started

Initialize environment

```sh
npm install # to install dependencies
cp node_modules/htmx.org/dist/htmx.min.js web/assets/js
cp node_modules/htmx.org/dist/ext/response-targets.js web/assets/js
cp node_modules/htmx.org/dist/ext/remove-me.js web/assets/js
cp node_modules/hyperscript.org/dist/_hyperscript.min.js web/assets/js/hyperscript.min.js
npm run build # to create CSS asset

go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air@latest
go install github.com/nicksnyder/go-i18n/v2/goi18n@latest
```

## Hot deploy

Watch CSS changes to update Tailwind

```sh
npm run watch
```

Watch `templ` and Golang changes to restart the server

```sh
air
```

## Update Translations

Merge translations iteratively after each code update

```sh
goi18n extract -format toml --outdir locale
cd locale
goi18n merge active.*.toml
```

Provide new translations and finally merge, e.g. `translate.*.toml`

```sh
cd locale
goi18n merge active.*.toml translate.*.toml
```

## Run Golang lint

Golang lint requires <https://golangci-lint.run/>

```sh
golangci-lint run
```

## Use PostgreSQL

Configure a server and create a database, e.g. Docker based

```sh
docker run --name task-db -p 5432:5432 \
  -e POSTGRES_USER=task-db-user \
  -e POSTGRES_PASSWORD=my53cr3tpa55w0rd \
  -d postgres:17-alpine
```

Access the database

```sh
docker exec -it task-db psql -U task-db-user
```

Install database migration tool and run migrations

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir postgres postgres "postgres://task-db-user:my53cr3tpa55w0rd@localhost" up
```

## Features (non functional techy stuff)

- [x] server side rendering (Golang templ)
- [x] hypermedia driven client interaction (htmx)
- [x] notification snackbar (htmx OOB and extension)
- [x] Markdown rending and styling (goldmark and Tailwind)
- [x] table sorting
- [x] table paging
- [x] Golang enum string mapping
- [x] browser history update and reload of subpages
- [x] JSON logging (Golang slog)
- [x] i18n - translate template labels and notification messages
- [x] l10n - display dates in localized format (simple formatting helpers)
