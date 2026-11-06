// Package kubefake has test helpers for the fake Kubernetes clients.
package kubefake

import (
	"encoding/json"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	coretesting "k8s.io/client-go/testing"
)

// AddServerSideApplyReactor makes the fake dynamic client handle server-side apply, which its tracker
// cannot do for unstructured objects nor for objects that do not exist yet.
// The applied object is merged into the existing one: maps are merged and the other values are replaced,
// like a server-side apply that takes over every field it sets and leaves the fields of other managers.
// The fake client drops ApplyOptions, so a dry run is written too.
func AddServerSideApplyReactor(client *fakedynamic.FakeDynamicClient) {
	tracker := client.Tracker()
	client.PrependReactor("patch", "*", func(action coretesting.Action) (bool, runtime.Object, error) {
		patchAction, ok := action.(coretesting.PatchActionImpl)
		if !ok || patchAction.GetPatchType() != types.ApplyPatchType {
			return false, nil, nil
		}

		gvr := patchAction.GetResource()
		ns := patchAction.GetNamespace()

		applied := &unstructured.Unstructured{}
		if err := json.Unmarshal(patchAction.GetPatch(), &applied.Object); err != nil {
			return true, nil, err
		}

		existing, err := tracker.Get(gvr, ns, patchAction.GetName(), metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			if err := tracker.Create(gvr, applied, ns); err != nil {
				return true, nil, err
			}
			return true, applied, nil
		} else if err != nil {
			return true, nil, err
		}

		existingMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(existing)
		if err != nil {
			return true, nil, err
		}
		merged := &unstructured.Unstructured{Object: mergeMaps(existingMap, applied.Object)}
		if err := tracker.Update(gvr, merged, ns); err != nil {
			return true, nil, err
		}
		return true, merged, nil
	})
}

func mergeMaps(dst map[string]any, src map[string]any) map[string]any {
	out := runtime.DeepCopyJSON(dst)
	for k, v := range src {
		srcMap, srcIsMap := v.(map[string]any)
		dstMap, dstIsMap := out[k].(map[string]any)
		if srcIsMap && dstIsMap {
			out[k] = mergeMaps(dstMap, srcMap)
		} else {
			out[k] = runtime.DeepCopyJSONValue(v)
		}
	}
	return out
}
