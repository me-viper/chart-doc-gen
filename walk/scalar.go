/*
Copyright The Kubepack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import "sigs.k8s.io/kustomize/kyaml/yaml"

// walkScalar returns the value of VisitScalar
func (l Walker) walkScalar() (*yaml.RNode, error) {
	return l.VisitScalar(l.Source, l.Schema)
}
