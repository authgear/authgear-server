package tutorial

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type Store interface {
	Get(appID string) (*Entry, error)
	Save(entry *Entry) error
}

type Service struct {
	GlobalDatabase *globaldb.Handle
	Store          Store
}

// Get acquires connection.
func (s *Service) Get(appID string) (*Entry, error) {
	var entry *Entry
	var err error
	err = s.GlobalDatabase.WithTx(func() error {
		entry, err = s.Store.Get(appID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Skip acquires connection.
func (s *Service) Skip(appID string) (err error) {
	err = s.GlobalDatabase.WithTx(func() error {
		entry, err := s.Store.Get(appID)
		if err != nil {
			return err
		}

		entry.Skip()

		err = s.Store.Save(entry)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return
	}

	return
}

// RecordProgresses acquires connection.
func (s *Service) RecordProgresses(appID string, ps []Progress) (err error) {
	err = s.GlobalDatabase.WithTx(func() error {
		return s.recordProgresses(appID, ps)
	})
	if err != nil {
		return
	}

	return
}

// OnUpdateResource0 assumes acquired connection.
func (s *Service) OnUpdateResource0(appID string, resourcesInAllFss []resource.ResourceFile, resourceInTargetFs *resource.ResourceFile, data []byte) (err error) {
	ps, err := s.detectProgresses(resourceInTargetFs, data)
	if err != nil {
		return
	}

	return s.recordProgresses(appID, ps)
}

func (s *Service) recordProgresses(appID string, ps []Progress) (err error) {
	entry, err := s.Store.Get(appID)
	if err != nil {
		return
	}

	entry.AddProgress(ps)

	err = s.Store.Save(entry)
	if err != nil {
		return
	}

	return
}

func (s *Service) detectProgresses(resourceInTargetFs *resource.ResourceFile, data []byte) (out []Progress, err error) {
	ps, err := s.detectAuthgearYAML(resourceInTargetFs, data)
	if err != nil {
		return
	}
	out = append(out, ps...)

	ps = s.detectUIChanges(resourceInTargetFs, data)
	out = append(out, ps...)

	return
}
func (s *Service) detectAuthgearYAML(resourceInTargetFs *resource.ResourceFile, data []byte) (out []Progress, err error) {
	d := configsource.AuthgearYAMLDescriptor{}
	_, ok := d.MatchResource(resourceInTargetFs.Location.Path)
	if !ok {
		return
	}

	original, err := config.Parse(resourceInTargetFs.Data)
	if err != nil {
		return
	}

	incoming, err := config.Parse(data)
	if err != nil {
		return
	}

	if len(incoming.OAuth.Clients) > len(original.OAuth.Clients) {
		out = append(out, ProgressCreateApplication)
	}

	if len(incoming.Identity.OAuth.Providers) > len(original.Identity.OAuth.Providers) {
		out = append(out, ProgressSSO)
	}

	return
}

func (s *Service) detectUIChanges(resourceInTargetFs *resource.ResourceFile, data []byte) (out []Progress) {
	detected := false
	ds := []resource.Descriptor{
		web.AuthgearLightThemeCSS,
		web.AuthgearDarkThemeCSS,
		web.AppLogo,
		web.AppLogoDark,
		web.Favicon,
		template.TranslationJSON,
	}
	for _, d := range ds {
		_, ok := d.MatchResource(resourceInTargetFs.Location.Path)
		if ok {
			detected = true
		}
	}
	if detected {
		out = append(out, ProgressCustomizeUI)
	}

	return
}
