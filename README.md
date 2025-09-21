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

## OpenAPI Support

`oq` supports all 3.* OpenAPI specification versions:

- 3.0
- 3.1
- 3.2

Both JSON and YAML formats are supported.

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
