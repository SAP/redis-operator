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

package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kyaml "sigs.k8s.io/yaml"

	operatorv1alpha1 "github.com/sap/redis-operator/api/v1alpha1"
)

func reconcileBinding(ctx context.Context, client client.Client, redis *operatorv1alpha1.Redis) error {
	params := make(map[string]any)

	if redis.Spec.Sentinel != nil && redis.Spec.Sentinel.Enabled {
		params["sentinelEnabled"] = true
		params["host"] = fmt.Sprintf("redis-%s.%s.svc.cluster.local", redis.Name, redis.Namespace)
		params["port"] = 6379
		params["sentinelHost"] = fmt.Sprintf("redis-%s.%s.svc.cluster.local", redis.Name, redis.Namespace)
		params["sentinelPort"] = 26379
	} else {
		params["masterHost"] = fmt.Sprintf("redis-%s-master.%s.svc.cluster.local", redis.Name, redis.Namespace)
		params["masterPort"] = 6379
		params["replicaHost"] = fmt.Sprintf("redis-%s-replicas.%s.svc.cluster.local", redis.Name, redis.Namespace)
		params["replicaPort"] = 6379
	}

	authSecret := &corev1.Secret{}
	authSecretName := fmt.Sprintf("redis-%s", redis.Name)
	if err := client.Get(ctx, types.NamespacedName{Namespace: redis.Namespace, Name: authSecretName}, authSecret); err != nil {
		return err
	}
	params["password"] = string(authSecret.Data["redis-password"])

	if redis.Spec.TLS != nil && redis.Spec.TLS.Enabled {
		params["tlsEnabled"] = true
		tlsSecret := &corev1.Secret{}
		tlsSecretName := ""
		if redis.Spec.TLS.CertManager == nil {
			tlsSecretName = fmt.Sprintf("redis-%s-crt", redis.Name)
		} else {
			tlsSecretName = fmt.Sprintf("redis-%s-tls", redis.Name)
		}
		if err := client.Get(ctx, types.NamespacedName{Namespace: redis.Namespace, Name: tlsSecretName}, tlsSecret); err != nil {
			return err
		}
		params["caData"] = string(tlsSecret.Data["ca.crt"])
	}

	var buf bytes.Buffer
	t := template.New("binding.yaml").Option("missingkey=zero").Funcs(sprig.TxtFuncMap())
	if redis.Spec.Binding != nil && redis.Spec.Binding.Template != nil {
		if _, err := t.Parse(*redis.Spec.Binding.Template); err != nil {
			return err
		}
	} else {
		if _, err := t.ParseFS(data, "data/binding.yaml"); err != nil {
			return err
		}
	}
	if err := t.Execute(&buf, params); err != nil {
		return err
	}

	var bindingData map[string]any
	if err := kyaml.Unmarshal(buf.Bytes(), &bindingData); err != nil {
		return err
	}

	bindingSecret := &corev1.Secret{}
	bindingSecretName := fmt.Sprintf("redis-%s-binding", redis.Name)
	if err := client.Get(ctx, types.NamespacedName{Namespace: redis.Namespace, Name: bindingSecretName}, bindingSecret); err != nil {
		return err
	}
	bindingSecret.Data = make(map[string][]byte)
	for key, value := range bindingData {
		if stringValue, ok := value.(string); ok {
			bindingSecret.Data[key] = []byte(stringValue)
		} else {
			rawValue, err := json.Marshal(value)
			if err != nil {
				return err
			}
			bindingSecret.Data[key] = rawValue
		}
	}
	// TODO: avoid this update call if not necessary (e.g. by checking if data have changed)
	if err := client.Update(ctx, bindingSecret); err != nil {
		return err
	}

	return nil
}
