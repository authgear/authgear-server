package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/feature/botprotection"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type AcceptResult struct {
	BotProtectionVerificationResult *BotProtectionVerificationResult `json:"bot_protection,omitempty"`
}

// Accept executes the flow to the deepest using input.
// In addition to the errors caused by intents and nodes,
// ErrEOF and ErrNoChange can be returned.
func Accept(ctx context.Context, deps *Dependencies, flows Flows, rawMessage json.RawMessage) (*AcceptResult, error) {
	return accept(ctx, deps, flows, func(inputSchema InputSchema) (Input, error) {
		if rawMessage != nil && inputSchema != nil {
			input, err := inputSchema.MakeInput(rawMessage)
			if err != nil {
				return nil, err
			}
			return input, nil
		}
		return nil, nil
	})
}

func AcceptSyntheticInput(ctx context.Context, deps *Dependencies, flows Flows, syntheticInput Input) (result *AcceptResult, err error) {
	return accept(ctx, deps, flows, func(inputSchema InputSchema) (Input, error) {
		return syntheticInput, nil
	})
}

// nolint: gocognit
func accept(ctx context.Context, deps *Dependencies, flows Flows, inputFn func(inputSchema InputSchema) (Input, error)) (result *AcceptResult, err error) {
	var changed bool
	defer func() {
		if changed {
			flows.Nearest.StateToken = newStateToken()
		}
		if !changed && err == nil {
			err = ErrNoChange
		}
	}()

	for {
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

		var nextNode *Node
		nextNode, err = findInputReactorResult.InputReactor.ReactTo(ctx, deps, findInputReactorResult.Flows, input)

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
		if errors.Is(err, ErrUpdateNode) {
			err = nil
			// We still consider the flow has something changes.
			changed = true

			nodeToReplace := nextNode

			// precondition: ErrUpdateNode requires at least one node.
			if len(findInputReactorResult.Flows.Nearest.Nodes) == 0 {
				panic(fmt.Errorf("input reactor %T returned ErrUpdateNode, but there are no nodes", findInputReactorResult.InputReactor))
			}

			// Update the node inplace.
			findInputReactorResult.Flows.Nearest.Nodes[len(findInputReactorResult.Flows.Nearest.Nodes)-1] = *nodeToReplace

			// We have to stop and return here because this edge will react to this input indefinitely.
			return
		}

		// Handle err == ErrBotProtectionVerificationFailed
		if errors.Is(err, ErrBotProtectionVerificationFailed) {
			err = nil
			// We still consider the flow has something changes.
			changed = true

			return &AcceptResult{
				BotProtectionVerificationResult: &BotProtectionVerificationResult{
					Outcome: BotProtectionVerificationOutcomeFailed,
				}}, botprotection.ErrVerificationFailed
		}

		// Handle err == ErrBotProtectionVerificationServiceUnavailable
		if errors.Is(err, ErrBotProtectionVerificationServiceUnavailable) {
			err = nil
			// We still consider the flow has something changes.
			changed = true
			return &AcceptResult{
				BotProtectionVerificationResult: &BotProtectionVerificationResult{
					Outcome: BotProtectionVerificationOutcomeFailed,
				}}, botprotection.ErrVerificationServiceUnavailable
		}

		if errors.Is(err, ErrBotProtectionVerificationSuccess) {
			err = nil
			// We still consider the flow has something changes.
			changed = true

			// FIXME: Below line will cause cyclic dependency
			// inputBPV, ok := input.(*declarative.InputBotProtectionVerification)
			// if !ok {
			// 	return nil, ErrIncompatibleInput
			// }
			return &AcceptResult{
				BotProtectionVerificationResult: &BotProtectionVerificationResult{
					Outcome: BotProtectionVerificationOutcomeVerified,
					// FIXME: Figure out how to store response here
					// Response: inputBPV.GetBotProtectionProviderResponse(),
				}}, nil
		}

		// Handle other error.
		if err != nil {
			err = errorutil.WithDetails(err, errorutil.Details{
				"FlowType": apierrors.APIErrorDetail.Value(flows.Nearest.Intent.(PublicFlow).FlowType()),
			})

			return
		}

		// We need to append the nextNode to the closest flow.
		err = appendNode(ctx, deps, findInputReactorResult.Flows, *nextNode)
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
