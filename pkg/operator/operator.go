/*
SPDX-FileCopyrightText: 2024SAP SE or an SAP affiliate company and redis-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package operator

import (
	"embed"
	"flag"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sap/component-operator-runtime/pkg/component"
	"github.com/sap/component-operator-runtime/pkg/manifests"
	"github.com/sap/component-operator-runtime/pkg/manifests/helm"
	"github.com/sap/component-operator-runtime/pkg/operator"

	operatorv1alpha1 "github.com/sap/redis-operator/api/v1alpha1"
	"github.com/sap/redis-operator/internal/transformer"
)

const Name = "redis-operator.cs.sap.com"

//go:embed all:data
var data embed.FS

type Options struct {
	Name       string
	FlagPrefix string
}

type Operator struct {
	options Options
}

var defaultOperator operator.Operator = New()

func GetName() string {
	return defaultOperator.GetName()
}

func InitScheme(scheme *runtime.Scheme) {
	defaultOperator.InitScheme(scheme)
}

func InitFlags(flagset *flag.FlagSet) {
	defaultOperator.InitFlags(flagset)
}

func ValidateFlags() error {
	return defaultOperator.ValidateFlags()
}

func GetUncacheableTypes() []client.Object {
	return defaultOperator.GetUncacheableTypes()
}

func Setup(mgr ctrl.Manager) error {
	return defaultOperator.Setup(mgr)
}

func New() *Operator {
	return NewWithOptions(Options{})
}

func NewWithOptions(options Options) *Operator {
	operator := &Operator{options: options}
	if operator.options.Name == "" {
		operator.options.Name = Name
	}
	return operator
}

func (o *Operator) GetName() string {
	return o.options.Name
}

func (o *Operator) InitScheme(scheme *runtime.Scheme) {
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
}

func (o *Operator) InitFlags(flagset *flag.FlagSet) {
}

func (o *Operator) ValidateFlags() error {
	return nil
}

func (o *Operator) GetUncacheableTypes() []client.Object {
	return []client.Object{&operatorv1alpha1.Redis{}}
}

func (o *Operator) Setup(mgr ctrl.Manager) error {
	parameterTransformer, err := manifests.NewTemplateParameterTransformer(data, "data/parameters.yaml")
	if err != nil {
		return errors.Wrap(err, "error initializing parameter transformer")
	}
	objectTransformer := transformer.NewObjectTransformer()
	resourceGenerator, err := helm.NewTransformableHelmGenerator(
		data,
		"data/charts/redis",
		mgr.GetClient(),
	)
	if err != nil {
		return errors.Wrap(err, "error initializing resource generator")
	}
	resourceGenerator.
		WithParameterTransformer(parameterTransformer).
		WithObjectTransformer(objectTransformer)

	// TODO: handle increases of persistence.size somehow (instead of making it immutable)
	// this would require to recreate the statefulset (since persistentVolumeClaimTemplate is immutable)
	// and to extend existing persistent volume claims (supposing that they are resizable)

	if err := component.NewReconciler[*operatorv1alpha1.Redis](
		o.options.Name,
		resourceGenerator,
		component.ReconcilerOptions{},
	).WithPostReconcileHook(
		reconcileBinding,
	).SetupWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create controller")
	}

	operatorv1alpha1.NewWebhook().SetupWithManager(mgr)

	return nil
}
