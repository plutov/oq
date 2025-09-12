# oq

<p align="center"><img src="preview.gif" width="500" alt="oq preview"></p>

A terminal-based OpenAPI Spec (OAS) viewer and processor built with Go and Bubble Tea.

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

`oq` supports major OpenAPI specification versions:
- 3.0.x
- 3.1.x

Both JSON and YAML formats are supported.

## Installation

### From Source

```bash
git clone git@github.com:plutov/oq.git
cd oq
go build -o oq .
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.