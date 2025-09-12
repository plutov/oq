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

- **↑/↓ or k/j** - Navigate up/down through items
- **Tab** - Switch between Endpoints and Components views
- **Enter or Space** - Toggle fold/unfold for endpoint and component details
- **q or Ctrl+C** - Quit the application

## OpenAPI Support

`oq` supports both modern major OpenAPI specification versions:

- **OpenAPI 3.0.x**
- **OpenAPI 3.1.x**

Both JSON and YAML formats are supported.

## Installation

```bash
go install github.com/plutov/oq@latest
```

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