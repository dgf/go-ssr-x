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
