package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type AuthflowBranch struct {
	Authentication   authflowclient.Authentication
	Index            int
	Channel          model.AuthenticatorOOBChannel
	MaskedClaimValue string
	OTPForm          otp.Form
}

func isAuthflowBranchSame(a AuthflowBranch, b AuthflowBranch) bool {
	return a.Index == b.Index && a.Channel == b.Channel
}

type AuthflowBranchViewModel struct {
	// FlowType is mainly for pages to tell if the flow is reauthentication or not.
	FlowType           authflowclient.FlowType
	ActionType         authflowclient.FlowActionType
	DeviceTokenEnabled bool
	Branches           []AuthflowBranch
}

func NewAuthflowBranchViewModel(screen *webapp.AuthflowScreenWithFlowResponse) AuthflowBranchViewModel {
	branchFlowResponse := screen.BranchStateTokenFlowResponse

	deviceTokenEnabled := false
	var branches []AuthflowBranch
	if branchFlowResponse != nil {
		dataAuthenticate, dataCreateAuthenticator, dataChannels, err := authflowclient.CastForBranch(branchFlowResponse.Action.Data)
		if err != nil {
			panic(err)
		}

		switch {
		case dataAuthenticate != nil:
			deviceTokenEnabled = dataAuthenticate.DeviceTokenEnabled
			branches = newAuthflowBranchViewModelStepAuthenticate(screen, dataAuthenticate)
		case dataCreateAuthenticator != nil:
			branches = newAuthflowBranchViewModelStepCreateAuthenticator(screen, dataCreateAuthenticator)
		case dataChannels != nil:
			branches = newAuthflowBranchViewModelVerify(screen, dataChannels)
		}
	}

	return AuthflowBranchViewModel{
		FlowType:           screen.StateTokenFlowResponse.Type,
		ActionType:         screen.StateTokenFlowResponse.Action.Type,
		DeviceTokenEnabled: deviceTokenEnabled,
		Branches:           branches,
	}
}

func newAuthflowBranchViewModelStepAuthenticate(screen *webapp.AuthflowScreenWithFlowResponse, branchData *authflowclient.DataAuthenticate) []AuthflowBranch {
	takenBranchIndex := *screen.Screen.TakenBranchIndex
	takenBranch := AuthflowBranch{
		Authentication: branchData.Options[takenBranchIndex].Authentication,
		Index:          takenBranchIndex,
		Channel:        screen.Screen.TakenChannel,
	}

	branches := []AuthflowBranch{}

	addIndexBranch := func(idx int, o authflowclient.DataAuthenticateOption) {
		branch := AuthflowBranch{
			Authentication: o.Authentication,
			Index:          idx,
		}
		if !isAuthflowBranchSame(branch, takenBranch) {
			branches = append(branches, branch)
		}
	}

	addChannelBranch := func(idx int, o authflowclient.DataAuthenticateOption) {
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
		case authflowclient.AuthenticationPrimaryPassword:
			addIndexBranch(idx, o)
		case authflowclient.AuthenticationPrimaryPasskey:
			addIndexBranch(idx, o)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationSecondaryPassword:
			addIndexBranch(idx, o)
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationRecoveryCode:
			addIndexBranch(idx, o)
		default:
			// Ignore other authentications.
			break
		}
	}

	return branches
}

func newAuthflowBranchViewModelStepCreateAuthenticator(screen *webapp.AuthflowScreenWithFlowResponse, branchData *authflowclient.DataCreateAuthenticator) []AuthflowBranch {
	takenBranchIndex := *screen.Screen.TakenBranchIndex
	takenBranch := AuthflowBranch{
		Authentication: branchData.Options[takenBranchIndex].Authentication,
		Index:          takenBranchIndex,
		Channel:        screen.Screen.TakenChannel,
	}

	branches := []AuthflowBranch{}

	addIndexBranch := func(idx int, o authflowclient.DataCreateAuthenticatorOption) {
		branch := AuthflowBranch{
			Authentication: o.Authentication,
			Index:          idx,
		}
		if !isAuthflowBranchSame(branch, takenBranch) {
			branches = append(branches, branch)
		}
	}

	addChannelBranch := func(idx int, o authflowclient.DataCreateAuthenticatorOption) {
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

	for idx, o := range branchData.Options {
		switch o.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			addIndexBranch(idx, o)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationSecondaryPassword:
			addIndexBranch(idx, o)
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			addChannelBranch(idx, o)
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
			addChannelBranch(idx, o)
		default:
			// Ignore other authentications.
			break
		}
	}

	return branches
}

func newAuthflowBranchViewModelVerify(screen *webapp.AuthflowScreenWithFlowResponse, branchData *authflowclient.DataChannels) []AuthflowBranch {
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
