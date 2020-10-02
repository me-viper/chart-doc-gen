// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// walkMap returns the value of VisitMap
//
// - call VisitMap
// - set the return value on l.Dest
// - walk each source field
// - set each source field value on l.Dest
func (l Walker) walkMap() (*yaml.RNode, error) {
	// get the new map value
	dest, err := l.VisitMap(l.Source, l.Schema)
	if dest == nil || err != nil {
		return nil, err
	}
	l.Source = dest

	// recursively set the field values on the map
	for _, key := range l.fieldNames() {
		if l.VisitKeysAsScalars {
			// visit the map keys as if they were scalars,
			// this is necessary if doing things such as copying
			// comments
			// construct the sources from the keys
			if l.Source == nil {
				continue
			}
			field := l.Source.Field(key)
			if field == nil || yaml.IsEmpty(field.Key) {
				continue
			}

			// visit the sources as a scalar
			// keys don't have any schema --pass in nil
			_, err := l.Visitor.VisitScalar(field.Key, nil)
			if err != nil {
				return nil, err
			}

			ignore := CommentValue(field.Key.YNode().LineComment) == "+doc-gen:ignore" ||
				CommentValue(field.Key.YNode().HeadComment) == "+doc-gen:ignore"

			if ignore {
				continue
			}

			breakout := CommentValue(field.Key.YNode().LineComment) == "+doc-gen:break"
			if field.Value.YNode().Kind == yaml.ScalarNode ||
				yaml.IsEmpty(field.Value) ||
				breakout {
				_, err = l.VisitLeaf(field.Key, field.Value, strings.Join(append(l.Path, key), "."), nil)
				if err != nil {
					return nil, err
				}
			}

			if breakout {
				continue
			}
		}

		var s *openapi.ResourceSchema
		if l.Schema != nil {
			s = l.Schema.Field(key)
		}
		fv, commentSch := l.fieldValue(key)
		if commentSch != nil {
			s = commentSch
		}
		val, err := Walker{
			VisitKeysAsScalars:    l.VisitKeysAsScalars,
			InferAssociativeLists: l.InferAssociativeLists,
			Visitor:               l,
			Schema:                s,
			Source:                fv,
			Path:                  append(l.Path, key)}.Walk()
		if err != nil {
			return nil, err
		}

		// this handles empty and non-empty values
		_, err = dest.Pipe(yaml.FieldSetter{Name: key, Value: val})
		if err != nil {
			return nil, err
		}
	}

	return dest, nil
}

// valueIfPresent returns node.Value if node is non-nil, otherwise returns nil
func (l Walker) valueIfPresent(node *yaml.MapNode) (*yaml.RNode, *openapi.ResourceSchema) {
	if node == nil {
		return nil, nil
	}

	// parse the schema for the field if present
	var s *openapi.ResourceSchema
	fm := fieldmeta.FieldMeta{}
	var err error
	// check the value for a schema
	if err = fm.Read(node.Value); err == nil {
		s = &openapi.ResourceSchema{Schema: &fm.Schema}
		if fm.Schema.Ref.String() != "" {
			r, err := openapi.Resolve(&fm.Schema.Ref)
			if err == nil && r != nil {
				s.Schema = r
			}
		}
	}
	// check the key for a schema -- this will be used
	// when the value is a Sequence (comments are attached)
	// to the key
	if fm.IsEmpty() {
		if err = fm.Read(node.Key); err == nil {
			s = &openapi.ResourceSchema{Schema: &fm.Schema}
		}
		if fm.Schema.Ref.String() != "" {
			r, err := openapi.Resolve(&fm.Schema.Ref)
			if err == nil && r != nil {
				s.Schema = r
			}
		}
	}
	return node.Value, s
}

// fieldNames returns a sorted slice containing the names of all fields that appear in any of
// the sources
func (l Walker) fieldNames() []string {
	if l.Source == nil {
		return nil
	}
	// don't check error, we know this is a mapping node
	result, _ := l.Source.Fields()
	return result
}

// fieldValue returns a slice containing each source's value for fieldName
func (l Walker) fieldValue(fieldName string) (*yaml.RNode, *openapi.ResourceSchema) {
	if l.Source == nil {
		return nil, nil
	}

	var sch *openapi.ResourceSchema
	field := l.Source.Field(fieldName)
	f, s := l.valueIfPresent(field)
	if sch == nil && !s.IsEmpty() {
		sch = s
	}
	return f, sch
}
