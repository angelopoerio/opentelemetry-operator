// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package targetallocator

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/open-telemetry/opentelemetry-operator/internal/manifests"
)

// Build is currently unused, but will be implemented to solve
// https://github.com/open-telemetry/opentelemetry-operator/issues/1876
func Build(params manifests.Params) ([]client.Object, error) {
	var resourceManifests []client.Object
	if !params.Instance.Spec.TargetAllocator.Enabled {
		return resourceManifests, nil
	}
	resourceFactories := []manifests.K8sManifestFactory{
		manifests.Factory(ConfigMap),
		manifests.FactoryWithoutError(Deployment),
		manifests.FactoryWithoutError(ServiceAccount),
		manifests.FactoryWithoutError(Service),
	}
	for _, factory := range resourceFactories {
		res, err := factory(params.Config, params.Log, params.Instance)
		if err != nil {
			return nil, err
		} else if res != nil {
			// because of pointer semantics, res is still nil-able here as this is an interface pointer
			// read here for details:
			// https://github.com/open-telemetry/opentelemetry-operator/pull/1965#discussion_r1281705719
			resourceManifests = append(resourceManifests, res)
		}
	}
	return resourceManifests, nil
}
