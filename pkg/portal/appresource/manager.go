package appresource

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/util/checksum"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

//go:generate mockgen -source=manager.go -destination=manager_mock_test.go -package appresource_test

const ConfigFileMaxSize = 100 * 1024

type DenoClient interface {
	Check(ctx context.Context, snippet string) error
}

type TutorialService interface {
	OnUpdateResource(ctx context.Context, appID string, resourcesInAllFss []resource.ResourceFile, resourceInTargetFs *resource.ResourceFile, data []byte) (err error)
}

type DomainService interface {
	ListDomains(appID string) ([]*apimodel.Domain, error)
}

type Manager struct {
	Context            context.Context
	AppResourceManager *resource.Manager
	AppFS              resource.Fs
	AppFeatureConfig   *config.FeatureConfig
	AppHostSuffixes    *config.AppHostSuffixes
	DomainService      DomainService
	Tutorials          TutorialService
	DenoClient         DenoClient
	Clock              clock.Clock
}

func (m *Manager) List() ([]string, error) {
	r := m.AppResourceManager

	// Find the union all known paths in all FSs.
	filePaths := make(map[string]struct{})
	for _, fs := range r.Fs {
		locations, err := resource.EnumerateAllLocations(fs)
		if err != nil {
			return nil, err
		}
		for _, location := range locations {
			filePaths[location.Path] = struct{}{}
		}
	}

	// Omit paths that are not resources.
	for p := range filePaths {
		found := false
		for _, desc := range r.Registry.Descriptors {
			if _, ok := desc.MatchResource(p); !ok {
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

func (m *Manager) AssociateDescriptor(paths ...string) ([]DescriptedPath, error) {
	r := m.AppResourceManager

	var matches []DescriptedPath
	for _, p := range paths {
		found := false
		for _, desc := range r.Registry.Descriptors {
			if _, ok := desc.MatchResource(p); !ok {
				continue
			}
			matches = append(matches, DescriptedPath{
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
	return matches, nil
}

func (m *Manager) ReadAppFile(desc resource.Descriptor, view resource.AppFileView) (interface{}, error) {
	return m.AppResourceManager.Read(desc, view)
}

func (m *Manager) ApplyUpdates(appID string, updates []Update) ([]*resource.ResourceFile, error) {
	// Construct new resource manager.
	newManager, files, err := m.applyUpdates(appID, m.AppFS, updates)
	if err != nil {
		return nil, err
	}

	// Validate resource FS by viewing ValidateResource.
	for _, desc := range newManager.Registry.Descriptors {
		_, err := newManager.Read(desc, resource.ValidateResource{})
		// Some resource may not have builtin value, e.g. app_logo_dark.
		if errors.Is(err, resource.ErrResourceNotFound) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("invalid resource: %w", err)
		}
	}

	// Validate configuration.
	cfg, err := configsource.LoadConfig(newManager)
	if err != nil {
		return nil, err
	}

	if string(cfg.AppConfig.ID) != appID {
		return nil, fmt.Errorf("invalid resource '%s': incorrect app ID", configsource.AuthgearYAML)
	}

	// Clean up orphaned resources if authgear.yaml is updated.
	// It is because the portal updates the resources, and then
	// update authgear.yaml in 2 consecutive calls.
	// If we cleans up unconditionally, we cannot save new Deno hooks.
	for _, update := range updates {
		if update.Path == configsource.AuthgearYAML {
			filesToDelete, err := m.cleanupOrphanedResources(newManager, cfg)
			if err != nil {
				return nil, err
			}

			if len(filesToDelete) > 0 {
				files = append(files, filesToDelete...)
			}
		}
	}

	return files, nil
}

func (m *Manager) cleanupOrphanedResources(manager *resource.Manager, cfg *config.Config) ([]*resource.ResourceFile, error) {
	paths := make(map[string]struct{})

	addToPaths := func(urlStr string) error {
		u, err := url.Parse(urlStr)
		if err != nil {
			return err
		}
		if u.Scheme == "authgeardeno" {
			key := strings.TrimPrefix(u.Path, "/")
			paths[key] = struct{}{}
		}
		return nil
	}

	for _, h := range cfg.AppConfig.Hook.BlockingHandlers {
		err := addToPaths(h.URL)
		if err != nil {
			return nil, err
		}
	}
	for _, h := range cfg.AppConfig.Hook.NonBlockingHandlers {
		err := addToPaths(h.URL)
		if err != nil {
			return nil, err
		}
	}
	customSMSProviderCfg := cfg.SecretConfig.GetCustomSMSProviderConfig()
	if customSMSProviderCfg != nil {
		err := addToPaths(customSMSProviderCfg.URL)
		if err != nil {
			return nil, err
		}
	}

	var filesToDelete []*resource.ResourceFile
	for _, fs := range manager.Filesystems() {
		if fs.GetFsLevel() == resource.FsLevelApp {
			locations, err := hook.DenoFile.FindResources(fs)
			if err != nil {
				return nil, err
			}

			for _, location := range locations {

				_, ok := paths[location.Path]
				// No longer referenced by the config, i.e. orphaned.
				if !ok {
					l := location
					filesToDelete = append(filesToDelete, &resource.ResourceFile{
						Location: l,
						Data:     nil,
					})
				}
			}
		}
	}

	return filesToDelete, nil
}

func (m *Manager) getFromAppFs(newAppFs resource.LeveledAferoFs, location resource.Location) (*resource.ResourceFile, error) {
	f, err := newAppFs.Fs.Open(location.Path)
	if os.IsNotExist(err) {
		return &resource.ResourceFile{
			Location: location,
			Data:     nil,
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return &resource.ResourceFile{
		Location: location,
		Data:     data,
	}, nil
}

func (m *Manager) getFromAllFss(desc resource.Descriptor) ([]resource.ResourceFile, error) {
	var locations []resource.Location
	for _, fs := range m.AppResourceManager.Fs {
		ls, err := desc.FindResources(fs)
		if err != nil {
			return nil, err
		}
		locations = append(locations, ls...)
	}

	files := make([]resource.ResourceFile, len(locations))
	for idx, location := range locations {
		data, err := resource.ReadLocation(location)
		if err != nil {
			return nil, err
		}
		files[idx] = resource.ResourceFile{
			Location: location,
			Data:     data,
		}
	}

	return files, nil
}

func (m *Manager) applyUpdates(appID string, appFs resource.Fs, updates []Update) (*resource.Manager, []*resource.ResourceFile, error) {
	manager := m.AppResourceManager

	newFs, err := cloneFS(appFs)
	if err != nil {
		return nil, nil, err
	}

	newAppFs := resource.LeveledAferoFs{Fs: newFs, FsLevel: resource.FsLevelApp}

	var files []*resource.ResourceFile
	for _, u := range updates {
		location := resource.Location{
			Fs:   newAppFs,
			Path: u.Path,
		}

		// Retrieve the original file.
		resrc, err := m.getFromAppFs(newAppFs, location)
		if err != nil {
			return nil, nil, err
		}

		if u.Checksum != "" && checksum.CRC32IEEEInHex(resrc.Data) != u.Checksum {
			msg := fmt.Sprintf("resource update conflict: %v", u.Path)
			return nil, nil, ResourceUpdateConflict.NewWithInfo(msg, apierrors.Details{"path": u.Path})
		}

		desc, ok := manager.Resolve(u.Path)
		if !ok {
			err = fmt.Errorf("invalid resource '%s': unknown resource path", resrc.Location.Path)
			return nil, nil, err
		}

		// Validate file size
		sizeLimit := ConfigFileMaxSize
		if sizeLimitDescriptor, ok := desc.(resource.SizeLimitDescriptor); ok {
			sizeLimit = sizeLimitDescriptor.GetSizeLimit()
		}
		if len(u.Data) > sizeLimit {
			message := fmt.Sprintf("invalid resource '%s': too large (%v > %v)", u.Path, len(u.Data), sizeLimit)
			err := ResouceTooLarge.NewWithInfo(message, apierrors.Details{"size": len(u.Data), "max_size": sizeLimit, "path": u.Path})
			return nil, nil, err
		}

		// Retrieve the file in all FSs.
		all, err := m.getFromAllFss(desc)
		if err != nil {
			return nil, nil, err
		}

		ctx := m.Context
		ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, m.AppFeatureConfig)
		ctx = context.WithValue(ctx, configsource.ContextKeyClock, m.Clock)
		ctx = context.WithValue(ctx, configsource.ContextKeyAppHostSuffixes, m.AppHostSuffixes)
		ctx = context.WithValue(ctx, configsource.ContextKeyDomainService, m.DomainService)
		ctx = context.WithValue(ctx, hook.ContextKeyDenoClient, m.DenoClient)

		err = m.Tutorials.OnUpdateResource(ctx, appID, all, resrc, u.Data)
		if err != nil {
			return nil, nil, err
		}

		resrc, err = desc.UpdateResource(ctx, all, resrc, u.Data)
		if err != nil {
			return nil, nil, err
		}

		if resrc.Data == nil {
			_ = newFs.Remove(resrc.Location.Path)
		} else {
			_ = newFs.MkdirAll(path.Dir(resrc.Location.Path), 0666)
			_ = afero.WriteFile(newFs, resrc.Location.Path, resrc.Data, 0666)
		}

		files = append(files, resrc)
	}

	var newResFs []resource.Fs
	for _, fs := range manager.Fs {
		if fs == appFs {
			newResFs = append(newResFs, newAppFs)
		} else {
			newResFs = append(newResFs, fs)
		}
	}
	return resource.NewManager(manager.Registry, newResFs), files, nil
}
