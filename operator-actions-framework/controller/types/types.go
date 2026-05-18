package types

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/go-logr/logr"
	helm "github.com/k8s-manifest-kit/renderer-helm/pkg"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/operator-actions-framework/api"
	"github.com/opendatahub-io/operator-actions-framework/controller/conditions"
	"github.com/opendatahub-io/operator-actions-framework/resources"
)

type Controller interface {
	Owns(gvk schema.GroupVersionKind) bool
	AddDynamicOwnedType(gvk schema.GroupVersionKind)
	GetClient() client.Client
	GetDiscoveryClient() discovery.DiscoveryInterface
	GetDynamicClient() dynamic.Interface
	IsDynamicOwnershipEnabled() bool
	IsExcludedFromDynamicOwnership(gvk schema.GroupVersionKind) bool
}

type ResourceObject interface {
	client.Object
	api.WithStatus
}

type WithLogger interface {
	GetLogger() logr.Logger
}

type ManifestInfo struct {
	Path       string
	ContextDir string
	SourcePath string
}

func (mi ManifestInfo) String() string {
	result := mi.Path

	if mi.ContextDir != "" {
		result = path.Join(result, mi.ContextDir)
	}

	if mi.SourcePath != "" {
		result = path.Join(result, mi.SourcePath)
	}

	return result
}

type TemplateInfo struct {
	FS   fs.FS
	Path string

	Labels      map[string]string
	Annotations map[string]string
}

type HookFn func(ctx context.Context, rr *ReconciliationRequest) error

type HelmChartInfo struct {
	helm.Source

	PreApply  []HookFn
	PostApply []HookFn
}

type ReconciliationRequest struct {
	Client            client.Client
	Controller        Controller
	Conditions        *conditions.Manager
	Instance          api.PlatformObject
	Release           api.Release
	ManifestsBasePath string
	Manifests         []ManifestInfo

	Templates  []TemplateInfo
	HelmCharts []HelmChartInfo
	Resources  []unstructured.Unstructured

	Generated bool
}

func (rr *ReconciliationRequest) AddResources(values ...client.Object) error {
	for i := range values {
		if values[i] == nil {
			continue
		}

		err := resources.EnsureGroupVersionKind(rr.Client.Scheme(), values[i])
		if err != nil {
			return fmt.Errorf("cannot normalize object: %w", err)
		}

		u, err := resources.ToUnstructured(values[i])
		if err != nil {
			return fmt.Errorf("cannot convert object to Unstructured: %w", err)
		}

		rr.Resources = append(rr.Resources, *u)
	}

	return nil
}

func (rr *ReconciliationRequest) ForEachResource(fn func(*unstructured.Unstructured) (bool, error)) error {
	for i := range rr.Resources {
		stop, err := fn(&rr.Resources[i])
		if err != nil {
			return fmt.Errorf("cannot process resource %s: %w", rr.Resources[i].GroupVersionKind(), err)
		}
		if stop {
			break
		}
	}

	return nil
}

func (rr *ReconciliationRequest) RemoveResources(predicate func(*unstructured.Unstructured) bool) error {
	writeIndex := 0
	for readIndex := range rr.Resources {
		if !predicate(&rr.Resources[readIndex]) {
			if writeIndex != readIndex {
				rr.Resources[writeIndex] = rr.Resources[readIndex]
			}
			writeIndex++
		}
	}

	for i := writeIndex; i < len(rr.Resources); i++ {
		rr.Resources[i] = unstructured.Unstructured{}
	}

	rr.Resources = rr.Resources[:writeIndex]
	return nil
}

func Hash(rr *ReconciliationRequest) ([]byte, error) {
	hash := sha256.New()

	instanceGeneration := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(instanceGeneration, rr.Instance.GetGeneration())

	if _, err := hash.Write([]byte(rr.Instance.GetUID())); err != nil {
		return nil, fmt.Errorf("failed to hash instance: %w", err)
	}
	if _, err := hash.Write(instanceGeneration); err != nil {
		return nil, fmt.Errorf("failed to hash instance generation: %w", err)
	}
	if _, err := hash.Write([]byte(rr.Release.Name)); err != nil {
		return nil, fmt.Errorf("failed to hash release: %w", err)
	}
	if _, err := hash.Write([]byte(rr.Release.Version.String())); err != nil {
		return nil, fmt.Errorf("failed to hash release: %w", err)
	}

	for i := range rr.Manifests {
		if _, err := hash.Write([]byte(rr.Manifests[i].String())); err != nil {
			return nil, fmt.Errorf("failed to hash manifest: %w", err)
		}
	}
	for i := range rr.Templates {
		if _, err := hash.Write([]byte(rr.Templates[i].Path)); err != nil {
			return nil, fmt.Errorf("failed to hash template: %w", err)
		}
	}
	for i := range rr.HelmCharts {
		if _, err := hash.Write([]byte(rr.HelmCharts[i].Chart)); err != nil {
			return nil, fmt.Errorf("failed to hash helm chart: %w", err)
		}
		if _, err := hash.Write([]byte(rr.HelmCharts[i].ReleaseName)); err != nil {
			return nil, fmt.Errorf("failed to hash helm chart release name: %w", err)
		}
		if rr.HelmCharts[i].Values != nil {
			values, err := rr.HelmCharts[i].Values(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("failed to get helm chart values: %w", err)
			}
			b, err := json.Marshal(values)
			if err != nil {
				return nil, fmt.Errorf("failed to hash helm chart values: %w", err)
			}
			if _, err := hash.Write(b); err != nil {
				return nil, fmt.Errorf("failed to hash helm chart values: %w", err)
			}
		}
	}

	return hash.Sum(nil), nil
}

func HashStr(rr *ReconciliationRequest) (string, error) {
	h, err := Hash(rr)
	if err != nil {
		return "", err
	}

	return resources.EncodeToString(h), nil
}
