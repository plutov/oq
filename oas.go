package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
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

	// Sort endpoints for stable ordering: first by path, then by method
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].path != endpoints[j].path {
			return endpoints[i].path < endpoints[j].path
		}
		return endpoints[i].method < endpoints[j].method
	})

	return endpoints
}

func extractWebhooks(doc *openapi3.T) []webhook {
	var webhooks []webhook

	// Check if this is OpenAPI 3.1+ first
	if !isOpenAPI31OrLater(doc) {
		return webhooks
	}

	// Check in Extensions first
	if webhookData, ok := doc.Extensions["webhooks"].(map[string]interface{}); ok {
		webhooks = parseWebhooksFromData(webhookData)
	}

	return webhooks
}

func parseWebhooksFromData(webhookData map[string]interface{}) []webhook {
	var webhooks []webhook

	for name, hookData := range webhookData {
		if hookMap, ok := hookData.(map[string]interface{}); ok {
			// Look for HTTP methods in the webhook
			for method, methodData := range hookMap {
				if isHTTPMethod(method) {
					if methodMap, ok := methodData.(map[string]interface{}); ok {
						// Create a mock operation from the webhook data
						op := &openapi3.Operation{}
						if summary, ok := methodMap["summary"].(string); ok {
							op.Summary = summary
						}
						if description, ok := methodMap["description"].(string); ok {
							op.Description = description
						}
						if operationId, ok := methodMap["operationId"].(string); ok {
							op.OperationID = operationId
						}

						webhooks = append(webhooks, webhook{
							name:   name,
							method: strings.ToUpper(method),
							op:     op,
							folded: true,
						})
					}
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

func isOpenAPI31OrLater(doc *openapi3.T) bool {
	if doc == nil || doc.OpenAPI == "" {
		return false
	}

	// Extract major and minor version
	parts := strings.Split(doc.OpenAPI, ".")
	if len(parts) < 2 {
		return false
	}

	major := parts[0]
	minor := parts[1]

	// Check for version 3.1 or later
	return major == "3" && (minor == "1" || minor > "1") || major > "3"
}

func isHTTPMethod(method string) bool {
	httpMethods := map[string]bool{
		"get":     true,
		"post":    true,
		"put":     true,
		"delete":  true,
		"patch":   true,
		"head":    true,
		"options": true,
		"trace":   true,
	}
	return httpMethods[strings.ToLower(method)]
}

func extractComponents(doc *openapi3.T) []component {
	var components []component

	if doc.Components != nil {
		if doc.Components.Schemas != nil {
			for name, schema := range doc.Components.Schemas {
				details := formatSchemaDetails(schema.Value)
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
				details := formatRequestBodyDetails(reqBody.Value)
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
				details := formatResponseDetails(resp.Value)
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
				details := formatParameterDetails(param.Value)
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
				details := formatHeaderDetails(header.Value)
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
			if param.Value != nil {
				details.WriteString(fmt.Sprintf("  - %s (%s): %s\n",
					param.Value.Name, param.Value.In, param.Value.Description))
			}
		}
	}

	if ep.op.RequestBody != nil && ep.op.RequestBody.Value != nil {
		details.WriteString("Request Body:\n")

		// Get media types and sort them for stable ordering
		var mediaTypes []string
		for mediaType := range ep.op.RequestBody.Value.Content {
			mediaTypes = append(mediaTypes, mediaType)
		}
		sort.Strings(mediaTypes)

		for _, mediaType := range mediaTypes {
			details.WriteString(fmt.Sprintf("  - %s\n", mediaType))
		}
	}

	if ep.op.Responses != nil {
		details.WriteString("Responses:\n")

		// Get response codes and sort them for stable ordering
		var codes []string
		for code := range ep.op.Responses.Map() {
			codes = append(codes, code)
		}

		// Sort by status code numerically, then alphabetically for non-numeric codes
		sortResponseCodes(codes)

		for _, code := range codes {
			resp := ep.op.Responses.Map()[code]
			if resp.Value != nil && resp.Value.Description != nil {
				details.WriteString(fmt.Sprintf("  - %s: %s\n", code, *resp.Value.Description))
			}
		}
	}

	return details.String()
}

func formatSchemaDetails(schema *openapi3.Schema) string {
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

		// Get property names and sort them for stable ordering
		var propNames []string
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			prop := schema.Properties[propName]
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

func formatRequestBodyDetails(reqBody *openapi3.RequestBody) string {
	var details strings.Builder

	if reqBody == nil {
		return "No request body details available"
	}

	if reqBody.Required {
		details.WriteString("Required: true\n")
	}

	if len(reqBody.Content) > 0 {
		details.WriteString("Content Types:\n")

		// Get media types and sort them for stable ordering
		var mediaTypes []string
		for mediaType := range reqBody.Content {
			mediaTypes = append(mediaTypes, mediaType)
		}
		sort.Strings(mediaTypes)

		for _, mediaType := range mediaTypes {
			mediaTypeObj := reqBody.Content[mediaType]
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

func formatResponseDetails(response *openapi3.Response) string {
	var details strings.Builder

	if response == nil {
		return "No response details available"
	}

	if len(response.Content) > 0 {
		details.WriteString("Content Types:\n")

		// Get media types and sort them for stable ordering
		var mediaTypes []string
		for mediaType := range response.Content {
			mediaTypes = append(mediaTypes, mediaType)
		}
		sort.Strings(mediaTypes)

		for _, mediaType := range mediaTypes {
			mediaTypeObj := response.Content[mediaType]
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

		// Get header names and sort them for stable ordering
		var headerNames []string
		for headerName := range response.Headers {
			headerNames = append(headerNames, headerName)
		}
		sort.Strings(headerNames)

		for _, headerName := range headerNames {
			details.WriteString(fmt.Sprintf("  - %s\n", headerName))
		}
	}

	return details.String()
}

func formatParameterDetails(param *openapi3.Parameter) string {
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

func formatHeaderDetails(header *openapi3.Header) string {
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

func formatWebhookDetails(hook webhook) string {
	var details strings.Builder

	if hook.op.Summary != "" {
		details.WriteString(fmt.Sprintf("Summary: %s\n", hook.op.Summary))
	}

	if hook.op.Description != "" {
		details.WriteString(fmt.Sprintf("Description: %s\n", hook.op.Description))
	}

	if hook.op.OperationID != "" {
		details.WriteString(fmt.Sprintf("Operation ID: %s\n", hook.op.OperationID))
	}

	return details.String()
}
