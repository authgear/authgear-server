package appresource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/util/checksum"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

//go:generate go tool mockgen -source=manager.go -destination=manager_mock_test.go -package appresource_test

const ConfigFileMaxSize = 100 * 1024

type DenoClient interface {
	Check(ctx context.Context, snippet string) error
}

type TutorialService interface {
	OnUpdateResource0(ctx context.Context, appID string, resourcesInAllFss []resource.ResourceFile, resourceInTargetFs *resource.ResourceFile, data []byte) (err error)
}

type DomainService interface {
	ListDomains(ctx context.Context, appID string) ([]*apimodel.Domain, error)
}

type Manager struct {
	AppResourceManager    *resource.Manager
	AppFS                 resource.Fs
	AppFeatureConfig      *config.FeatureConfig
	AppHostSuffixes       *config.AppHostSuffixes
	DomainService         DomainService
	Tutorials             TutorialService
	DenoClient            DenoClient
	Clock                 clock.Clock
	SAMLEnvironmentConfig config.SAMLEnvironmentConfig
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

func (m *Manager) ReadAppFile(ctx context.Context, desc resource.Descriptor, view resource.AppFileView) (any, error) {
	return m.AppResourceManager.Read(ctx, desc, view)
}

func (m *Manager) ReadEffectiveFile(ctx context.Context, desc resource.Descriptor, view resource.EffectiveResourceView) (any, error) {
	return m.AppResourceManager.Read(ctx, desc, view)
}

// ApplyUpdates0 assume acquired connection.
func (m *Manager) ApplyUpdates0(ctx context.Context, appID string, updates []Update) ([]*resource.ResourceFile, error) {
	// Construct new resource manager.
	newManager, files, err := m.applyUpdates(ctx, appID, m.AppFS, updates)
	if err != nil {
		return nil, err
	}

	// Validate resource FS by viewing ValidateResource.
	for _, desc := range newManager.Registry.Descriptors {
		_, err := newManager.Read(ctx, desc, resource.ValidateResource{})
		// Some resource may not have builtin value, e.g. app_logo_dark.
		if errors.Is(err, resource.ErrResourceNotFound) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("invalid resource: %w", err)
		}
	}

	// Validate configuration.
	cfg, err := configsource.LoadConfig(ctx, newManager)
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

			secretFile, err := m.cleanupOrphanedSecrets(ctx, newManager, cfg)
			if err != nil {
				return nil, err
			}
			if secretFile != nil {
				files = append(files, secretFile)
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

// cleanupOrphanedSecrets prunes telemetry.audit_logs.streams.tls items whose
// stream_name no longer matches a stream in authgear.yaml. It is not an
// error for such an item to exist -- config.SecretConfig.Validate tolerates
// it -- but a save is the moment the project's own state stops needing it.
func (m *Manager) cleanupOrphanedSecrets(ctx context.Context, manager *resource.Manager, cfg *config.Config) (*resource.ResourceFile, error) {
	keep := make(map[string]struct{})
	if cfg.AppConfig.Telemetry != nil && cfg.AppConfig.Telemetry.AuditLogs != nil {
		for _, stream := range cfg.AppConfig.Telemetry.AuditLogs.Streams {
			keep[stream.Name] = struct{}{}
		}
	}

	var location *resource.Location
	for _, fs := range manager.Filesystems() {
		if fs.GetFsLevel() == resource.FsLevelApp {
			l := resource.Location{Fs: fs, Path: configsource.AuthgearSecretYAML}
			location = &l
		}
	}
	if location == nil {
		return nil, nil
	}

	data, err := resource.ReadLocation(*location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Use the non-validating parser: the document being pruned may be one
	// that a validating parse would reject, e.g. because of the very
	// orphan this function is about to remove.
	secretConfig, err := config.ParsePartialSecret(ctx, data)
	if err != nil {
		return nil, err
	}

	idx, item, found := secretConfig.Lookup(config.TelemetryAuditLogStreamTLSMaterialsKey)
	if !found {
		return nil, nil
	}

	// Decode item.RawData directly, rather than using item.Data. item.Data
	// has been through config.SetFieldDefaults, which -- exactly like it
	// does for AppConfig -- materializes every nil struct pointer,
	// including ClientCertificate and its nested Key *JWK when
	// client_certificate was absent from the document. JWK.MarshalJSON
	// panics on that materialized-but-empty value, so re-marshaling
	// item.Data here would panic for exactly the certificate_authority-only
	// items the spec expects to be common. item.RawData is untouched by
	// SetFieldDefaults and re-marshals safely.
	var materials config.TelemetryAuditLogStreamTLSMaterials
	if err := json.Unmarshal(item.RawData, &materials); err != nil {
		return nil, err
	}

	var filtered config.TelemetryAuditLogStreamTLSMaterials
	for _, material := range materials {
		if _, ok := keep[material.StreamName]; ok {
			filtered = append(filtered, material)
		}
	}

	// Nothing was orphaned; do not rewrite the file.
	if len(filtered) == len(materials) {
		return nil, nil
	}

	if len(filtered) == 0 {
		secretConfig.Secrets = append(secretConfig.Secrets[:idx], secretConfig.Secrets[idx+1:]...)
	} else {
		jsonData, err := json.Marshal(filtered)
		if err != nil {
			return nil, err
		}
		secretConfig.Secrets[idx] = config.SecretItem{
			Key:     config.TelemetryAuditLogStreamTLSMaterialsKey,
			RawData: json.RawMessage(jsonData),
		}
	}

	updatedYAML, err := yaml.Marshal(secretConfig)
	if err != nil {
		return nil, err
	}

	return &resource.ResourceFile{
		Location: *location,
		Data:     updatedYAML,
	}, nil
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

	data, err := io.ReadAll(f)
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

func (m *Manager) applyUpdates(ctx context.Context, appID string, appFs resource.Fs, updates []Update) (*resource.Manager, []*resource.ResourceFile, error) {
	ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, m.AppFeatureConfig)
	ctx = context.WithValue(ctx, configsource.ContextKeyClock, m.Clock)
	ctx = context.WithValue(ctx, configsource.ContextKeyAppHostSuffixes, m.AppHostSuffixes)
	ctx = context.WithValue(ctx, configsource.ContextKeyDomainService, m.DomainService)
	ctx = context.WithValue(ctx, configsource.ContextKeySAMLEntityID, m.renderSAMLEntityID(appID))
	ctx = context.WithValue(ctx, hook.ContextKeyDenoClient, m.DenoClient)

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

		err = m.Tutorials.OnUpdateResource0(ctx, appID, all, resrc, u.Data)
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

func (m *Manager) renderSAMLEntityID(appID string) string {
	return saml.RenderSAMLEntityID(m.SAMLEnvironmentConfig, appID)
}
