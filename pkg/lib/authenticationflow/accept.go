package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/botprotection"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type AcceptResult struct {
	BotProtectionVerificationResult *BotProtectionVerificationResult `json:"bot_protection,omitempty"`
	DelayedOneTimeFunctions         []DelayedOneTimeFunction         `json:"-"`
}

func NewAcceptResult() *AcceptResult {
	return &AcceptResult{
		DelayedOneTimeFunctions: []DelayedOneTimeFunction{},
	}
}

const (
	MAX_LOOP = 100
)

// Accept executes the flow to the deepest using input.
// In addition to the errors caused by intents and nodes,
// ErrEOF and ErrNoChange can be returned.
func Accept(ctx context.Context, deps *Dependencies, flows Flows, result *AcceptResult, rawMessage json.RawMessage) error {
	return accept(ctx, deps, flows, result, func(inputSchema InputSchema) (Input, error) {
		if rawMessage != nil && inputSchema != nil {
			input, err := inputSchema.MakeInput(ctx, rawMessage)
			if err != nil {
				return nil, err
			}
			return input, nil
		}
		return nil, nil
	})
}

func AcceptSyntheticInput(ctx context.Context, deps *Dependencies, flows Flows, result *AcceptResult, syntheticInput Input) error {
	return accept(ctx, deps, flows, result, func(inputSchema InputSchema) (Input, error) {
		return syntheticInput, nil
	})
}

func accept(ctx context.Context, deps *Dependencies, flows Flows, result *AcceptResult, inputFn func(inputSchema InputSchema) (Input, error)) error {
	err := doAccept(ctx, deps, flows, result, inputFn)
	if err != nil {
		err = logAuthenticationBlockedErrorIfNeeded(ctx, deps, flows, err)
		if err != nil {
			return err
		}
	}
	return err
}

func logAuthenticationBlockedErrorIfNeeded(ctx context.Context, deps *Dependencies, flows Flows, err error) error {
	if !apierrors.IsAPIError(err) {
		return err
	}
	apiErr := apierrors.AsAPIError(err)
	if !user.IsAccountStatusError(apiErr) &&
		!apierrors.IsKind(apiErr, lockout.AccountLockout) &&
		!apierrors.IsKind(apiErr, hook.WebHookDisallowed) {
		return err
	}

	userID, getUserIDErr := GetUserID(flows)
	if getUserIDErr != nil {
		if errors.Is(getUserIDErr, ErrNoUserID) || errors.Is(getUserIDErr, ErrDifferentUserID) {
			userID = ""
		} else {
			return errors.Join(getUserIDErr, err)
		}
	}
	var user *model.User
	if userID != "" {
		u, getUserErr := deps.Users.Get(ctx, userID, accesscontrol.RoleGreatest)
		if getUserErr != nil {
			return errors.Join(getUserErr, err)
		}
		user = u
	}
	dispatchErr := deps.Events.DispatchEventImmediately(ctx, &nonblocking.AuthenticationBlockedEventPayload{
		User:  user,
		Error: apiErr,
	})
	if dispatchErr != nil {
		ServiceLogger.GetLogger(ctx).WithError(dispatchErr).Error(ctx, "failed to dispatch event")
	}
	return err
}

// nolint: gocognit
func doAccept(ctx context.Context, deps *Dependencies, flows Flows, result *AcceptResult, inputFn func(inputSchema InputSchema) (Input, error)) (err error) {
	var changed bool

	defer func() {
		if changed {
			flows.Nearest.StateToken = newStateToken()
		}
		if !changed && err == nil {
			err = ErrNoChange
		}
	}()

	loopCount := 0

	var nextNodeType string
	for {
		loopCount += 1
		if loopCount > MAX_LOOP {
			panic(fmt.Errorf("number of loops reached limit. next node is %s", nextNodeType))
		}
		var findInputReactorResult *FindInputReactorResult
		findInputReactorResult, err = FindInputReactor(ctx, deps, flows)
		if err != nil {
			return
		}

		// Otherwise we found an InputReactor that we can feed input to.

		// input by default is nil.
		var input Input
		input, err = inputFn(findInputReactorResult.InputSchema)
		// As a special case, if this loop has iterated at least once,
		// then we treat the validation error as ErrIncompatibleInput,
		// by setting err = nil.
		var valiationError *validation.AggregatedError
		if errors.As(err, &valiationError) {
			if changed {
				err = nil
			}
			return
		} else if err != nil {
			return
		}

		var reactToResult ReactToResult
		reactToResult, err = findInputReactorResult.InputReactor.ReactTo(ctx, deps, findInputReactorResult.Flows, input)

		// Handle err == ErrIncompatibleInput
		if errors.Is(err, ErrIncompatibleInput) {
			err = nil
			// Since this is a forever loop, changed may have been set to true already.
			return
		}

		// Handle err == ErrSameNode
		if errors.Is(err, ErrSameNode) {
			err = nil
			// We still consider the flow has something changes.
			changed = true
			// We have to stop and return here because this input reactor will react to this input indefinitely.
			return
		}

		// Handle err == ErrSameNode
		if errors.Is(err, ErrReplaceNode) {
			err = nil
			// We still consider the flow has something changes.
			changed = true

			var nodeToReplace *Node
			switch reactToResult := reactToResult.(type) {
			case *Node:
				nodeToReplace = reactToResult
			case *NodeWithDelayedOneTimeFunction:
				nodeToReplace = reactToResult.Node
				result.DelayedOneTimeFunctions = append(result.DelayedOneTimeFunctions, reactToResult.DelayedOneTimeFunction)
			default:
				panic(fmt.Errorf("failed to update node: uxepected type of ReactToResult %t", reactToResult))
			}

			// precondition: ErrUpdateNode requires at least one node.
			if len(findInputReactorResult.Flows.Nearest.Nodes) == 0 {
				panic(fmt.Errorf("input reactor %T returned ErrUpdateNode, but there are no nodes", findInputReactorResult.InputReactor))
			}

			// Update the node inplace.
			findInputReactorResult.Flows.Nearest.Nodes[len(findInputReactorResult.Flows.Nearest.Nodes)-1] = *nodeToReplace

			// We have to stop and return here because this edge will react to this input indefinitely.
			return
		}

		// Handle ErrBotProtectionVerification
		var errBotProtectionVerification *ErrorBotProtectionVerification
		if errors.As(err, &errBotProtectionVerification) {
			_, notMatched := errorutil.Partition(err, func(err error) bool {
				var _errBPV *ErrorBotProtectionVerification
				return errors.As(err, &_errBPV) && _errBPV.Status == ErrorBotProtectionVerificationStatusSuccess
			})
			err = notMatched

			switch errBotProtectionVerification.Status {
			case ErrorBotProtectionVerificationStatusSuccess:
				result.BotProtectionVerificationResult = &BotProtectionVerificationResult{
					Outcome: BotProtectionVerificationOutcomeVerified,
				}
			case ErrorBotProtectionVerificationStatusFailed:
				// We still consider the flow has something changes.
				changed = true
				result.BotProtectionVerificationResult = &BotProtectionVerificationResult{
					Outcome: BotProtectionVerificationOutcomeFailed,
				}
				err = botprotection.ErrVerificationFailed
				return
			case ErrorBotProtectionVerificationStatusServiceUnavailable:
				// We still consider the flow has something changes.
				changed = true
				result.BotProtectionVerificationResult = &BotProtectionVerificationResult{
					Outcome: BotProtectionVerificationOutcomeFailed,
				}
				err = botprotection.ErrVerificationServiceUnavailable
				return
			default:
				// unrecognized status
				panic("unrecognized bot protection special error status in accept loop")
			}
		}

		// Handle other error.
		if err != nil {
			err = newAuthenticationFlowError(flows, err)

			return
		}

		// We need to append the nextNode to the closest flow.
		var nextNode Node
		switch reactToResult := reactToResult.(type) {
		case *Node:
			nextNode = *reactToResult
		case *NodeWithDelayedOneTimeFunction:
			nextNode = *reactToResult.Node
			result.DelayedOneTimeFunctions = append(result.DelayedOneTimeFunctions, reactToResult.DelayedOneTimeFunction)
		default:
			panic(fmt.Errorf("failed to append node: uxepected type of ReactToResult %t", reactToResult))
		}
		switch nextNode.Type {
		case NodeTypeSimple:
			nextNodeType = fmt.Sprintf("%T", nextNode.Simple)
		case NodeTypeSubFlow:
			nextNodeType = fmt.Sprintf("%T", nextNode.SubFlow.Intent)
		default:
			nextNodeType = "unknown"
		}
		// Uncomment this line when you need to debug authflow
		// fmt.Println("The next node is", nextNodeType)
		err = appendNode(ctx, deps, findInputReactorResult.Flows, nextNode)
		if err != nil {
			return
		}
		changed = true
	}
}

func appendNode(ctx context.Context, deps *Dependencies, flows Flows, node Node) error {
	flows.Nearest.Nodes = append(flows.Nearest.Nodes, node)

	err := TraverseNode(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			effectGetter, ok := nodeSimple.(EffectGetter)
			if !ok {
				return nil
			}

			effs, err := effectGetter.GetEffects(ctx, deps, flows.Replace(w))
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if runEff, ok := eff.(RunEffect); ok {
					err = runEff.doNotCallThisDirectly(ctx, deps)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		// Intent cannot have run-effect.
		// We do not bother traversing intents here.
	}, flows.Nearest, &node)
	if err != nil {
		return err
	}

	return nil
}

type MilestoneDoUseUser interface {
	Milestone
	MilestoneDoUseUser() string
}

func GetUserID(flows Flows) (userID string, err error) {
	err = TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			if n, ok := nodeSimple.(MilestoneDoUseUser); ok {
				id := n.MilestoneDoUseUser()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if i, ok := intent.(MilestoneDoUseUser); ok {
				id := i.MilestoneDoUseUser()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
	}, flows.Root)

	if userID == "" {
		err = ErrNoUserID
	}

	if err != nil {
		return
	}

	return
}
