package main

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func sortedKeys[T any](entries *orderedmap.Map[string, T]) []string {
	keys := slices.Collect(entries.KeysFromOldest())
	slices.Sort(keys)
	return keys
}

func sortResponseCodes(codes []string) {
	slices.SortFunc(codes, func(a, b string) int {
		x, errX := strconv.Atoi(a)
		y, errY := strconv.Atoi(b)
		switch {
		case errX == nil && errY == nil:
			return cmp.Or(cmp.Compare(x, y), strings.Compare(a, b))
		case errX == nil:
			return -1
		case errY == nil:
			return 1
		default:
			return strings.Compare(a, b)
		}
	})
}

func extractOperations(paths *orderedmap.Map[string, *v3.PathItem]) []item {
	var items []item
	for name, path := range paths.FromOldest() {
		if path == nil {
			continue
		}
		for method, op := range path.GetOperations().FromOldest() {
			if op != nil {
				items = append(items, item{
					name: name, kind: strings.ToUpper(method),
					details: formatOperationDetails(op, path.Parameters), folded: true,
				})
			}
		}
	}
	slices.SortFunc(items, func(a, b item) int {
		return cmp.Or(strings.Compare(a.name, b.name), strings.Compare(a.kind, b.kind))
	})
	return items
}

func extractComponents(doc *v3.Document) []item {
	var items []item
	if doc.Components == nil {
		return items
	}
	add := func(name, kind, description, details string) {
		items = append(items, item{
			name: name, kind: kind, description: description, details: details, folded: true,
		})
	}
	components := doc.Components
	for name, proxy := range components.Schemas.FromOldest() {
		description := ""
		if proxy != nil && proxy.Schema() != nil {
			description = proxy.Schema().Description
		}
		add(name, "Schema", description, formatSchemaDetails(proxy))
	}
	for name, body := range components.RequestBodies.FromOldest() {
		if body != nil {
			add(name, "RequestBody", body.Description, formatRequestBodyDetails(body))
		}
	}
	for name, response := range components.Responses.FromOldest() {
		if response != nil {
			add(name, "Response", response.Description, formatResponseDetails(response))
		}
	}
	for name, parameter := range components.Parameters.FromOldest() {
		if parameter != nil {
			add(name, "Parameter", parameter.Description, formatParameterDetails(parameter))
		}
	}
	for name, header := range components.Headers.FromOldest() {
		if header != nil {
			add(name, "Header", header.Description, formatHeaderDetails(header))
		}
	}
	for name, scheme := range components.SecuritySchemes.FromOldest() {
		if scheme != nil {
			add(name, "SecurityScheme", scheme.Description, formatSecuritySchemeDetails(scheme))
		}
	}
	slices.SortFunc(items, func(a, b item) int {
		return cmp.Or(strings.Compare(a.kind, b.kind), strings.Compare(a.name, b.name))
	})
	return items
}

func formatOperationDetails(op *v3.Operation, inherited []*v3.Parameter) string {
	var details strings.Builder
	if op.Summary != "" {
		fmt.Fprintf(&details, "Summary: %s\n", op.Summary)
	}
	if op.Description != "" {
		fmt.Fprintf(&details, "Description: %s\n", op.Description)
	}
	if op.OperationId != "" {
		fmt.Fprintf(&details, "Operation ID: %s\n", op.OperationId)
	}
	if len(op.Tags) > 0 {
		fmt.Fprintf(&details, "Tags: %s\n", strings.Join(op.Tags, ", "))
	}
	if op.Deprecated != nil && *op.Deprecated {
		details.WriteString("Deprecated: true\n")
	}

	parameters := slices.Clone(op.Parameters)
	for _, parameter := range inherited {
		if parameter != nil && !slices.ContainsFunc(op.Parameters, func(p *v3.Parameter) bool {
			return p != nil && p.Name == parameter.Name && p.In == parameter.In
		}) {
			parameters = append(parameters, parameter)
		}
	}
	if len(parameters) > 0 {
		details.WriteString("Parameters:\n")
		for _, parameter := range parameters {
			if parameter != nil {
				fmt.Fprintf(&details, "  - %s (%s): %s\n", parameter.Name, parameter.In, parameter.Description)
				fmt.Fprintf(&details, "    Required: %t\n", parameter.Required != nil && *parameter.Required)
				details.WriteString(indentDetails(formatTypedSchema(parameter.Schema)))
			}
		}
	}
	if body := op.RequestBody; body != nil {
		details.WriteString("Request Body:\n")
		if body.Description != "" {
			fmt.Fprintf(&details, "  Description: %s\n", body.Description)
		}
		if body.Required != nil && *body.Required {
			details.WriteString("  Required: true\n")
		}
		details.WriteString(formatContent(body.Content))
	}
	if responses := op.Responses; responses != nil {
		details.WriteString("Responses:\n")
		codes := sortedKeys(responses.Codes)
		sortResponseCodes(codes)
		for _, code := range codes {
			if response := responses.Codes.GetOrZero(code); response != nil {
				fmt.Fprintf(&details, "  - %s: %s\n", code, response.Description)
				details.WriteString(indentDetails(formatResponseDetails(response)))
			}
		}
		if responses.Default != nil {
			fmt.Fprintf(&details, "  - default: %s\n", responses.Default.Description)
			details.WriteString(indentDetails(formatResponseDetails(responses.Default)))
		}
	}
	return details.String()
}

func indentDetails(text string) string {
	if text == "" {
		return ""
	}
	return "    " + strings.ReplaceAll(strings.TrimRight(text, "\n"), "\n", "\n    ") + "\n"
}

func formatContent(content *orderedmap.Map[string, *v3.MediaType]) string {
	var details strings.Builder
	for _, mediaType := range sortedKeys(content) {
		media := content.GetOrZero(mediaType)
		if media == nil {
			continue
		}
		fmt.Fprintf(&details, "  - %s", mediaType)
		if proxy := media.Schema; proxy != nil {
			if ref := proxy.GetReference(); ref != "" {
				parts := strings.Split(ref, "/")
				fmt.Fprintf(&details, " (schema: %s)", parts[len(parts)-1])
			} else if schema := proxy.Schema(); schema != nil && len(schema.Type) > 0 {
				if len(schema.Type) == 1 {
					fmt.Fprintf(&details, " (type: %s)", schema.Type[0])
				} else {
					fmt.Fprintf(&details, " (types: %v)", schema.Type)
				}
			}
		}
		details.WriteByte('\n')
	}
	return details.String()
}

func writeSchemaType(details *strings.Builder, label string, types []string) {
	switch len(types) {
	case 0:
	case 1:
		fmt.Fprintf(details, "%s: %s\n", label, types[0])
	default:
		fmt.Fprintf(details, "%ss: %v\n", label, types)
	}
}

func formatSchemaDetails(proxy *base.SchemaProxy) string {
	if proxy == nil || proxy.Schema() == nil {
		return "No schema details available"
	}
	schema := proxy.Schema()
	var details strings.Builder
	writeSchemaType(&details, "Type", schema.Type)
	if schema.Format != "" {
		fmt.Fprintf(&details, "Format: %s\n", schema.Format)
	}
	if len(schema.Required) > 0 {
		fmt.Fprintf(&details, "Required: %v\n", schema.Required)
	}
	if names := sortedKeys(schema.Properties); len(names) > 0 {
		details.WriteString("Properties:\n")
		for _, name := range names {
			if property := schema.Properties.GetOrZero(name); property != nil && property.Schema() != nil {
				types := property.Schema().Type
				kind := "unknown"
				if len(types) == 1 {
					kind = types[0]
				} else if len(types) > 1 {
					kind = fmt.Sprint(types)
				}
				fmt.Fprintf(&details, "  - %s: %s\n", name, kind)
			}
		}
	}
	if schema.Items != nil && schema.Items.A != nil && schema.Items.A.Schema() != nil {
		writeSchemaType(&details, "Items Type", schema.Items.A.Schema().Type)
	}
	return details.String()
}

func formatRequestBodyDetails(body *v3.RequestBody) string {
	if body == nil {
		return "No request body details available"
	}
	var details strings.Builder
	if body.Required != nil && *body.Required {
		details.WriteString("Required: true\n")
	}
	if content := formatContent(body.Content); content != "" {
		details.WriteString("Content Types:\n" + content)
	}
	return details.String()
}

func formatResponseDetails(response *v3.Response) string {
	if response == nil {
		return "No response details available"
	}
	var details strings.Builder
	if content := formatContent(response.Content); content != "" {
		details.WriteString("Content Types:\n" + content)
	}
	if headers := sortedKeys(response.Headers); len(headers) > 0 {
		details.WriteString("Headers:\n")
		for _, name := range headers {
			header := response.Headers.GetOrZero(name)
			if header != nil {
				fmt.Fprintf(&details, "  - %s", name)
				if header.Description != "" {
					fmt.Fprintf(&details, ": %s", header.Description)
				}
				details.WriteByte('\n')
				details.WriteString(indentDetails(formatHeaderDetails(header)))
			}
		}
	}
	return details.String()
}

func formatTypedSchema(proxy *base.SchemaProxy) string {
	var details strings.Builder
	if proxy != nil && proxy.Schema() != nil {
		schema := proxy.Schema()
		writeSchemaType(&details, "Type", schema.Type)
		if schema.Format != "" {
			fmt.Fprintf(&details, "Format: %s\n", schema.Format)
		}
	}
	return details.String()
}

func formatParameterDetails(parameter *v3.Parameter) string {
	if parameter == nil {
		return "No parameter details available"
	}
	var details strings.Builder
	fmt.Fprintf(&details, "In: %s\n", parameter.In)
	if parameter.Required != nil && *parameter.Required {
		details.WriteString("Required: true\n")
	}
	details.WriteString(formatTypedSchema(parameter.Schema))
	if parameter.Example != nil {
		fmt.Fprintf(&details, "Example: %v\n", parameter.Example)
	}
	return details.String()
}

func formatHeaderDetails(header *v3.Header) string {
	if header == nil {
		return "No header details available"
	}
	var details strings.Builder
	if header.Required {
		details.WriteString("Required: true\n")
	}
	details.WriteString(formatTypedSchema(header.Schema))
	return details.String()
}

func formatSecuritySchemeDetails(scheme *v3.SecurityScheme) string {
	if scheme == nil {
		return "No security scheme details available"
	}
	var details strings.Builder
	fmt.Fprintf(&details, "Type: %s\n", scheme.Type)
	for _, field := range [][2]string{
		{"Scheme", scheme.Scheme}, {"Bearer Format", scheme.BearerFormat},
		{"In", scheme.In}, {"Name", scheme.Name},
	} {
		if field[1] != "" {
			fmt.Fprintf(&details, "%s: %s\n", field[0], field[1])
		}
	}
	return details.String()
}
