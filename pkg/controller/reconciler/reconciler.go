package reconciler

import (
	fwapi "github.com/opendatahub-io/operator-actions-framework/api"
	fwreconciler "github.com/opendatahub-io/operator-actions-framework/controller/reconciler"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
)

type Reconciler = fwreconciler.Reconciler

type ReconcilerOpt = fwreconciler.ReconcilerOpt

var (
	WithConditionsManagerFactory  = fwreconciler.WithConditionsManagerFactory
	WithRelease                   = fwreconciler.WithRelease
	WithFinalizerName             = fwreconciler.WithFinalizerName
	WithProvisioningConditionType = fwreconciler.WithProvisioningConditionType
	WithPhaseNames                = fwreconciler.WithPhaseNames
	WithDynamicOwnership          = fwreconciler.WithDynamicOwnership
)

// NewReconciler creates a new reconciler with ODH defaults
// (Release from cluster.GetRelease()).
func NewReconciler[T common.PlatformObject](mgr manager.Manager, name string, object T, opts ...ReconcilerOpt) (*Reconciler, error) {
	rel := cluster.GetRelease()
	defaults := []ReconcilerOpt{
		fwreconciler.WithRelease(fwapi.Release{Name: rel.Name, Version: rel.Version.Version}),
	}
	return fwreconciler.NewReconciler(mgr, name, object, append(defaults, opts...)...)
}
