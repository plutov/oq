package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// convertOpenAPI31To30 converts an OpenAPI 3.1 spec to OpenAPI 3.0 compatible format
// This should be rrmoved once kin-openapi has full support for OpenAPI 3.1
func convertOpenAPI31To30(content []byte) ([]byte, error) {
	var spec map[string]interface{}

	// Try parsing as JSON first
	err := json.Unmarshal(content, &spec)
	if err != nil {
		// If JSON parsing fails, try YAML
		err = yaml.Unmarshal(content, &spec)
		if err != nil {
			return content, fmt.Errorf("failed to parse as JSON or YAML: %v", err)
		}
	}

	// Check if this is OpenAPI 3.1
	if version, ok := spec["openapi"].(string); !ok || !strings.HasPrefix(version, "3.1") {
		// Not 3.1, return as-is
		return content, nil
	}

	// Process the spec recursively
	processNode(spec)

	// Downgrade the version to 3.0.x to reflect the conversion
	spec["openapi"] = "3.0.3"

	// Convert back to JSON (kin-openapi works better with JSON)
	converted, err := json.Marshal(spec)
	if err != nil {
		return content, fmt.Errorf("failed to marshal converted spec: %v", err)
	}

	return converted, nil
}

// processNode recursively processes a node in the OpenAPI spec
func processNode(node interface{}) {
	switch v := node.(type) {
	case map[string]interface{}:
		processMap(v)
	case []interface{}:
		for _, item := range v {
			processNode(item)
		}
	}
}

// processMap processes a map node, handling OpenAPI 3.1 specific conversions
func processMap(m map[string]interface{}) {
	// Convert examples to example (take first example)
	if examples, ok := m["examples"].([]interface{}); ok && len(examples) > 0 {
		m["example"] = examples[0]
		delete(m, "examples")
	}

	// Convert exclusiveMinimum from number to boolean
	if exclusiveMin, ok := m["exclusiveMinimum"]; ok {
		switch val := exclusiveMin.(type) {
		case float64:
			// Set minimum to the exclusive value and exclusiveMinimum to true
			m["minimum"] = val
			m["exclusiveMinimum"] = true
		case int:
			m["minimum"] = float64(val)
			m["exclusiveMinimum"] = true
		}
	}

	// Convert exclusiveMaximum from number to boolean
	if exclusiveMax, ok := m["exclusiveMaximum"]; ok {
		switch val := exclusiveMax.(type) {
		case float64:
			// Set maximum to the exclusive value and exclusiveMaximum to true
			m["maximum"] = val
			m["exclusiveMaximum"] = true
		case int:
			m["maximum"] = float64(val)
			m["exclusiveMaximum"] = true
		}
	}

	// Handle schema type conversion (3.1 allows array of types, 3.0 only allows string)
	if schemaType, ok := m["type"]; ok {
		switch val := schemaType.(type) {
		case []interface{}:
			// Handle array of types for OpenAPI 3.0 compatibility
			if len(val) > 0 {
				var nonNullTypes []string
				var hasNull bool

				// Check for null type and collect non-null types
				for _, typeVal := range val {
					if typeStr, ok := typeVal.(string); ok {
						if typeStr == "null" {
							hasNull = true
						} else {
							nonNullTypes = append(nonNullTypes, typeStr)
						}
					}
				}

				// Set the type to the first non-null type (or null if that's all we have)
				if len(nonNullTypes) > 0 {
					m["type"] = nonNullTypes[0]
					// If there was a null type, mark as nullable for OpenAPI 3.0
					if hasNull {
						m["nullable"] = true
					}
				} else if hasNull {
					// Only null type - this is unusual but handle it
					delete(m, "type")
					m["nullable"] = true
				}
			}
		case string:
			// Handle standalone null type
			if val == "null" {
				delete(m, "type")
				m["nullable"] = true
			}
		}
	}

	// Remove OpenAPI 3.1 specific fields that are not supported in 3.0
	openapi31Fields := []string{
		"$schema",
		"unevaluatedProperties",
		"unevaluatedItems",
		"prefixItems",
		"contains",
		"minContains",
		"maxContains",
		"dependentRequired",
		"dependentSchemas",
		"patternProperties",
		"propertyNames",
		"if",
		"then",
		"else",
		"webhooks",         // OpenAPI 3.1 feature not supported in 3.0
		"contentEncoding",  // JSON Schema feature not supported in OpenAPI 3.0
		"contentMediaType", // JSON Schema feature not supported in OpenAPI 3.0
		"contentSchema",    // JSON Schema feature not supported in OpenAPI 3.0
	}
	for _, field := range openapi31Fields {
		delete(m, field)
	}

	// Handle const -> enum conversion
	if constVal, ok := m["const"]; ok {
		m["enum"] = []interface{}{constVal}
		delete(m, "const")
	}

	// Handle license.identifier -> license.url mapping
	if license, ok := m["license"].(map[string]interface{}); ok {
		if identifier, ok := license["identifier"].(string); ok {
			// Convert common license identifiers to URLs
			licenseURLs := map[string]string{
				"MIT":             "https://opensource.org/licenses/MIT",
				"Apache-2.0":      "https://www.apache.org/licenses/LICENSE-2.0.html",
				"CC-BY-NC-SA-4.0": "https://creativecommons.org/licenses/by-nc-sa/4.0/",
				"GPL-3.0":         "https://www.gnu.org/licenses/gpl-3.0.html",
				"BSD-3-Clause":    "https://opensource.org/licenses/BSD-3-Clause",
			}

			if url, exists := licenseURLs[identifier]; exists {
				license["url"] = url
			} else {
				// Fallback: create a generic SPDX URL
				license["url"] = "https://spdx.org/licenses/" + identifier + ".html"
			}
			delete(license, "identifier")
		}
	}

	// Handle oneOf/anyOf/allOf with type constraints
	for _, combiner := range []string{"oneOf", "anyOf", "allOf"} {
		if combinerVal, ok := m[combiner].([]interface{}); ok {
			for _, item := range combinerVal {
				processNode(item)
			}
		}
	}

	// Process all nested objects
	for _, value := range m {
		processNode(value)
	}
}
