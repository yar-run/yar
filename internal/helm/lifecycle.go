package helm

import (
	"context"
	stderrors "errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// Install creates a Yar-owned release.
func (c *client) Install(ctx context.Context, chart *chart.Chart, releaseID Release, values map[string]any, opts InstallOptions) (*release.Release, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureChart(chart); err != nil {
		return nil, err
	}
	if err := c.validateLifecycleRelease(releaseID); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if existing, err := c.config.Releases.Last(releaseID.Name); err == nil && existing != nil {
		return nil, lifecycleError("helm.install", releaseID, stderrors.New("release already exists"))
	} else if !stderrors.Is(err, driver.ErrReleaseNotFound) {
		return nil, lifecycleError("helm.install", releaseID, err)
	}

	install := action.NewInstall(c.config)
	install.ReleaseName = releaseID.Name
	install.Namespace = releaseID.Namespace
	install.Labels = releaseLabels(releaseID)
	install.DisableHooks = true
	install.TakeOwnership = false
	install.DryRun = opts.DryRun
	install.Timeout = c.timeout

	valueCopy, err := cloneValues(values)
	if err != nil {
		return nil, err
	}
	if err := validateChartValues(chart, valueCopy); err != nil {
		return nil, err
	}
	result, err := install.Run(chart, valueCopy)
	if err != nil {
		return nil, lifecycleError("helm.install", releaseID, err)
	}
	return result, nil
}

// Upgrade updates an exact Yar-owned release.
func (c *client) Upgrade(ctx context.Context, chart *chart.Chart, releaseID Release, values map[string]any, opts UpgradeOptions) (*release.Release, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureChart(chart); err != nil {
		return nil, err
	}
	if err := c.validateLifecycleRelease(releaseID); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := c.release(releaseID); err != nil {
		if stderrors.Is(err, errReleaseNotOwned) {
			return nil, ownershipError("helm.upgrade", releaseID)
		}
		return nil, lifecycleError("helm.upgrade", releaseID, err)
	}

	upgrade := action.NewUpgrade(c.config)
	upgrade.Namespace = releaseID.Namespace
	upgrade.Labels = releaseLabels(releaseID)
	upgrade.DisableHooks = true
	upgrade.TakeOwnership = false
	upgrade.DryRun = opts.DryRun
	upgrade.Timeout = c.timeout

	valueCopy, err := cloneValues(values)
	if err != nil {
		return nil, err
	}
	if err := validateChartValues(chart, valueCopy); err != nil {
		return nil, err
	}
	result, err := upgrade.Run(releaseID.Name, chart, valueCopy)
	if err != nil {
		return nil, lifecycleError("helm.upgrade", releaseID, err)
	}
	return result, nil
}

// Uninstall removes an exact Yar-owned release.
func (c *client) Uninstall(ctx context.Context, release Release, opts UninstallOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validateLifecycleRelease(release); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := c.release(release); err != nil {
		if stderrors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}
		if stderrors.Is(err, errReleaseNotOwned) {
			return ownershipError("helm.uninstall", release)
		}
		return lifecycleError("helm.uninstall", release, err)
	}

	uninstall := action.NewUninstall(c.config)
	uninstall.DisableHooks = true
	uninstall.KeepHistory = false
	uninstall.IgnoreNotFound = true
	uninstall.DryRun = opts.DryRun
	uninstall.Timeout = c.timeout

	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := uninstall.Run(release.Name)
	if err != nil {
		return lifecycleError("helm.uninstall", release, err)
	}
	return nil
}
