/*
Copyright 2023 SAP SE.

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

package transformer

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type objectTransformer struct{}

func NewObjectTransformer() *objectTransformer {
	return &objectTransformer{}
}

func (t *objectTransformer) TransformObjects(namespace string, name string, objects []client.Object) ([]client.Object, error) {
	for i := 0; i < len(objects); i++ {
		if statefulSet := asStatefulSet(objects[i]); statefulSet != nil {
			if len(statefulSet.Spec.Template.Spec.TopologySpreadConstraints) == 0 {
				statefulSet.Spec.Template.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{
					{
						MaxSkew:            1,
						TopologyKey:        "kubernetes.io/hostname",
						WhenUnsatisfiable:  corev1.ScheduleAnyway,
						NodeAffinityPolicy: &[]corev1.NodeInclusionPolicy{corev1.NodeInclusionPolicyHonor}[0],
						NodeTaintsPolicy:   &[]corev1.NodeInclusionPolicy{corev1.NodeInclusionPolicyHonor}[0],
					},
				}
			}
			for j := 0; j < len(statefulSet.Spec.Template.Spec.TopologySpreadConstraints); j++ {
				constraint := &statefulSet.Spec.Template.Spec.TopologySpreadConstraints[j]
				if constraint.LabelSelector == nil && len(constraint.MatchLabelKeys) == 0 {
					constraint.LabelSelector = statefulSet.Spec.Selector
					constraint.MatchLabelKeys = []string{"controller-revision-hash"}
				}
			}
			objects[i] = asUnstructurable(statefulSet)
		}
	}
	// TODO: set persistentVolumeClaimRetentionPolicy to Delete (available from 1.27; unless chart natively supports it)
	return objects, nil
}

func asStatefulSet(object client.Object) *appsv1.StatefulSet {
	if statefulSet, ok := object.(*appsv1.StatefulSet); ok {
		return statefulSet
	}
	if object, ok := object.(*unstructured.Unstructured); ok && (object.GetObjectKind().GroupVersionKind() == schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}) {
		statefulSet := &appsv1.StatefulSet{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, statefulSet); err != nil {
			panic(err)
		}
		return statefulSet
	}
	return nil
}

func asUnstructurable(object client.Object) *unstructured.Unstructured {
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		panic(err)
	}
	return &unstructured.Unstructured{Object: m}
}
