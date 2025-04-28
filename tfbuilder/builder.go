package tfbuilder

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type ProviderType string

const (
	KongMesh    ProviderType = "kong-mesh"
	Konnect     ProviderType = "konnect"
	KonnectBeta ProviderType = "konnect-beta"
)

type Builder struct {
	provider         ProviderType
	providerProperty ProviderType // optional
	scheme           string
	host             string
	port             int
	controlPlanes    map[string]*ControlPlane
	meshes           map[string]*MeshBuilder
	policies         map[string]*PolicyBuilder
}

type ModifyMeshBuilder interface {
	OnlyBuild
	AddressableBuilder
	AddMesh(mesh *MeshBuilder) ModifyMeshBuilder
	RemoveMesh(name string) ModifyMeshBuilder
}

type ModifyPolicyBuilder interface {
	OnlyBuild
	AddressableBuilder
	AddPolicy(builder *PolicyBuilder) ModifyPolicyBuilder
	RemovePolicy(name string) ModifyPolicyBuilder
}

type AddressableBuilder interface {
	ResourceAddress(s string, resource string) string
}

type OnlyBuild interface {
	Build() string
}

func NewBuilder(provider ProviderType, scheme, host string, port int) *Builder {
	return &Builder{
		provider:      provider,
		scheme:        scheme,
		host:          host,
		port:          port,
		controlPlanes: make(map[string]*ControlPlane),
		meshes:        make(map[string]*MeshBuilder),
		policies:      make(map[string]*PolicyBuilder),
	}
}

func (b *Builder) AddControlPlane(cp *ControlPlane) *Builder {
	b.controlPlanes[cp.ResourceName] = cp
	return b
}

func (b *Builder) AddMesh(mesh *MeshBuilder) *Builder {
	b.meshes[mesh.ResourceName] = mesh
	return b
}

func (b *Builder) RemoveMesh(name string) *Builder {
	delete(b.meshes, name)
	return b
}

func (b *Builder) AddPolicy(p *PolicyBuilder) *Builder {
	b.policies[p.ResourceName] = p
	return b
}

func (b *Builder) RemovePolicy(name string) *Builder {
	delete(b.policies, name)
	return b
}

func (b *Builder) WithProviderProperty(providerProperty ProviderType) *Builder {
	b.providerProperty = providerProperty
	return b
}

func (b *Builder) Build() string {
	var sb strings.Builder

	sb.WriteString(b.renderTemplate("provider.tmpl", map[string]interface{}{
		"Provider":         b.provider,
		"ProviderProperty": b.providerProperty,
		"Scheme":           b.scheme,
		"Host":             b.host,
		"Port":             b.port,
	}))
	sb.WriteString("\n")

	for _, cp := range b.controlPlanes {
		sb.WriteString(cp.Render(b))
		sb.WriteString("\n")
	}

	for _, mesh := range b.meshes {
		sb.WriteString(mesh.Render(b))
		sb.WriteString("\n")
	}

	for _, policy := range b.policies {
		sb.WriteString(policy.Render(b))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (b *Builder) ResourceAddress(resourceType, resourceName string) string {
	return fmt.Sprintf("%s_%s.%s", b.provider, resourceType, resourceName)
}

func (b *Builder) renderTemplate(file string, data interface{}) string {
	tmplBytes, err := templatesFS.ReadFile("templates/" + file)
	if err != nil {
		panic(fmt.Errorf("failed to read template %s: %w", file, err))
	}

	tmpl, err := template.New(file).Parse(string(tmplBytes))
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %w", err))
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		panic(fmt.Errorf("failed to render template: %w", err))
	}
	return out.String()
}
