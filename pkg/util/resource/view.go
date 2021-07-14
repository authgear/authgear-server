package resource

// View is an specific view on the resources described by an Descriptor.
//
// Views are defined within this package only,
// while descriptor is meant to be defined outside this package.
// Therefore, when a new descriptor is being introduced,
// the author has to think about what views the descriptor supports.
//
// Most descriptors only support a subset of all defined views.
//
// Some resources has language tag in their path.
//
// View
// |- AppFileView
// |- EffectiveFileView
// |- EffectiveResourceView
// |- ValidateResourceView
type View interface {
	view()
}

// ViewWithConfig is a wrapper that provides config to view
type ViewWithConfig interface {
	View
	SecretKeyAllowlist() []string
}

// AppFileView is an view on the resources at specific path in the App FS.
// Since the path is specific, so the view is single-locale.
type AppFileView interface {
	View
	AppFilePath() string
}

// EffectiveFileView is an view on the resources at specific path in all FSs.
// Since the path is specific, so the view is single-locale.
type EffectiveFileView interface {
	View
	EffectiveFilePath() string
}

// EffectiveResourceView is an view on the resources in all FSs.
// Since there is no path, the view is locale-resolved.
type EffectiveResourceView interface {
	View
	SupportedLanguageTags() []string
	DefaultLanguageTag() string
	PreferredLanguageTags() []string
}

// ValidateResourceView validates the resource itself.
type ValidateResourceView interface {
	View
	validateResource()
}

type AppFileWithConfig struct {
	AppFileView
	AllowedSecretKeys []string
}

func (f AppFileWithConfig) SecretKeyAllowlist() []string {
	return f.AllowedSecretKeys
}

var _ ViewWithConfig = AppFileWithConfig{}

type AppFile struct {
	Path string
}

var _ AppFileView = AppFile{}

func (v AppFile) view() {}
func (v AppFile) AppFilePath() string {
	return v.Path
}

type EffectiveFile struct {
	Path string
}

var _ EffectiveFileView = EffectiveFile{}

func (v EffectiveFile) view() {}
func (v EffectiveFile) EffectiveFilePath() string {
	return v.Path
}

type EffectiveResource struct {
	SupportedTags []string
	DefaultTag    string
	PreferredTags []string
}

var _ EffectiveResourceView = EffectiveResource{}

func (v EffectiveResource) view() {}
func (v EffectiveResource) SupportedLanguageTags() []string {
	return v.SupportedTags
}
func (v EffectiveResource) DefaultLanguageTag() string {
	return v.DefaultTag
}
func (v EffectiveResource) PreferredLanguageTags() []string {
	return v.PreferredTags
}

type ValidateResource struct{}

var _ ValidateResourceView = ValidateResource{}

func (v ValidateResource) view()             {}
func (v ValidateResource) validateResource() {}
