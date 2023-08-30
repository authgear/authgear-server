package latte

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func GetAMR(w *workflow.Workflow) []string {
	amrSet := map[string]interface{}{}
	workflows := workflow.FindSubWorkflows[AMRGetter](w)

	authCount := 0
	for _, perWorkflow := range workflows {
		if amrs := perWorkflow.Intent.(AMRGetter).GetAMR(perWorkflow); len(amrs) > 0 {
			authCount++
			for _, value := range amrs {
				amrSet[value] = struct{}{}
			}
		}
	}

	if authCount >= 2 {
		amrSet[model.AMRMFA] = struct{}{}
	}

	amr := make([]string, 0, len(amrSet))
	for k := range amrSet {
		amr = append(amr, k)
	}
	sort.Strings(amr)

	return amr
}

type AMRGetter interface {
	workflow.Intent
	GetAMR(w *workflow.Workflow) []string
}
