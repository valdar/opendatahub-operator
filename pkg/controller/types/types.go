package types

import (
	fwtypes "github.com/opendatahub-io/operator-actions-framework/controller/types"
)

type Controller = fwtypes.Controller

type ResourceObject = fwtypes.ResourceObject

type WithLogger = fwtypes.WithLogger

type ManifestInfo = fwtypes.ManifestInfo

type TemplateInfo = fwtypes.TemplateInfo

type HookFn = fwtypes.HookFn

type HelmChartInfo = fwtypes.HelmChartInfo

type ReconciliationRequest = fwtypes.ReconciliationRequest

var (
	Hash    = fwtypes.Hash
	HashStr = fwtypes.HashStr
)
