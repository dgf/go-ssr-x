# Go Server Side Rendering Experiment

using Golang with `templ` and `htmx` to serve HDA with Tailwind layout

## How to get started

init environment

```sh
npm install # to install dependencies
cp node_modules/htmx.org/dist/htmx.min.js assets/js
cp node_modules/htmx.org/dist/ext/response-targets.js assets/js
cp node_modules/htmx.org/dist/ext/remove-me.js assets/js
npm run build # to create CSS asset

go install github.com/a-h/templ/cmd/templ@latest
go install github.com/cosmtrek/air@latest
go install github.com/nicksnyder/go-i18n/v2/goi18n@latest
```

## Hot deploy

watch CSS changes to update Tailwind

```sh
npm run watch
```

watch `templ` and Golang changes to restart the server

```sh
air
```

## Update Translations

extract defaults from code base (initial steps)

```sh
goi18n extract -format toml --outdir locale
cd locale
goi18n merge active.en.toml translate.de.toml
```

```sh
goi18n extract -format toml --outdir locale
cd locale
goi18n merge active.*.toml
```

## Features (non functional techy stuff)

- [x] server side rendering (Golang templ)
- [x] hypermedia driven client interaction (htmx)
- [x] notification snackbar (htmx OOB and extension)
- [x] Markdown rending and styling (goldmark and Tailwind)
- [x] table sorting
- [ ] table paging
- [x] Golang enum string mapping
- [x] browser history update and reload of subpages
- [x] JSON logging (Golang slog)
- [x] i18n - translate template labels and notification messages
- [x] l10n - display dates in localized format (simple formatting helpers)
