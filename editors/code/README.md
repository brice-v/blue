# Blue for VS Code

Minimal client for the blue language server (`blue lsp`). It brings diagnostics,
completion, hover, navigation and semantic highlighting to `.b` files.

## Prerequisites

1. Go 1.25 or newer.
2. Build the binary the extension will launch:

   ```sh
   go build -o blue .
   ```

   If your `GOFLAGS` need build tags (for example `-tags=rgfw`), build with them
   here too, or install the binary into a folder you keep on `PATH`.

## Install from this repo

From inside the repository:

```sh
cd editors/code
npm install          # pulls vscode-languageclient
npx @vscode/vsce package
code --install-extension blue-vscode-0.0.1.vsix
```

By default the extension finds `blue` on `PATH`, so nothing to configure when the
editor inherits your shell environment. To pin a specific build instead, set
`blue.path` in settings.json:

```json
{
  "blue.path": "/absolute/path/to/blue"
}
```

Any VS Code window with `.b` files then gets Blue in the language picker, and the
server starts automatically per window. `Blue: Restart language server` restarts it without
reloading the window.

## Settings

- `blue.path`: executable used to run `lsp`. Empty means discover `blue` on PATH,
  which skips any build whose `help` output lacks the `lsp` subcommand so a stale
  install cannot shadow a working one. Set it explicitly whenever the editor's PATH
  differs from your terminal (typical for GUI launches), or when you want an exact
  build such as `/home/brice/wss/blue/blue`.
- `blue.trace.server`: set to `verbose` to launch the server with `--trace`, which
  logs every protocol message to stderr so a broken request can be diagnosed.

## Notes

- Colors come entirely from the server's semantic tokens (keywords, comments,
  strings, numbers, functions, parameters, constants, modules, properties). There is
  no TextMate grammar yet, so text that the server does not classify is left at the
  theme default rather than guessed.
- Formatting and rename are not offered because emitting an edit that cannot be
  compiled is worse than offering nothing.
- If the server process dies, run `Blue: Restart language server` from the command
  palette instead of reloading the window.

## Packaging files

`node_modules/`, `out/` and `*.vsix` are ignored by git, so packaging always
happens against a fresh install in this folder.
