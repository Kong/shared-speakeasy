package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type TemplateParams struct {
	ResourceName      string // e.g., MeshTrafficPermission
	ResourceVarName   string // e.g., meshTrafficPermission
	ResourceModelName string // e.g., MeshTrafficPermissionResourceModel
	ProviderName      string // e.g., terraform-provider-kong-mesh
}

func toLowerCamel(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run main.go <templatePath> <outputPath or -> <ResourceName> <ProviderName>")
		os.Exit(1)
	}

	templatePath := os.Args[1]
	outputPath := os.Args[2]
	resourceName := os.Args[3]
	providerName := os.Args[4]

	params := TemplateParams{
		ResourceName:      resourceName,
		ResourceVarName:   toLowerCamel(resourceName),
		ResourceModelName: resourceName + "ResourceModel",
		ProviderName:      providerName,
	}

	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading template file: %v\n", err)
		os.Exit(1)
	}

	tmpl, err := template.New("planmod").Parse(string(tmplContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	var output *os.File
	if outputPath == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer output.Close()
	}

	if err := tmpl.Execute(output, params); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	if outputPath != "-" {
		fmt.Printf("Generated plan modification code at: %s\n", outputPath)
	}
}
