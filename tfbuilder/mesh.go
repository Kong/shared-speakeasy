package tfbuilder

import (
	"bytes"
	"fmt"
	"text/template"
)

type MeshBuilder struct {
	ResourceName string
	MeshName     string
	DependsOn    []string
	Spec         string
	CPID         string // Optional
}

func NewMeshBuilder(resourceName, meshName string) *MeshBuilder {
	return &MeshBuilder{
		ResourceName: resourceName,
		MeshName:     meshName,
	}
}

func (m *MeshBuilder) WithSpec(spec string) *MeshBuilder {
	m.Spec = spec
	return m
}

func (m *MeshBuilder) WithCPID(cpID string) *MeshBuilder {
	m.CPID = cpID
	return m
}

func (m *MeshBuilder) WithDependsOn(deps ...string) *MeshBuilder {
	m.DependsOn = append(m.DependsOn, deps...)
	return m
}

func (m *MeshBuilder) AddToSpec(after, newLine string) *MeshBuilder {
	m.Spec = addToSpec(m.Spec, after, newLine)
	return m
}

func (m *MeshBuilder) RemoveFromSpec(match string) *MeshBuilder {
	m.Spec = removeFromSpec(m.Spec, match)
	return m
}

func (m *MeshBuilder) UpdateSpec(match, newValue string) *MeshBuilder {
	m.Spec = updateSpec(m.Spec, match, newValue)
	return m
}

func (m *MeshBuilder) Render(provider *Builder) string {
	data := map[string]interface{}{
		"Provider":         provider.provider,
		"ProviderProperty": provider.providerProperty,
		"ResourceName":     m.ResourceName,
		"MeshName":         m.MeshName,
		"DependsOn":        m.DependsOn,
		"Spec":             m.Spec,
		"CPID":             m.CPID,
	}

	tmplBytes, err := templatesFS.ReadFile("templates/mesh.tmpl")
	if err != nil {
		panic(fmt.Errorf("failed to read mesh template: %w", err))
	}

	tmpl, err := template.New("mesh").Parse(string(tmplBytes))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	return buf.String()
}
