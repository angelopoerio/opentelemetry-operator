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

package collector

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	"github.com/open-telemetry/opentelemetry-operator/internal/manifests"
	"github.com/open-telemetry/opentelemetry-operator/pkg/featuregate"
)

// Build is currently unused, but will be implemented to solve
// https://github.com/open-telemetry/opentelemetry-operator/issues/1876
func Build(params manifests.Params) ([]client.Object, error) {
	var resourceManifests []client.Object
	var manifestFactories []manifests.K8sManifestFactory
	switch params.Instance.Spec.Mode {
	case v1alpha1.ModeDeployment:
		manifestFactories = append(manifestFactories, manifests.FactoryWithoutError(Deployment))
	case v1alpha1.ModeStatefulSet:
		manifestFactories = append(manifestFactories, manifests.FactoryWithoutError(StatefulSet))
	case v1alpha1.ModeDaemonSet:
		manifestFactories = append(manifestFactories, manifests.FactoryWithoutError(DaemonSet))
	case v1alpha1.ModeSidecar:
		params.Log.V(5).Info("not building sidecar...")
	}
	manifestFactories = append(manifestFactories, []manifests.K8sManifestFactory{
		manifests.FactoryWithoutError(ConfigMap),
		manifests.FactoryWithoutError(HorizontalPodAutoscaler),
		manifests.FactoryWithoutError(ServiceAccount),
		manifests.FactoryWithoutError(Service),
		manifests.FactoryWithoutError(HeadlessService),
		manifests.FactoryWithoutError(MonitoringService),
		manifests.FactoryWithoutError(Ingress),
	}...)
	if params.Instance.Spec.Observability.Metrics.EnableMetrics && featuregate.PrometheusOperatorIsAvailable.IsEnabled() {
		manifestFactories = append(manifestFactories, manifests.Factory(ServiceMonitor))
	}
	for _, factory := range manifestFactories {
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
	routes := Routes(params.Config, params.Log, params.Instance)
	// NOTE: we cannot just unpack the slice, the type checker doesn't coerce the type correctly.
	for _, route := range routes {
		resourceManifests = append(resourceManifests, route)
	}
	return resourceManifests, nil
}
