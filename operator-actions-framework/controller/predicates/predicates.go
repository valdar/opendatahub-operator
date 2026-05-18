package predicates

import (
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/opendatahub-io/operator-actions-framework/controller/predicates/generation"
	"github.com/opendatahub-io/operator-actions-framework/controller/predicates/resources"
)

var (
	DefaultPredicate = predicate.Or(
		generation.New(),
		predicate.LabelChangedPredicate{},
		predicate.AnnotationChangedPredicate{},
	)

	DefaultDeploymentPredicate = predicate.Or(
		resources.NewDeploymentPredicate(),
		predicate.LabelChangedPredicate{},
		predicate.AnnotationChangedPredicate{},
	)
)
