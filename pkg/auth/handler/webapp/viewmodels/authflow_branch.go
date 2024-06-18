package viewmodels

import (
	"slices"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthflowBranch struct {
	Authentication        config.AuthenticationFlowAuthentication
	Index                 int
	Channel               model.AuthenticatorOOBChannel
	MaskedClaimValue      string
	OTPForm               otp.Form
	VerificationSkippable bool
}

func isAuthflowBranchSame(a AuthflowBranch, b AuthflowBranch) bool {
	return a.Index == b.Index && a.Channel == b.Channel
}

type AuthflowBranchViewModel struct {
	// FlowType is mainly for pages to tell if the flow is reauthentication or not.
	FlowType           authflow.FlowType
	ActionType         authflow.FlowActionType
	DeviceTokenEnabled bool
	Branches           []AuthflowBranch
}

type AuthflowBranchFilter func([]AuthflowBranch) []AuthflowBranch

func NewAuthflowBranchViewModel(
	screen *webapp.AuthflowScreenWithFlowResponse,
	filters ...AuthflowBranchFilter) AuthflowBranchViewModel {
	branchFlowResponse := screen.BranchStateTokenFlowResponse

	deviceTokenEnabled := false
	var branches []AuthflowBranch
	if branchFlowResponse != nil {
		switch branchData := branchFlowResponse.Action.Data.(type) {
		case declarative.StepAuthenticateData:
			deviceTokenEnabled = branchData.DeviceTokenEnabled
			branches = newAuthflowBranchViewModelStepAuthenticate(screen, branchData)
		case declarative.IntentSignupFlowStepCreateAuthenticatorData:
			branches = newAuthflowBranchViewModelStepCreateAuthenticator(screen, branchData)
		case declarative.SelectOOBOTPChannelsData:
			branches = newAuthflowBranchViewModelVerify(screen, branchData)
		}
	}

	for _, filter := range filters {
		branches = filter(branches)
	}

	branches = reorderBranches(branches)

	return AuthflowBranchViewModel{
		FlowType:           screen.StateTokenFlowResponse.Type,
		ActionType:         screen.StateTokenFlowResponse.Action.Type,
		DeviceTokenEnabled: deviceTokenEnabled,
		Branches:           branches,
	}
}

func NewInlinePreviewAuthflowBranchViewModel() AuthflowBranchViewModel {
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionTypeFinished,
		DeviceTokenEnabled: false,
		Branches:           []AuthflowBranch{},
	}
}

func reorderBranches(branches []AuthflowBranch) []AuthflowBranch {
	// Put passkey branches at the end
	result := []AuthflowBranch{}
	for idx := range branches {
		b := branches[len(branches)-idx-1]
		if b.Authentication == config.AuthenticationFlowAuthenticationPrimaryPasskey {
			// Put to the beginning of the reversed slice
			result = append([]AuthflowBranch{b}, result...)
		} else {
			result = append(result, b)
		}
	}

	slices.Reverse(result)
	return result
}

func newAuthflowBranchViewModelStepAuthenticate(screen *webapp.AuthflowScreenWithFlowResponse, branchData declarative.StepAuthenticateData) []AuthflowBranch {
	takenBranchIndex := *screen.Screen.TakenBranchIndex
	takenBranch := AuthflowBranch{
		Authentication: branchData.Options[takenBranchIndex].Authentication,
		Index:          takenBranchIndex,
		Channel:        screen.Screen.TakenChannel,
	}

	branches := []AuthflowBranch{}

	addIndexBranch := func(idx int, o declarative.AuthenticateOptionForOutput) {
		branch := AuthflowBranch{
			Authentication: o.Authentication,
			Index:          idx,
		}
		if !isAuthflowBranchSame(branch, takenBranch) {
			branches = append(branches, branch)
		}
	}

	addChannelBranch := func(idx int, o declarative.AuthenticateOptionForOutput) {
		for _, channel := range o.Channels {
			branch := AuthflowBranch{
				Authentication:   o.Authentication,
				Index:            idx,
				Channel:          channel,
				MaskedClaimValue: o.MaskedDisplayName,
				OTPForm:          o.OTPForm,
			}
			if !isAuthflowBranchSame(branch, takenBranch) {
				branches = append(branches, branch)
			}
		}
	}

	for idx, o := range branchData.Options {
		switch o.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			addIndexBranch(idx, o)
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			addIndexBranch(idx, o)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			addChannelBranch(idx, o)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			addChannelBranch(idx, o)
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			addIndexBranch(idx, o)
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			addChannelBranch(idx, o)
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			addChannelBranch(idx, o)
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			addIndexBranch(idx, o)
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			addIndexBranch(idx, o)
		default:
			// Ignore other authentications.
			break
		}
	}

	return branches
}

func newAuthflowBranchViewModelStepCreateAuthenticator(screen *webapp.AuthflowScreenWithFlowResponse, branchData declarative.IntentSignupFlowStepCreateAuthenticatorData) []AuthflowBranch {
	takenBranchIndex := *screen.Screen.TakenBranchIndex
	takenBranch := AuthflowBranch{
		Authentication: branchData.Options[takenBranchIndex].Authentication,
		Index:          takenBranchIndex,
		Channel:        screen.Screen.TakenChannel,
	}

	branches := []AuthflowBranch{}

	addIndexBranch := func(idx int, o declarative.CreateAuthenticatorOption) {
		branch := AuthflowBranch{
			Authentication: o.Authentication,
			Index:          idx,
		}
		if !isAuthflowBranchSame(branch, takenBranch) {
			branches = append(branches, branch)
		}
	}

	addChannelBranch := func(idx int, o declarative.CreateAuthenticatorOption) {
		for _, channel := range o.Channels {
			branch := AuthflowBranch{
				Authentication:   o.Authentication,
				Index:            idx,
				Channel:          channel,
				MaskedClaimValue: "",
				OTPForm:          o.OTPForm,
			}
			if !isAuthflowBranchSame(branch, takenBranch) {
				branches = append(branches, branch)
			}
		}
	}

	addSkipBranch := func(idx int, o declarative.CreateAuthenticatorOption) {
		branch := AuthflowBranch{
			Authentication:        o.Authentication,
			Index:                 idx,
			MaskedClaimValue:      o.Target.MaskedDisplayName,
			OTPForm:               o.OTPForm,
			VerificationSkippable: true,
		}
		if !isAuthflowBranchSame(branch, takenBranch) {
			branches = append(branches, branch)
		}
	}

	for idx, o := range branchData.Options {
		switch o.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			addIndexBranch(idx, o)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			if o.Target != nil && !o.Target.VerificationRequired {
				addSkipBranch(idx, o)
			} else {
				addChannelBranch(idx, o)
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			addIndexBranch(idx, o)
		default:
			// Ignore other authentications.
			break
		}
	}

	return branches
}

func newAuthflowBranchViewModelVerify(screen *webapp.AuthflowScreenWithFlowResponse, branchData declarative.SelectOOBOTPChannelsData) []AuthflowBranch {
	takenBranch := AuthflowBranch{
		Channel: screen.Screen.TakenChannel,
	}

	branches := []AuthflowBranch{}

	for _, channel := range branchData.Channels {
		branch := AuthflowBranch{
			Channel:          channel,
			MaskedClaimValue: branchData.MaskedClaimValue,
		}
		if !isAuthflowBranchSame(branch, takenBranch) {
			branches = append(branches, branch)
		}
	}

	return branches
}
