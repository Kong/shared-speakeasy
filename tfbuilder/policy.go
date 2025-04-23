package tfbuilder

import (
	"bytes"
	"fmt"
	"text/template"
)

const AllowAllTrafficPermissionSpec = `
  spec = {
    from = [
      {
        target_ref = {
          kind = "Mesh"
        }
        default = {
          action = "Allow"
        }
      }
    ]
  }`

type PolicyBuilder struct {
	ResourceType string // e.g., "mesh_traffic_permission"
	ResourceName string // e.g., "allow_all"
	MeshRef      string // e.g., kong-mesh_mesh.default.name
	DependsOn    []string
	Labels       map[string]string
	PolicyName   string
	Type         string
	SpecHCL      string // literal HCL block
	CPID         string // Optional
}

func NewPolicyBuilder(resourceType, resourceName, policyName, policyType string) *PolicyBuilder {
	return &PolicyBuilder{
		ResourceType: resourceType,
		ResourceName: resourceName,
		PolicyName:   policyName,
		Type:         policyType,
		Labels:       make(map[string]string),
	}
}

func (p *PolicyBuilder) WithMeshRef(ref string) *PolicyBuilder {
	p.MeshRef = ref
	return p
}

func (p *PolicyBuilder) WithLabels(labels map[string]string) *PolicyBuilder {
	p.Labels = labels
	return p
}

func (p *PolicyBuilder) WithDependsOn(deps ...string) *PolicyBuilder {
	p.DependsOn = append(p.DependsOn, deps...)
	return p
}

func (p *PolicyBuilder) WithSpecHCL(hcl string) *PolicyBuilder {
	p.SpecHCL = hcl
	return p
}

func (p *PolicyBuilder) WithCPID(cpID string) *PolicyBuilder {
	p.CPID = cpID
	return p
}

func (p *PolicyBuilder) Render(provider *Builder) string {
	tmplBytes, err := templatesFS.ReadFile("templates/policy.tmpl")
	if err != nil {
		panic(fmt.Errorf("failed to read policy template: %w", err))
	}

	tmpl, err := template.New("policy").Parse(string(tmplBytes))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]interface{}{
		"Provider":         provider.provider,
		"ProviderProperty": provider.providerProperty,
		"ResourceType":     p.ResourceType,
		"ResourceName":     p.ResourceName,
		"MeshRef":          p.MeshRef,
		"DependsOn":        p.DependsOn,
		"Labels":           p.Labels,
		"Name":             p.PolicyName,
		"Type":             p.Type,
		"Spec":             p.SpecHCL,
		"CPID":             p.CPID,
	}); err != nil {
		panic(err)
	}

	return buf.String()
}
