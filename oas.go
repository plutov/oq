package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// sortResponseCodes sorts HTTP response codes with stable ordering:
// 1. Numeric codes sorted numerically (100, 200, 201, 400, 404, 500)
// 2. Non-numeric codes sorted alphabetically (default)
func sortResponseCodes(codes []string) {
	sort.Slice(codes, func(i, j int) bool {
		codeI, errI := strconv.Atoi(codes[i])
		codeJ, errJ := strconv.Atoi(codes[j])

		// Both are numeric - sort numerically
		if errI == nil && errJ == nil {
			return codeI < codeJ
		}

		// One numeric, one non-numeric - numeric comes first
		if errI == nil && errJ != nil {
			return true
		}
		if errI != nil && errJ == nil {
			return false
		}

		// Both non-numeric - sort alphabetically
		return codes[i] < codes[j]
	})
}

func extractEndpoints(doc *v3.Document) []endpoint {
	var endpoints []endpoint

	if doc.Paths == nil || doc.Paths.PathItems == nil {
		return endpoints
	}

	// Iterate through path items using the orderedmap methods
	for pair := doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		path := pair.Key()
		pathItem := pair.Value()

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

	// Sort endpoints for stable ordering: first by path, then by method
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].path != endpoints[j].path {
			return endpoints[i].path < endpoints[j].path
		}
		return endpoints[i].method < endpoints[j].method
	})

	return endpoints
}

func extractWebhooks(doc *v3.Document) []webhook {
	var webhooks []webhook

	if doc.Webhooks != nil {
		for pair := doc.Webhooks.First(); pair != nil; pair = pair.Next() {
			name := pair.Key()
			hook := pair.Value()
			if hook != nil {
				if hook.Get != nil {
					webhooks = append(webhooks, webhook{name: name, method: "GET", op: hook.Get, folded: true})
				}
				if hook.Post != nil {
					webhooks = append(webhooks, webhook{name: name, method: "POST", op: hook.Post, folded: true})
				}
				if hook.Put != nil {
					webhooks = append(webhooks, webhook{name: name, method: "PUT", op: hook.Put, folded: true})
				}
				if hook.Delete != nil {
					webhooks = append(webhooks, webhook{name: name, method: "DELETE", op: hook.Delete, folded: true})
				}
				if hook.Patch != nil {
					webhooks = append(webhooks, webhook{name: name, method: "PATCH", op: hook.Patch, folded: true})
				}
				if hook.Head != nil {
					webhooks = append(webhooks, webhook{name: name, method: "HEAD", op: hook.Head, folded: true})
				}
				if hook.Options != nil {
					webhooks = append(webhooks, webhook{name: name, method: "OPTIONS", op: hook.Options, folded: true})
				}
				if hook.Trace != nil {
					webhooks = append(webhooks, webhook{name: name, method: "TRACE", op: hook.Trace, folded: true})
				}
			}
		}
	}

	// Sort webhooks for stable ordering: first by name, then by method
	sort.Slice(webhooks, func(i, j int) bool {
		if webhooks[i].name != webhooks[j].name {
			return webhooks[i].name < webhooks[j].name
		}
		return webhooks[i].method < webhooks[j].method
	})

	return webhooks
}

func extractComponents(doc *v3.Document) []component {
	var components []component

	if doc.Components != nil {
		if doc.Components.Schemas != nil {
			for pair := doc.Components.Schemas.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				schema := pair.Value()
				details := formatSchemaDetails(schema)
				description := ""
				if schema != nil && schema.Schema() != nil && schema.Schema().Description != "" {
					description = schema.Schema().Description
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
			for pair := doc.Components.RequestBodies.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				reqBody := pair.Value()
				details := formatRequestBodyDetails(reqBody)
				description := ""
				if reqBody != nil && reqBody.Description != "" {
					description = reqBody.Description
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
			for pair := doc.Components.Responses.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				resp := pair.Value()
				details := formatResponseDetails(resp)
				description := ""
				if resp != nil && resp.Description != "" {
					description = resp.Description
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
			for pair := doc.Components.Parameters.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				param := pair.Value()
				details := formatParameterDetails(param)
				description := ""
				if param != nil && param.Description != "" {
					description = param.Description
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
			for pair := doc.Components.Headers.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				header := pair.Value()
				details := formatHeaderDetails(header)
				description := ""
				if header != nil && header.Description != "" {
					description = header.Description
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
			for pair := doc.Components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				secScheme := pair.Value()
				details := formatSecuritySchemeDetails(name, secScheme)
				description := ""
				if secScheme != nil && secScheme.Description != "" {
					description = secScheme.Description
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

	// Sort components for stable ordering: first by type, then by name
	sort.Slice(components, func(i, j int) bool {
		if components[i].compType != components[j].compType {
			return components[i].compType < components[j].compType
		}
		return components[i].name < components[j].name
	})

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
			if param != nil {
				details.WriteString(fmt.Sprintf("  - %s (%s): %s\n",
					param.Name, param.In, param.Description))
			}
		}
	}

	if ep.op.RequestBody != nil {
		details.WriteString("Request Body:\n")

		if ep.op.RequestBody.Description != "" {
			details.WriteString(fmt.Sprintf("  Description: %s\n", ep.op.RequestBody.Description))
		}

		if ep.op.RequestBody.Required != nil && *ep.op.RequestBody.Required {
			details.WriteString("  Required: true\n")
		}

		// Get media types and sort them for stable ordering
		var mediaTypes []string
		if ep.op.RequestBody.Content != nil {
			for pair := ep.op.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
				mediaTypes = append(mediaTypes, pair.Key())
			}
		}
		sort.Strings(mediaTypes)

		// Display media type with schema information - either a reference to a component schema
		// (e.g., "#/components/schemas/Pet") or an inline schema type (e.g., "object", "string")
		for _, mediaType := range mediaTypes {
			if mediaTypeObj, ok := ep.op.RequestBody.Content.Get(mediaType); ok && mediaTypeObj != nil {
				details.WriteString(fmt.Sprintf("  - %s", mediaType))
				if mediaTypeObj.Schema != nil {
					// Check if it's a reference first
					if mediaTypeObj.Schema.GetReference() != "" {
						ref := mediaTypeObj.Schema.GetReference()
						// Extract just the schema name from the reference path
						parts := strings.Split(ref, "/")
						if len(parts) > 0 {
							schemaName := parts[len(parts)-1]
							details.WriteString(fmt.Sprintf(" (schema: %s)", schemaName))
						}
					} else if mediaTypeObj.Schema.Schema() != nil && len(mediaTypeObj.Schema.Schema().Type) > 0 {
						// If it's an inline schema with a type
						types := mediaTypeObj.Schema.Schema().Type
						if len(types) == 1 {
							details.WriteString(fmt.Sprintf(" (type: %s)", types[0]))
						} else {
							details.WriteString(fmt.Sprintf(" (types: %v)", types))
						}
					}
				}
				details.WriteString("\n")
			}
		}
	}

	if ep.op.Responses != nil {
		details.WriteString("Responses:\n")

		// Get response codes and sort them for stable ordering
		var codes []string
		if ep.op.Responses.Codes != nil {
			for pair := ep.op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
				codes = append(codes, pair.Key())
			}
		}

		// Sort by status code numerically, then alphabetically for non-numeric codes
		sortResponseCodes(codes)

		for _, code := range codes {
			if resp, ok := ep.op.Responses.Codes.Get(code); ok && resp != nil {
				if resp.Description != "" {
					details.WriteString(fmt.Sprintf("  - %s: %s\n", code, resp.Description))
				}
			}
		}
	}

	return details.String()
}

func formatSchemaDetails(schema *base.SchemaProxy) string {
	var details strings.Builder

	if schema == nil || schema.Schema() == nil {
		return "No schema details available"
	}

	s := schema.Schema()

	// Handle both single type (OpenAPI 3.0) and array of types (OpenAPI 3.1)
	if len(s.Type) > 0 {
		if len(s.Type) == 1 {
			details.WriteString(fmt.Sprintf("Type: %s\n", s.Type[0]))
		} else {
			details.WriteString(fmt.Sprintf("Types: %v\n", s.Type))
		}
	}

	if s.Format != "" {
		details.WriteString(fmt.Sprintf("Format: %s\n", s.Format))
	}

	if len(s.Required) > 0 {
		details.WriteString(fmt.Sprintf("Required: %v\n", s.Required))
	}

	if s.Properties != nil && s.Properties.Len() > 0 {
		details.WriteString("Properties:\n")

		// Get property names and sort them for stable ordering
		var propNames []string
		for pair := s.Properties.First(); pair != nil; pair = pair.Next() {
			propNames = append(propNames, pair.Key())
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			if prop, ok := s.Properties.Get(propName); ok && prop != nil && prop.Schema() != nil {
				propType := "unknown"
				if len(prop.Schema().Type) > 0 {
					if len(prop.Schema().Type) == 1 {
						propType = prop.Schema().Type[0]
					} else {
						propType = fmt.Sprintf("%v", prop.Schema().Type)
					}
				}
				details.WriteString(fmt.Sprintf("  - %s: %s\n", propName, propType))
			}
		}
	}

	if s.Items != nil && s.Items.A != nil && s.Items.A.Schema() != nil && len(s.Items.A.Schema().Type) > 0 {
		itemsType := s.Items.A.Schema().Type
		if len(itemsType) == 1 {
			details.WriteString(fmt.Sprintf("Items Type: %s\n", itemsType[0]))
		} else {
			details.WriteString(fmt.Sprintf("Items Types: %v\n", itemsType))
		}
	}

	return details.String()
}

func formatRequestBodyDetails(reqBody *v3.RequestBody) string {
	var details strings.Builder

	if reqBody == nil {
		return "No request body details available"
	}

	if reqBody.Required != nil && *reqBody.Required {
		details.WriteString("Required: true\n")
	}

	if reqBody.Content != nil && reqBody.Content.Len() > 0 {
		details.WriteString("Content Types:\n")

		// Get media types and sort them for stable ordering
		var mediaTypes []string
		for pair := reqBody.Content.First(); pair != nil; pair = pair.Next() {
			mediaTypes = append(mediaTypes, pair.Key())
		}
		sort.Strings(mediaTypes)

		for _, mediaType := range mediaTypes {
			if mediaTypeObj, ok := reqBody.Content.Get(mediaType); ok && mediaTypeObj != nil {
				details.WriteString(fmt.Sprintf("  - %s", mediaType))
				if mediaTypeObj.Schema != nil && mediaTypeObj.Schema.Schema() != nil && len(mediaTypeObj.Schema.Schema().Type) > 0 {
					types := mediaTypeObj.Schema.Schema().Type
					if len(types) == 1 {
						details.WriteString(fmt.Sprintf(" (type: %s)", types[0]))
					} else {
						details.WriteString(fmt.Sprintf(" (types: %v)", types))
					}
				}
				details.WriteString("\n")
			}
		}
	}

	return details.String()
}

func formatResponseDetails(response *v3.Response) string {
	var details strings.Builder

	if response == nil {
		return "No response details available"
	}

	if response.Content != nil && response.Content.Len() > 0 {
		details.WriteString("Content Types:\n")

		// Get media types and sort them for stable ordering
		var mediaTypes []string
		for pair := response.Content.First(); pair != nil; pair = pair.Next() {
			mediaTypes = append(mediaTypes, pair.Key())
		}
		sort.Strings(mediaTypes)

		for _, mediaType := range mediaTypes {
			if mediaTypeObj, ok := response.Content.Get(mediaType); ok && mediaTypeObj != nil {
				details.WriteString(fmt.Sprintf("  - %s", mediaType))
				if mediaTypeObj.Schema != nil && mediaTypeObj.Schema.Schema() != nil && len(mediaTypeObj.Schema.Schema().Type) > 0 {
					types := mediaTypeObj.Schema.Schema().Type
					if len(types) == 1 {
						details.WriteString(fmt.Sprintf(" (type: %s)", types[0]))
					} else {
						details.WriteString(fmt.Sprintf(" (types: %v)", types))
					}
				}
				details.WriteString("\n")
			}
		}
	}

	if response.Headers != nil && response.Headers.Len() > 0 {
		details.WriteString("Headers:\n")

		// Get header names and sort them for stable ordering
		var headerNames []string
		for pair := response.Headers.First(); pair != nil; pair = pair.Next() {
			headerNames = append(headerNames, pair.Key())
		}
		sort.Strings(headerNames)

		for _, headerName := range headerNames {
			details.WriteString(fmt.Sprintf("  - %s\n", headerName))
		}
	}

	return details.String()
}

func formatParameterDetails(param *v3.Parameter) string {
	var details strings.Builder

	if param == nil {
		return "No parameter details available"
	}

	details.WriteString(fmt.Sprintf("In: %s\n", param.In))

	if param.Required != nil && *param.Required {
		details.WriteString("Required: true\n")
	}

	if param.Schema != nil && param.Schema.Schema() != nil && len(param.Schema.Schema().Type) > 0 {
		types := param.Schema.Schema().Type
		if len(types) == 1 {
			details.WriteString(fmt.Sprintf("Type: %s\n", types[0]))
		} else {
			details.WriteString(fmt.Sprintf("Types: %v\n", types))
		}
		if param.Schema.Schema().Format != "" {
			details.WriteString(fmt.Sprintf("Format: %s\n", param.Schema.Schema().Format))
		}
	}

	if param.Example != nil {
		details.WriteString(fmt.Sprintf("Example: %v\n", param.Example))
	}

	return details.String()
}

func formatHeaderDetails(header *v3.Header) string {
	var details strings.Builder

	if header == nil {
		return "No header details available"
	}

	if header.Required {
		details.WriteString("Required: true\n")
	}

	if header.Schema != nil && header.Schema.Schema() != nil && len(header.Schema.Schema().Type) > 0 {
		types := header.Schema.Schema().Type
		if len(types) == 1 {
			details.WriteString(fmt.Sprintf("Type: %s\n", types[0]))
		} else {
			details.WriteString(fmt.Sprintf("Types: %v\n", types))
		}
		if header.Schema.Schema().Format != "" {
			details.WriteString(fmt.Sprintf("Format: %s\n", header.Schema.Schema().Format))
		}
	}

	return details.String()
}

func formatSecuritySchemeDetails(name string, secScheme *v3.SecurityScheme) string {
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

func formatWebhookDetails(hook webhook) string {
	var details strings.Builder

	if hook.op.Summary != "" {
		details.WriteString(fmt.Sprintf("Summary: %s\n", hook.op.Summary))
	}

	if hook.op.Description != "" {
		details.WriteString(fmt.Sprintf("Description: %s\n", hook.op.Description))
	}

	if hook.op.OperationId != "" {
		details.WriteString(fmt.Sprintf("Operation ID: %s\n", hook.op.OperationId))
	}

	return details.String()
}
