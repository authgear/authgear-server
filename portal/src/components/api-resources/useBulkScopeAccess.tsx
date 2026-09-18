import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "../../intl";
import {
  AccessPolicyInput,
  Scope,
} from "../../graphql/adminapi/globalTypes.generated";
import { useUpdateScopeMutationMutation } from "../../graphql/adminapi/mutations/updateScopeMutation.generated";
import { parseRawError } from "../../error/parse";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useCalloutToast } from "../v2/Callout/Callout";
import {
  VISIBLE_ACCESS_POLICY_KEYS,
  VisibleAccessPolicyKey,
} from "./accessPolicy";

const TOAST_DURATION_MS = 2000;

// The categories to set on every selected scope; a missing key is left as
// each scope has it.
export type BulkAccessChanges = Partial<
  Record<VisibleAccessPolicyKey, boolean>
>;

export interface BulkScopeAccess {
  // Selected scopes by id. The Scope is kept so scopes on other pages can
  // still be updated; entries on the current page are refreshed from the
  // latest query result so an edit made after ticking is not applied from
  // stale data.
  selected: ReadonlyMap<string, Scope>;
  toggle: (scope: Scope, checked: boolean) => void;
  toggleAll: (scopes: readonly Scope[], checked: boolean) => void;
  clear: () => void;
  apply: (changes: BulkAccessChanges) => Promise<void>;
  isApplying: boolean;
}

function inputFor(
  scope: Scope,
  changes: BulkAccessChanges
): AccessPolicyInput | null {
  const input: AccessPolicyInput = {};
  let differs = false;
  for (const key of VISIBLE_ACCESS_POLICY_KEYS) {
    const next = changes[key];
    if (next != null && scope.accessPolicy[key] !== next) {
      input[key] = next;
      differs = true;
    }
  }
  return differs ? input : null;
}

// One updateScope per selected scope that actually changes, sequentially,
// sending only the keys that change. Not atomic: a failure part-way leaves
// the earlier scopes updated, so the failed ones stay selected for a retry
// and the toast says how many.
export function useBulkScopeAccess({
  resourceURI,
  scopes,
  refetch,
}: {
  resourceURI: string;
  // The scopes currently loaded (this page of the list).
  scopes: readonly Scope[];
  refetch: () => Promise<unknown>;
}): BulkScopeAccess {
  const { setErrors } = useErrorMessageBarContext();
  const { showToast } = useCalloutToast();
  const [updateScope] = useUpdateScopeMutationMutation();
  const [stored, setSelected] = useState<ReadonlyMap<string, Scope>>(
    () => new Map()
  );
  const selected = useMemo(() => {
    const fresh = new Map(stored);
    for (const scope of scopes) {
      if (fresh.has(scope.id)) {
        fresh.set(scope.id, scope);
      }
    }
    return fresh;
  }, [stored, scopes]);
  const [isApplying, setIsApplying] = useState(false);

  const toggle = useCallback((scope: Scope, checked: boolean) => {
    setSelected((prev) => {
      const next = new Map(prev);
      if (checked) {
        next.set(scope.id, scope);
      } else {
        next.delete(scope.id);
      }
      return next;
    });
  }, []);

  const toggleAll = useCallback(
    (scopes: readonly Scope[], checked: boolean) => {
      setSelected((prev) => {
        const next = new Map(prev);
        for (const scope of scopes) {
          if (checked) {
            next.set(scope.id, scope);
          } else {
            next.delete(scope.id);
          }
        }
        return next;
      });
    },
    []
  );

  const clear = useCallback(() => {
    setSelected(new Map());
  }, []);

  const apply = useCallback(
    async (changes: BulkAccessChanges) => {
      const targets = Array.from(selected.values());
      setIsApplying(true);
      const failed = new Map<string, Scope>();
      let firstError: unknown = null;
      try {
        for (const scope of targets) {
          const accessPolicy = inputFor(scope, changes);
          if (accessPolicy == null) {
            continue;
          }
          try {
            await updateScope({
              variables: {
                input: { resourceURI, scope: scope.scope, accessPolicy },
              },
            });
          } catch (e: unknown) {
            failed.set(scope.id, scope);
            firstError ??= e;
          }
        }
        await refetch();
      } finally {
        setIsApplying(false);
      }
      const total = targets.length;
      if (failed.size === 0) {
        setSelected(new Map());
        showToast({
          type: "success",
          text: (
            <FormattedMessage
              id="ScopeBulkAccess.toast.done"
              values={{ count: total }}
            />
          ),
          duration: TOAST_DURATION_MS,
        });
        return;
      }
      setSelected(failed);
      setErrors(parseRawError(firstError));
      showToast({
        type: "warning",
        text: (
          <FormattedMessage
            id="ScopeBulkAccess.toast.partial"
            values={{ done: total - failed.size, total }}
          />
        ),
        duration: TOAST_DURATION_MS,
      });
    },
    [selected, updateScope, resourceURI, refetch, showToast, setErrors]
  );

  return useMemo(
    () => ({ selected, toggle, toggleAll, clear, apply, isApplying }),
    [selected, toggle, toggleAll, clear, apply, isApplying]
  );
}
