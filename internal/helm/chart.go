package helm

import (
	"context"
	"os"

	yarerrors "github.com/yar-run/yar/internal/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

// LoadChart loads and validates a chart from an explicit path.
func (c *client) LoadChart(path string) (*chart.Chart, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, &yarerrors.NotFoundError{Resource: "chart", Name: path, Message: "path not found"}
		}
		return nil, chartError(nil, "failed to read chart", err)
	}
	chart, err := loader.Load(path)
	if err != nil {
		return nil, chartError(nil, "failed to load chart", err)
	}
	return chart, nil
}

// LoadChartFiles loads and validates a chart from in-memory files.
func (c *client) LoadChartFiles(files []ChartFile) (*chart.Chart, error) {
	chartFiles := make([]*loader.BufferedFile, len(files))
	for i, file := range files {
		chartFiles[i] = &loader.BufferedFile{Name: file.Name, Data: append([]byte(nil), file.Data...)}
	}
	chart, err := loader.LoadFiles(chartFiles)
	if err != nil {
		return nil, chartError(nil, "failed to load chart", err)
	}
	return chart, nil
}

// Render deterministically renders a chart without contacting a cluster.
func (c *client) Render(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts RenderOptions) (*RenderResult, error) {
	if err := c.ensureChart(chart); err != nil {
		return nil, err
	}
	if err := release.validate(); err != nil {
		return nil, err
	}

	valueCopy, err := cloneValues(values)
	if err != nil {
		return nil, err
	}
	if err := validateChartValues(chart, valueCopy); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	configuration := &action.Configuration{Log: func(string, ...interface{}) {}}
	install := action.NewInstall(configuration)
	install.ClientOnly = true
	install.DryRun = true
	install.HideSecret = true
	install.EnableDNS = false
	install.IncludeCRDs = opts.IncludeCRDs
	install.Namespace = release.Namespace
	install.ReleaseName = release.Name
	install.KubeVersion, err = renderKubeVersion(opts.KubeVersion)
	if err != nil {
		return nil, err
	}
	install.APIVersions = append([]string(nil), opts.APIVersions...)

	result, err := install.Run(chart, valueCopy)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		return nil, redactedChartError(chart, "failed to render chart")
	}
	return &RenderResult{Manifest: result.Manifest, Notes: result.Info.Notes}, nil
}

func renderKubeVersion(value string) (*chartutil.KubeVersion, error) {
	if value == "" {
		return nil, nil
	}
	version, err := chartutil.ParseKubeVersion(value)
	if err != nil {
		return nil, validationError("render.kubeVersion", value, "must be a valid Kubernetes version")
	}
	return version, nil
}

func redactedChartError(chart *chart.Chart, message string) *yarerrors.PackError {
	name := "chart"
	if chart != nil && chart.Metadata != nil && chart.Name() != "" {
		name = chart.Name()
	}
	return &yarerrors.PackError{Pack: name, Message: message}
}
