package hclbuilder

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Builder provides a fluent API for building and modifying HCL configurations
type Builder struct {
	file             *hclwrite.File
	ProviderType     ProviderType
	ProviderProperty ProviderType
	addedBuilders    map[*Builder]bool
}

// New creates a new empty HCL builder
func New() *Builder {
	return &Builder{
		file:          hclwrite.NewEmptyFile(),
		addedBuilders: make(map[*Builder]bool),
	}
}

// NewWithProvider creates a new builder with a provider block
func NewWithProvider(provider ProviderType, serverURL string) *Builder {
	return New().WithProvider(provider, serverURL)
}

// FromFile loads an HCL configuration from a file
func FromFile(path string) (*Builder, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	file, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing HCL: %s", diags.Error())
	}

	return &Builder{
		file:          file,
		addedBuilders: make(map[*Builder]bool),
	}, nil
}

// FromString parses an HCL configuration from a string
func FromString(content string) (*Builder, error) {
	file, diags := hclwrite.ParseConfig([]byte(content), "<string>", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing HCL: %s", diags.Error())
	}

	return &Builder{
		file:          file,
		addedBuilders: make(map[*Builder]bool),
	}, nil
}

// Build returns the HCL configuration as a string
func (b *Builder) Build() string {
	return string(b.file.Bytes())
}

// NewWithProvider creates a new builder with a provider block
func (b *Builder) WithProvider(provider ProviderType, serverURL string) *Builder {
	b.ProviderType = provider

	if serverURL == "" {
		serverURL = "http://localhost:5681"
	}

	b.SetBlock(fmt.Sprintf("provider.%s", provider), map[string]any{
		"server_url": serverURL,
	})

	return b
}

// Add embeds another builder's content into this builder
func (b *Builder) Add(other *Builder) *Builder {
	if other == nil || other.file == nil {
		return b
	}

	// Check if this builder has already been added
	if b.addedBuilders[other] {
		return b
	}

	// Mark this builder as added
	b.addedBuilders[other] = true

	// Merge all blocks from the other builder into this one
	for _, block := range other.file.Body().Blocks() {
		b.file.Body().AppendBlock(block)
	}

	// Merge all attributes from the other builder into this one
	for name, attr := range other.file.Body().Attributes() {
		b.file.Body().SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
	}

	return b
}

// WriteFile writes the HCL configuration to a file
func (b *Builder) WriteFile(path string) error {
	return os.WriteFile(path, b.file.Bytes(), 0o600)
}

// AddAttribute adds or updates an attribute on the first block in this builder.
// This is designed for builders that contain a single resource/block.
// The path uses dot notation for nested attributes.
// Value can be a Go value or a string containing HCL expression.
// Example: builder.AddAttribute("skip_creating_initial_policies", `["*"]`)
// Example: builder.AddAttribute("routing.default_forbid_mesh_external_service_access", "true")
// Example: builder.AddAttribute("constraints.dataplane_proxy.requirements", `[{ tags = { key = "a" } }]`)
func (b *Builder) AddAttribute(path string, value any) *Builder {
	blocks := b.file.Body().Blocks()
	if len(blocks) == 0 {
		// No blocks, can't add attribute
		return b
	}

	// Work with the first block (typical use case: single resource)
	block := blocks[0]
	parts := strings.Split(path, ".")

	// If value is a string, try to parse it as HCL
	if strValue, ok := value.(string); ok {
		parsedValue := parseHCLValue(strValue)
		if parsedValue != nil {
			value = parsedValue
		}
	}

	if len(parts) == 1 {
		// Simple attribute
		block.Body().SetAttributeValue(path, convertToCtyValue(value))
	} else {
		// Nested attribute - need to merge with existing value
		rootAttr := parts[0]

		// Get existing value if present
		var existingValue any
		if attr := block.Body().GetAttribute(rootAttr); attr != nil {
			// Convert the expression tokens to string and parse it
			exprTokens := attr.Expr().BuildTokens(nil)
			exprStr := string(exprTokens.Bytes())
			existingValue = parseHCLValue(exprStr)
		}

		// Build the new nested structure
		newNested := buildNestedStructureRecursive(parts[1:], value)

		// Merge with existing value if it's a map
		var finalValue any
		if existingMap, ok := existingValue.(map[string]any); ok {
			finalValue = deepMerge(existingMap, newNested.(map[string]any))
		} else {
			finalValue = newNested
		}

		// Set the root attribute with the merged structure
		block.Body().SetAttributeValue(rootAttr, convertToCtyValue(finalValue))
	}

	return b
}

// RemoveAttribute removes an attribute from the first block in this builder.
// Uses dot notation for nested attributes.
func (b *Builder) RemoveAttribute(path string) *Builder {
	blocks := b.file.Body().Blocks()
	if len(blocks) == 0 {
		return b
	}

	block := blocks[0]
	parts := strings.Split(path, ".")

	if len(parts) == 1 {
		// Simple attribute
		block.Body().RemoveAttribute(path)
	} else {
		// For nested attributes, we'd need to read the structure, modify it, and write it back
		// For now, just remove the top-level attribute
		block.Body().RemoveAttribute(parts[0])
	}

	return b
}

// parseHCLValue attempts to parse a string as an HCL expression and return its Go value
func parseHCLValue(hclExpr string) any {
	// Wrap in a dummy attribute to make it valid HCL
	wrapped := fmt.Sprintf("dummy = %s", hclExpr)

	// Parse using hclparse to evaluate the expression
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(wrapped), "<inline>")
	if diags.HasErrors() {
		return nil
	}

	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return nil
	}

	if dummyAttr, ok := attrs["dummy"]; ok {
		val, diags := dummyAttr.Expr.Value(nil) // nil context for static evaluation
		if diags.HasErrors() {
			return nil
		}
		return convertCtyToGo(val)
	}

	return nil
}

// convertCtyToGo converts a cty.Value to a Go value
func convertCtyToGo(val cty.Value) any {
	if val.IsNull() {
		return nil
	}

	ty := val.Type()

	switch {
	case ty == cty.String:
		return val.AsString()
	case ty == cty.Number:
		var f float64
		_ = gocty.FromCtyValue(val, &f)
		// Check if it's actually an integer
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	case ty == cty.Bool:
		return val.True()
	case ty.IsListType() || ty.IsTupleType():
		var result []any
		it := val.ElementIterator()
		for it.Next() {
			_, elemVal := it.Element()
			result = append(result, convertCtyToGo(elemVal))
		}
		return result
	case ty.IsMapType() || ty.IsObjectType():
		result := make(map[string]any)
		it := val.ElementIterator()
		for it.Next() {
			keyVal, elemVal := it.Element()
			key := keyVal.AsString()
			result[key] = convertCtyToGo(elemVal)
		}
		return result
	default:
		return nil
	}
}

// deepMerge recursively merges two maps, with values from 'b' taking precedence
func deepMerge(a, b map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy all keys from 'a'
	for k, v := range a {
		result[k] = v
	}

	// Merge keys from 'b'
	for k, v := range b {
		if existingV, exists := result[k]; exists {
			// If both values are maps, merge them recursively
			if existingMap, ok := existingV.(map[string]any); ok {
				if newMap, ok := v.(map[string]any); ok {
					result[k] = deepMerge(existingMap, newMap)
					continue
				}
			}
		}
		// Otherwise, value from 'b' takes precedence
		result[k] = v
	}

	return result
}

// buildNestedStructure builds a nested map from dot-separated path and value

func buildNestedStructureRecursive(parts []string, value any) any {
	if len(parts) == 1 {
		return map[string]any{parts[0]: value}
	}
	return map[string]any{
		parts[0]: buildNestedStructureRecursive(parts[1:], value),
	}
}

// SetAttribute sets an attribute value at the given path.
//
// Path format: "block_type.block_label.attribute_name" or nested paths.
// Example: "variable.name.default" sets variable "name" { default = value }.
//
// If the path is invalid (fewer than 3 parts), this method does nothing.
func (b *Builder) SetAttribute(path string, value any) {
	parts := strings.Split(path, ".")
	if len(parts) < 3 {
		// Need at least: block_type.block_label.attribute_name
		return
	}

	body := b.file.Body()
	attributeName := parts[len(parts)-1]

	// Navigate/create the block structure (all parts except the last, which is the attribute)
	for i := 0; i < len(parts)-1; i += 2 {
		if i+1 >= len(parts)-1 {
			// We've reached the end, shouldn't happen with proper path
			break
		}

		blockType := parts[i]
		blockLabel := parts[i+1]

		// Check if this is the last block before the attribute
		if i+2 == len(parts)-1 {
			// This is the block that should contain the attribute
			block := findOrCreateBlock(body, blockType, []string{blockLabel})
			block.Body().SetAttributeValue(attributeName, convertToCtyValue(value))
			return
		}

		// Navigate deeper into nested blocks
		block := findOrCreateBlock(body, blockType, []string{blockLabel})
		body = block.Body()
	}
}

// SetBlock creates or replaces a block with the given attributes.
//
// Path format: "block_type.block_label1.block_label2...".
// Example: "resource.aws_instance.web".
//
// If the path is invalid (fewer than 2 parts), this method does nothing.
// Nested maps are treated as nested blocks.
func (b *Builder) SetBlock(path string, attributes map[string]any) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return
	}

	blockType := parts[0]
	labels := parts[1:]

	body := b.file.Body()

	// Remove existing block if it exists
	removeBlock(body, blockType, labels)

	// Create new block
	block := body.AppendNewBlock(blockType, labels)
	setBlockAttributes(block.Body(), attributes)
}

// RemoveBlock removes a block at the given path.
//
// If the path is invalid or the block doesn't exist, this method does nothing.
func (b *Builder) RemoveBlock(path string) *Builder {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return b
	}

	blockType := parts[0]
	labels := parts[1:]

	removeBlock(b.file.Body(), blockType, labels)

	return b
}

// Helper functions

func findBlock(body *hclwrite.Body, blockType string, labels []string) *hclwrite.Block {
	for _, block := range body.Blocks() {
		if block.Type() == blockType && matchLabels(block.Labels(), labels) {
			return block
		}
	}
	return nil
}

func findOrCreateBlock(body *hclwrite.Body, blockType string, labels []string) *hclwrite.Block {
	block := findBlock(body, blockType, labels)
	if block == nil {
		block = body.AppendNewBlock(blockType, labels)
	}
	return block
}

func removeBlock(body *hclwrite.Body, blockType string, labels []string) {
	block := findBlock(body, blockType, labels)
	if block != nil {
		body.RemoveBlock(block)
	}
}

func matchLabels(blockLabels, targetLabels []string) bool {
	if len(blockLabels) != len(targetLabels) {
		return false
	}
	for i := range blockLabels {
		if blockLabels[i] != targetLabels[i] {
			return false
		}
	}
	return true
}

func setBlockAttributes(body *hclwrite.Body, attributes map[string]any) {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := attributes[key]
		// All values are set as attributes (including maps which become object values)
		// Nested blocks must be created explicitly via SetBlock, not through this function
		body.SetAttributeValue(key, convertToCtyValue(value))
	}
}

func convertToCtyValue(value any) cty.Value {
	switch v := value.(type) {
	case string:
		return cty.StringVal(v)
	case int:
		return cty.NumberIntVal(int64(v))
	case int64:
		return cty.NumberIntVal(v)
	case float64:
		return cty.NumberFloatVal(v)
	case bool:
		return cty.BoolVal(v)
	case []string:
		if len(v) == 0 {
			return cty.ListValEmpty(cty.String)
		}
		vals := make([]cty.Value, len(v))
		for i, s := range v {
			vals[i] = cty.StringVal(s)
		}
		return cty.ListVal(vals)
	case []any:
		if len(v) == 0 {
			return cty.EmptyTupleVal
		}
		vals := make([]cty.Value, len(v))
		for i, item := range v {
			vals[i] = convertToCtyValue(item)
		}
		return cty.ListVal(vals)
	case map[string]any:
		vals := make(map[string]cty.Value)
		for k, item := range v {
			vals[k] = convertToCtyValue(item)
		}
		return cty.ObjectVal(vals)
	default:
		return cty.StringVal(fmt.Sprintf("%v", v))
	}
}

// Provider-specific helper methods

// ResourceAddress returns the Terraform resource address for provider-specific resources
func (b *Builder) ResourceAddress(resourceType, resourceName string) string {
	if b.ProviderType == "" {
		return fmt.Sprintf("%s.%s", resourceType, resourceName)
	}
	return fmt.Sprintf("%s_%s.%s", b.ProviderType, resourceType, resourceName)
}

// RemoveMesh removes a mesh resource
func (b *Builder) RemoveMesh(meshResourceName string) *Builder {
	return b.RemoveBlock(fmt.Sprintf("resource.%s_mesh.%s", b.ProviderType, meshResourceName))
}

// AddControlPlane adds a mesh control plane resource (for Konnect providers)
func (b *Builder) AddControlPlane(resourceName, name, description string) *Builder {
	attrs := map[string]any{
		"name":        name,
		"description": description,
	}

	b.SetBlock(fmt.Sprintf("resource.%s_mesh_control_plane.%s", b.ProviderProperty, resourceName), attrs)
	return b
}

// AddPolicy adds a policy resource
func (b *Builder) AddPolicy(policyType, policyName, policyResourceName, meshRef string, spec map[string]any) *Builder {
	// Convert policy type from snake_case to PascalCase for the type attribute
	pascalCaseType := resourceTypeToPolicyType(policyType)

	attrs := map[string]any{
		"provider": strings.ToLower(string(b.ProviderProperty)),
		"type":     pascalCaseType,
		"name":     policyName,
		"mesh":     meshRef,
	}

	// Merge spec into attrs
	for k, v := range spec {
		attrs[k] = v
	}

	b.SetBlock(fmt.Sprintf("resource.%s_%s.%s", b.ProviderProperty, policyType, policyResourceName), attrs)
	return b
}

func (b *Builder) ResourceName() string {
	blocks := b.file.Body().Blocks()
	if len(blocks) == 0 {
		return ""
	}

	block := blocks[0]
	labels := block.Labels()
	if len(labels) == 0 {
		return ""
	}

	// For resources, the first label is the resource type, second is the name
	// e.g., resource "konnect_mesh_control_plane" "my_meshcontrolplane"
	// labels[0] = "konnect_mesh_control_plane", labels[1] = "my_meshcontrolplane"
	if len(labels) >= 2 {
		return labels[1]
	}

	return labels[0]
}

// Helper functions for policy type conversion

func resourceTypeToPolicyType(resourceType string) string {
	// Special case: mesh_secret -> Secret (not MeshSecret)
	if resourceType == "mesh_secret" {
		return "Secret"
	}
	// Convert "mesh_traffic_permission" to "MeshTrafficPermission"
	parts := strings.Split(resourceType, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return result
}
