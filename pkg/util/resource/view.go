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
type View interface {
	view()
}

// AppFileView is an view on the resources at specific path in the App FS.
// Since the path is specific, so the view is single-locale.
type AppFileView interface {
	View
	AppFilePath() string
	SecretKeyAllowlist() []string
}

// EffectiveFileView is an view on the resources at specific path in all FSs.
// Since the path is specific, so the view is single-locale.
type EffectiveFileView interface {
	View
	DefaultLanguageTag() string
	EffectiveFilePath() string
}

// EffectiveResourceView is an view on the resources in all FSs.
// Since there is no path, the view is locale-resolved.
type EffectiveResourceView interface {
	View
	PreferredLanguageTags() []string
	DefaultLanguageTag() string
}

type AppFile struct {
	Path              string
	AllowedSecretKeys []string
}

var _ AppFileView = AppFile{}

func (v AppFile) view() {}
func (v AppFile) AppFilePath() string {
	return v.Path
}
func (v AppFile) SecretKeyAllowlist() []string {
	return v.AllowedSecretKeys
}

type EffectiveFile struct {
	DefaultTag string
	Path       string
}

var _ EffectiveFileView = EffectiveFile{}

func (v EffectiveFile) view() {}
func (v EffectiveFile) DefaultLanguageTag() string {
	return v.DefaultTag
}
func (v EffectiveFile) EffectiveFilePath() string {
	return v.Path
}

type EffectiveResource struct {
	PreferredTags []string
	DefaultTag    string
}

var _ EffectiveResourceView = EffectiveResource{}

func (v EffectiveResource) view() {}
func (v EffectiveResource) PreferredLanguageTags() []string {
	return v.PreferredTags
}
func (v EffectiveResource) DefaultLanguageTag() string {
	return v.DefaultTag
}
