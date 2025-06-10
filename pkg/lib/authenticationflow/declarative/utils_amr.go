package declarative

import (
	"context"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

func collectAMRFromNode(node authflow.NodeOrIntent, amr []string, usedAuthenticatorIDs, usedRecoveryCodeIDs setutil.Set[string]) []string {
	if n, ok := node.(MilestoneDidAuthenticate); ok {
		amr = append(amr, n.MilestoneDidAuthenticate()...)
		if authInfo, ok := n.MilestoneDidAuthenticateAuthenticator(); ok && authInfo != nil {
			usedAuthenticatorIDs[authInfo.ID] = struct{}{}
		}
	}
	if n, ok := node.(MilestoneDoCreateAuthenticator); ok {
		info, ok := n.MilestoneDoCreateAuthenticator()
		if ok {
			amr = append(amr, info.AMR()...)
			usedAuthenticatorIDs[info.ID] = struct{}{}
		}
	}
	if n, ok := node.(MilestoneDidConsumeRecoveryCode); ok {
		rc := n.MilestoneDidConsumeRecoveryCode()
		if rc != nil {
			usedRecoveryCodeIDs[rc.ID] = struct{}{}
		}
	}
	return amr
}

func collectAMR(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (amr []string, err error) {
	usedAuthenticatorIDs := setutil.Set[string]{}
	usedRecoveryCodeIDs := setutil.Set[string]{}

	err = authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			amr = collectAMRFromNode(nodeSimple, amr, usedAuthenticatorIDs, usedRecoveryCodeIDs)
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			amr = collectAMRFromNode(intent, amr, usedAuthenticatorIDs, usedRecoveryCodeIDs)
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

func findAMRConstraints(flows authflow.Flows) ([]string, bool) {
	var constraints []string
	found := false

	_ = authflow.TraverseFlowIntentFirst(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneConstraintsProvider); ok {
				if c := n.MilestoneConstraintsProvider(); c != nil {
					constraints = c.AMR
					found = true
				}
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneConstraintsProvider); ok {
				if c := i.MilestoneConstraintsProvider(); c != nil {
					constraints = c.AMR
					found = true
				}
			}
			return nil
		},
	}, flows.Root)

	return constraints, found
}

func RemainingAMRConstraintsInFlow(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]string, error) {
	amrConstraints, found := findAMRConstraints(flows)
	if !found {
		return []string{}, nil
	}
	currentAMRs, err := collectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	remainingContrains := remainingAMRConstraints(amrConstraints, currentAMRs)
	return remainingContrains, nil
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

type AMROption interface {
	GetAMR() []string
}

func filterAMROptionsByAMRConstraint[T AMROption](options []T, amrConstraints []string) []T {
	// Special case: mfa can be fulfilled by any authenticators
	if len(amrConstraints) == 1 && amrConstraints[0] == model.AMRMFA {
		return options
	}

	// If there are other constraints, the user can only choose options that can fulfil the remaining constraints
	var newOptions []T = []T{}
	for _, option := range options {
		option := option
		for _, amr := range option.GetAMR() {
			if slice.ContainsString(amrConstraints, amr) {
				newOptions = append(newOptions, option)
				break
			}
		}
	}

	return newOptions
}
