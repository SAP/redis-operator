/*
SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and redis-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"context"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/sap/admission-webhook-runtime/pkg/admission"
)

// +kubebuilder:object:generate=false
type Webhook struct {
}

var _ admission.ValidatingWebhook[*Redis] = &Webhook{}

func NewWebhook() *Webhook {
	return &Webhook{}
}

func (w *Webhook) ValidateCreate(ctx context.Context, redis *Redis) error {
	if err := validateTemplate(redis); err != nil {
		return err
	}
	return nil
}

func (w *Webhook) ValidateUpdate(ctx context.Context, oldRedis *Redis, redis *Redis) error {
	if (oldRedis.Spec.Sentinel != nil && oldRedis.Spec.Sentinel.Enabled) != (redis.Spec.Sentinel != nil && redis.Spec.Sentinel.Enabled) {
		return fmt.Errorf(".spec.sentinel.enabled is immutable")
	}
	if (oldRedis.Spec.Persistence != nil && oldRedis.Spec.Persistence.Enabled) != (redis.Spec.Persistence != nil && redis.Spec.Persistence.Enabled) {
		return fmt.Errorf(".spec.persistence.enabled is immutable")
	}
	if oldRedis.Spec.Persistence == nil && redis.Spec.Persistence != nil && redis.Spec.Persistence.Size != nil ||
		oldRedis.Spec.Persistence != nil && oldRedis.Spec.Persistence.Size != nil && redis.Spec.Persistence == nil ||
		oldRedis.Spec.Persistence != nil && redis.Spec.Persistence != nil && !quantityEqual(oldRedis.Spec.Persistence.Size, redis.Spec.Persistence.Size) {
		return fmt.Errorf(".spec.persistence.size is immutable")
	}
	if oldRedis.Spec.Persistence == nil && redis.Spec.Persistence != nil && redis.Spec.Persistence.StorageClass != "" ||
		oldRedis.Spec.Persistence != nil && oldRedis.Spec.Persistence.StorageClass != "" && redis.Spec.Persistence == nil ||
		oldRedis.Spec.Persistence != nil && redis.Spec.Persistence != nil && oldRedis.Spec.Persistence.StorageClass != redis.Spec.Persistence.StorageClass {
		return fmt.Errorf(".spec.persistence.storageClass is immutable")
	}
	if err := validateTemplate(redis); err != nil {
		return err
	}
	return nil
}

func (w *Webhook) ValidateDelete(ctx context.Context, redis *Redis) error {
	return nil
}

func (w *Webhook) SetupWithManager(mgr manager.Manager) {
	mgr.GetWebhookServer().Register(
		fmt.Sprintf("/admission/%s/redis/validate", GroupVersion),
		admission.NewValidatingWebhookHandler[*Redis](w, mgr.GetScheme(), mgr.GetLogger().WithName("webhook-runtime")),
	)
}

func validateTemplate(redis *Redis) error {
	if redis.Spec.Binding != nil && redis.Spec.Binding.Template != nil {
		t := template.New("binding.yaml").Option("missingkey=zero").Funcs(sprig.TxtFuncMap())
		if _, err := t.Parse(*redis.Spec.Binding.Template); err != nil {
			return errors.Wrapf(err, "specified .spec.binding.template is invalid")
		}
	}
	return nil
}

func quantityEqual(x *resource.Quantity, y *resource.Quantity) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil && y != nil || x != nil && y == nil {
		return false
	}
	return *x == *y
}
