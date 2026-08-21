package docker

import (
	"context"
	stderrors "errors"
	"fmt"
	"path/filepath"

	composecli "github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	yarerrors "github.com/yar-run/yar/internal/errors"
)

// LoadCompose loads and validates explicitly supplied Compose files.
// It never reads ambient process environment variables or .env files.
func LoadCompose(ctx context.Context, paths ...string) (*composetypes.Project, error) {
	if len(paths) == 0 {
		return nil, composeConfigError("<compose>", fmt.Errorf("at least one Compose file is required"))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	absPaths := make([]string, len(paths))
	for i, path := range paths {
		if path == "" {
			return nil, composeConfigError("<compose>", fmt.Errorf("Compose file path cannot be empty"))
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return nil, composeConfigError(path, err)
		}
		absPaths[i] = absolutePath
	}

	options, err := composecli.NewProjectOptions(
		absPaths,
		composecli.WithWorkingDirectory(filepath.Dir(absPaths[0])),
		composecli.WithInterpolation(false),
		composecli.WithoutEnvironmentResolution,
		composecli.WithoutLabelsResolution,
		composecli.WithLoadOptions(func(options *loader.Options) {
			// Includes may load their own .env files, which violates Yar's parser boundary.
			options.SkipInclude = true
		}),
	)
	if err != nil {
		return nil, composeLoadError(absPaths[0], err)
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		return nil, composeLoadError(absPaths[0], err)
	}
	return project, nil
}

func composeLoadError(path string, err error) error {
	if stderrors.Is(err, context.Canceled) || stderrors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return composeConfigError(path, err)
}

func composeConfigError(path string, err error) *yarerrors.ConfigError {
	return &yarerrors.ConfigError{
		Path:    path,
		Message: "failed to load Compose file",
		Err:     err,
	}
}
