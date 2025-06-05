package declarative

import (
	"context"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

func collectAMR(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (amr []string, err error) {
	usedAuthenticatorIDs := setutil.Set[string]{}
	usedRecoveryCodeIDs := setutil.Set[string]{}

	err = authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				amr = append(amr, n.MilestoneDidAuthenticate()...)
				if authInfo, ok := n.MilestoneDidAuthenticateAuthenticator(); ok && authInfo != nil {
					usedAuthenticatorIDs[authInfo.ID] = struct{}{}
				}
			}
			if n, ok := nodeSimple.(MilestoneDoCreateAuthenticator); ok {
				info := n.MilestoneDoCreateAuthenticator()
				if info != nil {
					amr = append(amr, info.AMR()...)
					usedAuthenticatorIDs[info.ID] = struct{}{}
				}
			}
			if n, ok := nodeSimple.(MilestoneDidConsumeRecoveryCode); ok {
				rc := n.MilestoneDidConsumeRecoveryCode()
				if rc != nil {
					usedRecoveryCodeIDs[rc.ID] = struct{}{}
				}
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				amr = append(amr, i.MilestoneDidAuthenticate()...)
				if authInfo, ok := i.MilestoneDidAuthenticateAuthenticator(); ok && authInfo != nil {
					usedAuthenticatorIDs[authInfo.ID] = struct{}{}
				}
			}
			if i, ok := intent.(MilestoneDoCreateAuthenticator); ok {
				info := i.MilestoneDoCreateAuthenticator()
				if info != nil {
					amr = append(amr, info.AMR()...)
					usedAuthenticatorIDs[info.ID] = struct{}{}
				}
			}
			if i, ok := intent.(MilestoneDidConsumeRecoveryCode); ok {
				rc := i.MilestoneDidConsumeRecoveryCode()
				if rc != nil {
					usedRecoveryCodeIDs[rc.ID] = struct{}{}
				}
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	if len(usedAuthenticatorIDs) > 1 {
		amr = append(amr, model.AMRMFA)
	} else if len(usedRecoveryCodeIDs) > 0 && len(usedAuthenticatorIDs) > 0 {
		// Also count as MFA if the user has used one authenticator AND recovery code
		amr = append(amr, model.AMRMFA)
	}

	amr = slice.Deduplicate(amr)
	sort.Strings(amr)

	return
}

// TODO(tung): Write a unit test for this
func findAMRContraints(flows authflow.Flows) ([]string, bool) {
	var constraints []string
	found := false

	_ = authflow.TraverseFlowIntentFirst(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneContraintsProvider); ok {
				if c := n.MilestoneContraintsProvider(); c != nil {
					constraints = c.AMR
					found = true
				}
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneContraintsProvider); ok {
				if c := i.MilestoneContraintsProvider(); c != nil {
					constraints = c.AMR
					found = true
				}
			}
			return nil
		},
	}, flows.Root)

	return constraints, found
}

func remainingAMRConstraints(constraints []string, amrs []string) []string {
	var unfulfilledConstraints []string
	for _, c := range constraints {
		if !slice.ContainsString(amrs, c) {
			unfulfilledConstraints = append(unfulfilledConstraints, c)
		}
	}
	return unfulfilledConstraints
}

func filterAuthenticateOptionsByAMRConstraint(options []AuthenticateOption, amrConstraints []string) []AuthenticateOption {
	// Special case: mfa can be fulfilled by any authenticators
	if len(amrConstraints) == 1 && amrConstraints[0] == model.AMRMFA {
		return options
	}

	// If there are other contraints, the user can only choose options that can fulfil the remaining contraints
	var newOptions []AuthenticateOption = []AuthenticateOption{}
	for _, option := range options {
		option := option
		for _, amr := range option.AMR {
			if slice.ContainsString(amrConstraints, amr) {
				newOptions = append(newOptions, option)
				break
			}
		}
	}

	return newOptions
}
