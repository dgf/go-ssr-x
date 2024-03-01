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

## Features (non functional techy stuff)

- [x] server side rendering (Golang templ)
- [x] hypermedia driven client interaction (htmx)
- [x] notification snackbar (htmx OOB and extension)
- [x] Markdown rending and styling (goldmark and tailwind)
- [x] table sorting
- [ ] table paging
- [x] Golang enum string mapping
- [x] browser history update and reload of subpages
- [ ] i18n - translate template labels and notification messages
- [ ] l10n - display dates in localized format
