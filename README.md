# oq - a terminal-based OpenAPI Spec (OAS) viewer

<p align="center"><img src="preview.gif" width="500" alt="oq preview"></p>

## Usage

```bash
oq openapi.yaml
# or
cat openapi.yaml | oq
# or
curl https://api.example.com/openapi.json | oq
```

### Keyboard Shortcuts

Press `?` to see the help screen with all available keyboard shortcuts.

| Key | Action |
| --- | --- |
| `j`/`k`, `↓`/`↑` | Select next/previous item |
| `gg` / `G` | Select first/last item |
| `Ctrl-U` / `Ctrl-D` | Move up/down half a screen of items |
| `Enter` / `Space` | Expand/collapse details (collapsed by default) |
| `PgUp` / `PgDown`, mouse wheel | Scroll lines, including long expanded details |
| `Tab` / `Shift-Tab`, `L` / `H` | Cycle forward/backward through views |
| `/` | Start incremental, case-insensitive literal search |
| `Enter` / `Esc` during search | Accept / cancel and restore the previous selection |
| `n` / `N` | Next/previous match, wrapping at either end |
| `q` / `Ctrl-C` | Quit (`Ctrl-C` also works during search) |

Search covers names, kinds, descriptions and details in Requests, Components and
Webhooks, even while details are collapsed. Each view remembers its accepted
query. While typing a query, spaces, `q`, `?` and pasted Unicode text are input,
not commands; Backspace edits the query. The footer shows match counts or
`No matches`. Accept an empty search to clear it. Press `Esc` or `q` to close help.

A proportional scrollbar on the right shows the visible portion of overflowing
content. Line scrolling can move the selected item offscreen; item navigation
brings it back into view.

## OpenAPI Support

`oq` supports all 3.* OpenAPI specification versions:

- 3.0
- 3.1
- 3.2

Both JSON and YAML formats are supported. Try `oq examples/catalog-3.2.yaml`
for an OpenAPI 3.2 example with `QUERY`, `additionalOperations`, webhooks and
schemas. Expanded operations show inherited parameters, required/type/format
information, request and response content, response headers, tags and deprecation.

Note: `oq` uses the [libopenapi](https://github.com/pb33f/libopenapi) library as it supports all OpenAPI versions and is actively maintained.

## Installation

Using go install:

```bash
go install github.com/plutov/oq@latest
```

<details>
<summary>Package managers</summary>

Using Homebrew (macOS/Linux):

```bash
brew install plutov/tap/oq
```

Arch Linux (AUR):

```bash
yay -S oq-openapi-viewer-git
```

</details>

You can also download the compiled binaries from the Releases page.

### From source

```bash
git clone git@github.com:plutov/oq.git
cd oq
go build -o oq .
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

When contributing:

1. Ensure tests pass: `go test -v`
2. Test all supported OpenAPI versions (3.0, 3.1, 3.2)
3. If the UI changes, make sure to run `vhs preview.tape` to generate a new preview GIF
4. Try to extend test coverage by introducing new example OpenAPI specs in the `examples` folder
