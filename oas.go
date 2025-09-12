package main

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func extractEndpoints(doc *openapi3.T) []endpoint {
	var endpoints []endpoint

	for path, pathItem := range doc.Paths.Map() {
		if pathItem.Get != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "GET", op: pathItem.Get, folded: true})
		}
		if pathItem.Post != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "POST", op: pathItem.Post, folded: true})
		}
		if pathItem.Put != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "PUT", op: pathItem.Put, folded: true})
		}
		if pathItem.Delete != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "DELETE", op: pathItem.Delete, folded: true})
		}
		if pathItem.Patch != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "PATCH", op: pathItem.Patch, folded: true})
		}
		if pathItem.Head != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "HEAD", op: pathItem.Head, folded: true})
		}
		if pathItem.Options != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "OPTIONS", op: pathItem.Options, folded: true})
		}
		if pathItem.Trace != nil {
			endpoints = append(endpoints, endpoint{path: path, method: "TRACE", op: pathItem.Trace, folded: true})
		}
	}

	return endpoints
}

func extractComponents(doc *openapi3.T) []component {
	var components []component

	if doc.Components != nil {
		if doc.Components.Schemas != nil {
			for name, schema := range doc.Components.Schemas {
				details := formatSchemaDetails(name, schema.Value)
				description := ""
				if schema.Value != nil && schema.Value.Description != "" {
					description = schema.Value.Description
				}
				components = append(components, component{
					name:        name,
					compType:    "Schema",
					description: description,
					details:     details,
					folded:      true,
				})
			}
		}
		if doc.Components.RequestBodies != nil {
			for name, reqBody := range doc.Components.RequestBodies {
				details := formatRequestBodyDetails(name, reqBody.Value)
				description := ""
				if reqBody.Value != nil && reqBody.Value.Description != "" {
					description = reqBody.Value.Description
				}
				components = append(components, component{
					name:        name,
					compType:    "RequestBody",
					description: description,
					details:     details,
					folded:      true,
				})
			}
		}
		if doc.Components.Responses != nil {
			for name, resp := range doc.Components.Responses {
				details := formatResponseDetails(name, resp.Value)
				description := ""
				if resp.Value != nil && resp.Value.Description != nil {
					description = *resp.Value.Description
				}
				components = append(components, component{
					name:        name,
					compType:    "Response",
					description: description,
					details:     details,
					folded:      true,
				})
			}
		}
		if doc.Components.Parameters != nil {
			for name, param := range doc.Components.Parameters {
				details := formatParameterDetails(name, param.Value)
				description := ""
				if param.Value != nil && param.Value.Description != "" {
					description = param.Value.Description
				}
				components = append(components, component{
					name:        name,
					compType:    "Parameter",
					description: description,
					details:     details,
					folded:      true,
				})
			}
		}
		if doc.Components.Headers != nil {
			for name, header := range doc.Components.Headers {
				details := formatHeaderDetails(name, header.Value)
				description := ""
				if header.Value != nil && header.Value.Description != "" {
					description = header.Value.Description
				}
				components = append(components, component{
					name:        name,
					compType:    "Header",
					description: description,
					details:     details,
					folded:      true,
				})
			}
		}
		if doc.Components.SecuritySchemes != nil {
			for name, secScheme := range doc.Components.SecuritySchemes {
				details := formatSecuritySchemeDetails(name, secScheme.Value)
				description := ""
				if secScheme.Value != nil && secScheme.Value.Description != "" {
					description = secScheme.Value.Description
				}
				components = append(components, component{
					name:        name,
					compType:    "SecurityScheme",
					description: description,
					details:     details,
					folded:      true,
				})
			}
		}
	}

	return components
}

func formatEndpointDetails(ep endpoint) string {
	var details strings.Builder

	if ep.op.Summary != "" {
		details.WriteString(fmt.Sprintf("Summary: %s\n", ep.op.Summary))
	}

	if ep.op.Description != "" {
		details.WriteString(fmt.Sprintf("Description: %s\n", ep.op.Description))
	}

	if len(ep.op.Parameters) > 0 {
		details.WriteString("Parameters:\n")
		for _, param := range ep.op.Parameters {
			if param.Value != nil {
				details.WriteString(fmt.Sprintf("  - %s (%s): %s\n",
					param.Value.Name, param.Value.In, param.Value.Description))
			}
		}
	}

	if ep.op.RequestBody != nil && ep.op.RequestBody.Value != nil {
		details.WriteString("Request Body:\n")
		for mediaType := range ep.op.RequestBody.Value.Content {
			details.WriteString(fmt.Sprintf("  - %s\n", mediaType))
		}
	}

	if ep.op.Responses != nil {
		details.WriteString("Responses:\n")
		for code, resp := range ep.op.Responses.Map() {
			if resp.Value != nil && resp.Value.Description != nil {
				details.WriteString(fmt.Sprintf("  - %s: %s\n", code, *resp.Value.Description))
			}
		}
	}

	return details.String()
}

func formatSchemaDetails(name string, schema *openapi3.Schema) string {
	var details strings.Builder

	if schema == nil {
		return "No schema details available"
	}

	// Handle both single type (OpenAPI 3.0) and array of types (OpenAPI 3.1)
	if schema.Type != nil && len(*schema.Type) > 0 {
		types := *schema.Type
		if len(types) == 1 {
			details.WriteString(fmt.Sprintf("Type: %s\n", types[0]))
		} else {
			details.WriteString(fmt.Sprintf("Types: %v\n", types))
		}
	}

	if schema.Format != "" {
		details.WriteString(fmt.Sprintf("Format: %s\n", schema.Format))
	}

	if len(schema.Required) > 0 {
		details.WriteString(fmt.Sprintf("Required: %v\n", schema.Required))
	}

	if len(schema.Properties) > 0 {
		details.WriteString("Properties:\n")
		for propName, prop := range schema.Properties {
			propType := "unknown"
			if prop.Value != nil && prop.Value.Type != nil && len(*prop.Value.Type) > 0 {
				types := *prop.Value.Type
				if len(types) == 1 {
					propType = types[0]
				} else {
					propType = fmt.Sprintf("%v", types)
				}
			}
			details.WriteString(fmt.Sprintf("  - %s: %s\n", propName, propType))
		}
	}

	if schema.Items != nil && schema.Items.Value != nil && schema.Items.Value.Type != nil && len(*schema.Items.Value.Type) > 0 {
		types := *schema.Items.Value.Type
		if len(types) == 1 {
			details.WriteString(fmt.Sprintf("Items Type: %s\n", types[0]))
		} else {
			details.WriteString(fmt.Sprintf("Items Types: %v\n", types))
		}
	}

	return details.String()
}

func formatRequestBodyDetails(name string, reqBody *openapi3.RequestBody) string {
	var details strings.Builder

	if reqBody == nil {
		return "No request body details available"
	}

	if reqBody.Required {
		details.WriteString("Required: true\n")
	}

	if len(reqBody.Content) > 0 {
		details.WriteString("Content Types:\n")
		for mediaType, mediaTypeObj := range reqBody.Content {
			details.WriteString(fmt.Sprintf("  - %s", mediaType))
			if mediaTypeObj.Schema != nil && mediaTypeObj.Schema.Value != nil && mediaTypeObj.Schema.Value.Type != nil && len(*mediaTypeObj.Schema.Value.Type) > 0 {
				types := *mediaTypeObj.Schema.Value.Type
				if len(types) == 1 {
					details.WriteString(fmt.Sprintf(" (type: %s)", types[0]))
				} else {
					details.WriteString(fmt.Sprintf(" (types: %v)", types))
				}
			}
			details.WriteString("\n")
		}
	}

	return details.String()
}

func formatResponseDetails(name string, response *openapi3.Response) string {
	var details strings.Builder

	if response == nil {
		return "No response details available"
	}

	if len(response.Content) > 0 {
		details.WriteString("Content Types:\n")
		for mediaType, mediaTypeObj := range response.Content {
			details.WriteString(fmt.Sprintf("  - %s", mediaType))
			if mediaTypeObj.Schema != nil && mediaTypeObj.Schema.Value != nil && mediaTypeObj.Schema.Value.Type != nil && len(*mediaTypeObj.Schema.Value.Type) > 0 {
				types := *mediaTypeObj.Schema.Value.Type
				if len(types) == 1 {
					details.WriteString(fmt.Sprintf(" (type: %s)", types[0]))
				} else {
					details.WriteString(fmt.Sprintf(" (types: %v)", types))
				}
			}
			details.WriteString("\n")
		}
	}

	if len(response.Headers) > 0 {
		details.WriteString("Headers:\n")
		for headerName := range response.Headers {
			details.WriteString(fmt.Sprintf("  - %s\n", headerName))
		}
	}

	return details.String()
}

func formatParameterDetails(name string, param *openapi3.Parameter) string {
	var details strings.Builder

	if param == nil {
		return "No parameter details available"
	}

	details.WriteString(fmt.Sprintf("In: %s\n", param.In))

	if param.Required {
		details.WriteString("Required: true\n")
	}

	if param.Schema != nil && param.Schema.Value != nil && param.Schema.Value.Type != nil && len(*param.Schema.Value.Type) > 0 {
		types := *param.Schema.Value.Type
		if len(types) == 1 {
			details.WriteString(fmt.Sprintf("Type: %s\n", types[0]))
		} else {
			details.WriteString(fmt.Sprintf("Types: %v\n", types))
		}
		if param.Schema.Value.Format != "" {
			details.WriteString(fmt.Sprintf("Format: %s\n", param.Schema.Value.Format))
		}
	}

	if param.Example != nil {
		details.WriteString(fmt.Sprintf("Example: %v\n", param.Example))
	}

	return details.String()
}

func formatHeaderDetails(name string, header *openapi3.Header) string {
	var details strings.Builder

	if header == nil {
		return "No header details available"
	}

	if header.Required {
		details.WriteString("Required: true\n")
	}

	if header.Schema != nil && header.Schema.Value != nil && header.Schema.Value.Type != nil && len(*header.Schema.Value.Type) > 0 {
		types := *header.Schema.Value.Type
		if len(types) == 1 {
			details.WriteString(fmt.Sprintf("Type: %s\n", types[0]))
		} else {
			details.WriteString(fmt.Sprintf("Types: %v\n", types))
		}
		if header.Schema.Value.Format != "" {
			details.WriteString(fmt.Sprintf("Format: %s\n", header.Schema.Value.Format))
		}
	}

	return details.String()
}

func formatSecuritySchemeDetails(name string, secScheme *openapi3.SecurityScheme) string {
	var details strings.Builder

	if secScheme == nil {
		return "No security scheme details available"
	}

	details.WriteString(fmt.Sprintf("Type: %s\n", secScheme.Type))

	if secScheme.Scheme != "" {
		details.WriteString(fmt.Sprintf("Scheme: %s\n", secScheme.Scheme))
	}

	if secScheme.BearerFormat != "" {
		details.WriteString(fmt.Sprintf("Bearer Format: %s\n", secScheme.BearerFormat))
	}

	if secScheme.In != "" {
		details.WriteString(fmt.Sprintf("In: %s\n", secScheme.In))
	}

	if secScheme.Name != "" {
		details.WriteString(fmt.Sprintf("Name: %s\n", secScheme.Name))
	}

	return details.String()
}
