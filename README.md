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

`oq` supports both modern major OpenAPI specification versions:

- **OpenAPI 3.0.x**
- **OpenAPI 3.1.x** - by hacky conversion for now until [kin-openapi adds 3.1 support](https://github.com/getkin/kin-openapi/issues/230)

Both JSON and YAML formats are supported.

## Installation

Using Homebrew (macOS/Linux):

```bash
brew install plutov/tap/oq
```

Or using go install:

```bash
go install github.com/plutov/oq@latest
```

You can also download the compiled binaries from the Releases page.

### From source

```bash
git clone git@github.com:plutov/oq.git
cd oq
go build -o oq .
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

When contributing:
1. Ensure tests pass: `go test -v`
2. Test with both OpenAPI 3.0 and 3.1 examples
