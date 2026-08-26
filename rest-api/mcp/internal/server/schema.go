// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/jsonschema-go/jsonschema"
)

// NicoOpenApiHandler builds the MCP tool input schema for a single OpenAPI
// operation, combining path/query parameters, a JSON request body when one is
// defined, and the common per-call config fields.
type NicoOpenApiHandler struct {
	schema jsonschema.Schema
	defs   map[string]*jsonschema.Schema
}

type paramKey struct {
	in   string
	name string
}

// mergeParameters combines path-item and operation parameters, with
// operation-level definitions overriding path-item-level ones that share the
// same {in,name} tuple per OpenAPI override semantics.
func mergeParameters(item *openapi3.PathItem, op *openapi3.Operation) []*openapi3.Parameter {
	merged := map[paramKey]*openapi3.Parameter{}
	add := func(refs openapi3.Parameters) {
		for _, ref := range refs {
			if ref == nil || ref.Value == nil {
				continue
			}
			p := ref.Value
			merged[paramKey{in: p.In, name: p.Name}] = p
		}
	}
	add(item.Parameters)
	add(op.Parameters)

	out := make([]*openapi3.Parameter, 0, len(merged))
	for _, p := range merged {
		out = append(out, p)
	}
	return out
}

// buildInput populates the handler's schema from the operation's path and
// query parameters merged with the four common config fields (org,
// base_url, api_name, token) and returns it. Path parameters are required;
// OpenAPI-required query parameters are required; the config fields are
// always optional.
func (h *NicoOpenApiHandler) buildInput(item *openapi3.PathItem, op *openapi3.Operation) *jsonschema.Schema {
	props := map[string]*jsonschema.Schema{}
	requiredSet := map[string]struct{}{}
	h.defs = map[string]*jsonschema.Schema{}

	for _, p := range mergeParameters(item, op) {
		if p.Name == "org" {
			// Resolved from per-call args or server startup defaults.
			// The OpenAPI {org} segment is filled in by appcli.Client.Do.
			continue
		}
		if p.In != "path" && p.In != "query" {
			continue
		}
		props[p.Name] = h.fromParam(p)
		if p.In == "path" || p.Required {
			requiredSet[p.Name] = struct{}{}
		}
	}
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		mediaType := op.RequestBody.Value.GetMediaType("application/json")
		if mediaType != nil && mediaType.Schema != nil {
			props["body"] = h.fromSchemaRef(mediaType.Schema)
			if op.RequestBody.Value.Required {
				requiredSet["body"] = struct{}{}
			}
		}
	}

	for _, c := range commonConfigDescriptions {
		if _, exists := props[c.Name]; exists {
			continue
		}
		props[c.Name] = &jsonschema.Schema{
			Type:        "string",
			Description: c.Desc,
		}
	}

	h.schema = jsonschema.Schema{
		Type:                 "object",
		Properties:           props,
		Required:             slices.Sorted(maps.Keys(requiredSet)),
		AdditionalProperties: falseJSONSchema(),
		Defs:                 h.defs,
	}
	return &h.schema
}

func falseJSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

// fromParam converts a single OpenAPI parameter to a JSON schema fragment.
// Types are normalised to integer/boolean/number/string; everything else
// falls back to string. Scalar validation hints such as format, min/max,
// length bounds, defaults, and enums are preserved where present so MCP
// clients get the same guardrails as the generated CLI flags.
func (*NicoOpenApiHandler) fromParam(p *openapi3.Parameter) *jsonschema.Schema {
	s := &jsonschema.Schema{Description: p.Description}
	if p.Schema == nil || p.Schema.Value == nil {
		s.Type = "string"
		return s
	}
	sch := p.Schema.Value
	switch {
	case sch.Type.Is("integer"):
		s.Type = "integer"
	case sch.Type.Is("boolean"):
		s.Type = "boolean"
	case sch.Type.Is("number"):
		s.Type = "number"
	default:
		s.Type = "string"
	}
	if len(sch.Enum) > 0 {
		s.Enum = slices.Clone(sch.Enum)
	}
	s.Format = sch.Format
	if sch.MinLength > 0 {
		v := int(sch.MinLength)
		s.MinLength = &v
	}
	if sch.MaxLength != nil {
		v := int(*sch.MaxLength)
		s.MaxLength = &v
	}
	if sch.Min != nil {
		v := *sch.Min
		s.Minimum = &v
	}
	if sch.Max != nil {
		v := *sch.Max
		s.Maximum = &v
	}
	if sch.Default != nil {
		if b, err := json.Marshal(sch.Default); err == nil {
			s.Default = b
		}
	}
	return s
}

func (h *NicoOpenApiHandler) fromSchemaRef(ref *openapi3.SchemaRef) *jsonschema.Schema {
	if ref == nil || ref.Value == nil {
		return &jsonschema.Schema{}
	}
	const componentPrefix = "#/components/schemas/"
	if name, ok := strings.CutPrefix(ref.Ref, componentPrefix); ok {
		if _, exists := h.defs[name]; !exists {
			h.defs[name] = &jsonschema.Schema{}
			*h.defs[name] = *h.fromSchema(ref.Value)
		}
		return &jsonschema.Schema{Ref: "#/$defs/" + name}
	}
	return h.fromSchema(ref.Value)
}

func (h *NicoOpenApiHandler) fromSchema(source *openapi3.Schema) *jsonschema.Schema {
	target := &jsonschema.Schema{
		Title:       source.Title,
		Description: source.Description,
		Format:      source.Format,
		Deprecated:  source.Deprecated,
		ReadOnly:    source.ReadOnly,
		WriteOnly:   source.WriteOnly,
		Enum:        slices.Clone(source.Enum),
		Pattern:     source.Pattern,
		UniqueItems: source.UniqueItems,
		Minimum:     source.Min,
		Maximum:     source.Max,
		MultipleOf:  source.MultipleOf,
		Required:    slices.Clone(source.Required),
	}

	types := slices.Clone(source.Type.Slice())
	if source.Nullable && !slices.Contains(types, "null") {
		types = append(types, "null")
	}
	if len(types) == 1 {
		target.Type = types[0]
	} else if len(types) > 1 {
		target.Types = types
	}

	if source.Default != nil {
		if encoded, err := json.Marshal(source.Default); err == nil {
			target.Default = encoded
		}
	}
	if source.MinLength > 0 {
		value := int(source.MinLength)
		target.MinLength = &value
	}
	if source.MaxLength != nil {
		value := int(*source.MaxLength)
		target.MaxLength = &value
	}
	if source.MinItems > 0 {
		value := int(source.MinItems)
		target.MinItems = &value
	}
	if source.MaxItems != nil {
		value := int(*source.MaxItems)
		target.MaxItems = &value
	}
	if source.MinProps > 0 {
		value := int(source.MinProps)
		target.MinProperties = &value
	}
	if source.MaxProps != nil {
		value := int(*source.MaxProps)
		target.MaxProperties = &value
	}
	if source.ExclusiveMin.Value != nil {
		target.ExclusiveMinimum = source.ExclusiveMin.Value
	} else if source.ExclusiveMin.IsTrue() {
		target.ExclusiveMinimum = source.Min
		target.Minimum = nil
	}
	if source.ExclusiveMax.Value != nil {
		target.ExclusiveMaximum = source.ExclusiveMax.Value
	} else if source.ExclusiveMax.IsTrue() {
		target.ExclusiveMaximum = source.Max
		target.Maximum = nil
	}
	if source.Items != nil {
		target.Items = h.fromSchemaRef(source.Items)
	}
	if len(source.Properties) > 0 {
		target.Properties = make(map[string]*jsonschema.Schema, len(source.Properties))
		for name, property := range source.Properties {
			target.Properties[name] = h.fromSchemaRef(property)
		}
	}
	if source.AdditionalProperties.Has != nil {
		if *source.AdditionalProperties.Has {
			target.AdditionalProperties = &jsonschema.Schema{}
		} else {
			target.AdditionalProperties = falseJSONSchema()
		}
	} else if source.AdditionalProperties.Schema != nil {
		target.AdditionalProperties = h.fromSchemaRef(source.AdditionalProperties.Schema)
	}
	target.AllOf = h.fromSchemaRefs(source.AllOf)
	target.AnyOf = h.fromSchemaRefs(source.AnyOf)
	target.OneOf = h.fromSchemaRefs(source.OneOf)
	if source.Not != nil {
		target.Not = h.fromSchemaRef(source.Not)
	}
	return target
}

func (h *NicoOpenApiHandler) fromSchemaRefs(refs openapi3.SchemaRefs) []*jsonschema.Schema {
	if len(refs) == 0 {
		return nil
	}
	result := make([]*jsonschema.Schema, 0, len(refs))
	for _, ref := range refs {
		result = append(result, h.fromSchemaRef(ref))
	}
	return result
}
