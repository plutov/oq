# Code Structure

The oq OpenAPI viewer has been refactored into multiple files for better organization:

## Files Overview

### `main.go`
- Application entry point
- Handles command-line arguments and input (file vs stdin)
- Initializes OpenAPI document parsing and validation
- Sets up and runs the Bubble Tea program

### `model.go`
- Core Bubble Tea model definition and logic
- Model struct with application state
- Init, Update, and View methods for the TUI
- Event handling (keyboard navigation, mode switching)
- Navigation logic between endpoints and components

### `oas.go` (OpenAPI Specification)
- OpenAPI document parsing and data extraction
- Type definitions for `endpoint` and `component` structs
- Functions to extract endpoints and components from OpenAPI docs
- Detailed formatting functions for all component types:
  - Schemas, Request Bodies, Responses
  - Parameters, Headers, Security Schemes
- OpenAPI-specific business logic

### `view.go`
- UI rendering and presentation logic
- Methods for rendering different views:
  - `renderEndpoints()` - Displays API endpoints with color-coded HTTP methods
  - `renderComponents()` - Shows components with expandable details
  - `renderHeader()` - Title and tab navigation
  - `renderFooter()` - Help text
- Styling and color definitions
- Lipgloss-based UI components

### `example.yaml`
- Sample OpenAPI specification for testing
- Comprehensive example covering all supported features

### `README.md`
- Documentation and usage instructions
- Feature overview and keyboard shortcuts

## Benefits of This Structure

1. **Separation of Concerns**: Each file has a clear, single responsibility
2. **Maintainability**: Easier to find and modify specific functionality
3. **Readability**: Smaller, focused files are easier to understand
4. **Testability**: Functions can be tested in isolation
5. **Extensibility**: New features can be added without cluttering existing files

## Architecture Flow

```
main.go → Parse OpenAPI → NewModel (model.go)
                            ↓
Model uses oas.go for data extraction
                            ↓
Model uses view.go for UI rendering
                            ↓
Bubble Tea handles events via model.go
```