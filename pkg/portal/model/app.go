package model

import (
	"path"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type App struct {
	ID      string
	Context *config.AppContext
}

func (a *App) LoadRawAppConfig() (*config.AppConfig, error) {
	files, err := configsource.AppConfig.ReadResource(a.Context.AppFs)
	if err != nil {
		return nil, err
	}

	var cfg *config.AppConfig
	if err := yaml.Unmarshal(files[0].Data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (a *App) LoadRawSecretConfig() (*config.SecretConfig, error) {
	files, err := configsource.SecretConfig.ReadResource(a.Context.AppFs)
	if err != nil {
		return nil, err
	}

	var cfg *config.SecretConfig
	if err := yaml.Unmarshal(files[0].Data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (a *App) ListResources() ([]string, error) {
	filePaths := make(map[string]struct{})
	var list func(fs resource.Fs, dir string) error
	list = func(fs resource.Fs, dir string) error {
		f, err := fs.Open(dir)
		if err != nil {
			return err
		}
		defer f.Close()

		files, err := f.Readdirnames(0)
		if err != nil {
			return err
		}
		for _, f := range files {
			p := path.Join(dir, f)
			f, err := fs.Stat(p)
			if err != nil {
				return err
			}

			if f.IsDir() {
				if err := list(fs, p); err != nil {
					return err
				}
				continue
			}
			filePaths[strings.TrimPrefix(p, "/")] = struct{}{}
		}
		return nil
	}

	for _, fs := range a.Context.Resources.Fs {
		if err := list(fs, "/"); err != nil {
			return nil, err
		}
	}

	// Filter out non-resource file paths in case of local FS
	for p := range filePaths {
		found := false
		for _, desc := range a.Context.Resources.Registry.Descriptors {
			if !desc.MatchResource(p) {
				continue
			}
			found = true
			break
		}

		if !found {
			delete(filePaths, p)
		}
	}

	var paths []string
	for p := range filePaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	return paths, nil
}

func (a *App) LoadResources(paths ...string) ([]*AppResource, error) {
	type matchedResource struct {
		Path       string
		Descriptor resource.Descriptor
	}

	// Match input path with corresponding resource descriptors
	var matches []matchedResource
	for _, p := range paths {
		found := false
		for _, desc := range a.Context.Resources.Registry.Descriptors {
			if !desc.MatchResource(p) {
				continue
			}
			matches = append(matches, matchedResource{
				Path:       p,
				Descriptor: desc,
			})
			found = true
			break
		}
		if !found {
			return nil, apierrors.NewInvalid("unknown resource: " + p)
		}
	}

	// Load resource file layers for each match
	var resources []*AppResource
	for _, match := range matches {
		var files []AppResourceFile
		for _, fs := range a.Context.Resources.Fs {
			layers, err := match.Descriptor.ReadResource(fs)
			if err != nil {
				return nil, err
			}

			for _, l := range layers {
				if l.Path != match.Path {
					continue
				}
				files = append(files, AppResourceFile{Fs: fs, Data: l.Data})
			}
		}

		resources = append(resources, &AppResource{
			Context:    a.Context,
			Descriptor: match.Descriptor,
			Path:       match.Path,
			Files:      files,
		})
	}

	return resources, nil
}

type AppResourceFile struct {
	Fs   resource.Fs
	Data []byte
}

type AppResource struct {
	Context    *config.AppContext
	Descriptor resource.Descriptor
	Path       string
	Files      []AppResourceFile
}

type AppConfigFile struct {
	Path    string
	Content string
}
