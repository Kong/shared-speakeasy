package hclbuilder

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// Builder provides a fluent API for building and modifying HCL configurations
type Builder struct {
	file *hclwrite.File
}

// New creates a new empty HCL builder
func New() *Builder {
	return &Builder{
		file: hclwrite.NewEmptyFile(),
	}
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

	return &Builder{file: file}, nil
}

// FromString parses an HCL configuration from a string
func FromString(content string) (*Builder, error) {
	file, diags := hclwrite.ParseConfig([]byte(content), "<string>", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing HCL: %s", diags.Error())
	}

	return &Builder{file: file}, nil
}

// Build returns the HCL configuration as a string
func (b *Builder) Build() string {
	return string(b.file.Bytes())
}

// WriteFile writes the HCL configuration to a file
func (b *Builder) WriteFile(path string) error {
	return os.WriteFile(path, b.file.Bytes(), 0o600)
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

// RemoveAttribute removes an attribute at the given path.
//
// If the path is invalid or the attribute doesn't exist, this method does nothing.
func (b *Builder) RemoveAttribute(path string) {
	parts := strings.Split(path, ".")
	if len(parts) < 3 {
		// Need at least: block_type.block_label.attribute_name
		return
	}

	body := b.file.Body()
	attributeName := parts[len(parts)-1]

	// Navigate to the block containing the attribute
	for i := 0; i < len(parts)-1; i += 2 {
		if i+1 >= len(parts)-1 {
			// We've reached the end
			break
		}

		blockType := parts[i]
		blockLabel := parts[i+1]

		// Check if this is the last block before the attribute
		if i+2 == len(parts)-1 {
			// This is the block that contains the attribute
			block := findBlock(body, blockType, []string{blockLabel})
			if block != nil {
				block.Body().RemoveAttribute(attributeName)
			}
			return
		}

		// Navigate deeper
		block := findBlock(body, blockType, []string{blockLabel})
		if block == nil {
			return
		}
		body = block.Body()
	}
}

// RemoveBlock removes a block at the given path.
//
// If the path is invalid or the block doesn't exist, this method does nothing.
func (b *Builder) RemoveBlock(path string) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return
	}

	blockType := parts[0]
	labels := parts[1:]

	removeBlock(b.file.Body(), blockType, labels)
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
