import React, { useCallback, useMemo, useState } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { Callout } from "../v2/Callout/Callout";
import { AccessPolicyChips } from "./AccessPolicyChips";
import {
  AccessPolicyState,
  CLOSED_ACCESS_POLICY,
  VISIBLE_ACCESS_POLICY_KEYS,
  VisibleAccessPolicyKey,
  useAccessPolicyCategoryList,
} from "./accessPolicy";
import { BulkAccessChanges } from "./useBulkScopeAccess";
import styles from "./ScopeBulkAccessDialog.module.css";

type ChipState = "on" | "off" | "mixed";
type ChipStates = Record<VisibleAccessPolicyKey, ChipState>;

function aggregate(scopes: readonly Scope[]): ChipStates {
  const out = {} as ChipStates;
  for (const key of VISIBLE_ACCESS_POLICY_KEYS) {
    const allowed = scopes.filter((scope) => scope.accessPolicy[key]).length;
    out[key] =
      allowed === 0 ? "off" : allowed === scopes.length ? "on" : "mixed";
  }
  return out;
}

export interface ScopeBulkAccessDialogProps {
  open: boolean;
  scopes: readonly Scope[];
  isApplying: boolean;
  onApply: (changes: BulkAccessChanges) => void;
  onDismiss: () => void;
}

// Chips start from the selected scopes' aggregate: on when every scope
// allows the category, off when none does, mixed (dashed) when they differ.
// A mixed chip left alone keeps each scope as it is; once clicked it is a
// definite on or off for all of them.
export const ScopeBulkAccessDialog: React.VFC<ScopeBulkAccessDialogProps> =
  function ScopeBulkAccessDialog({
    open,
    scopes,
    isApplying,
    onApply,
    onDismiss,
  }) {
    const formatCategories = useAccessPolicyCategoryList();
    const initial = useMemo(() => aggregate(scopes), [scopes]);
    const [states, setStates] = useState<ChipStates>(initial);

    // Start from the aggregate again whenever the dialog opens or the
    // selection changes, without an effect (see the React docs on adjusting
    // state when a prop changes).
    const [prevOpen, setPrevOpen] = useState(open);
    const [prevInitial, setPrevInitial] = useState(initial);
    if (prevOpen !== open || prevInitial !== initial) {
      setPrevOpen(open);
      setPrevInitial(initial);
      setStates(initial);
    }

    const chipValue = useMemo<AccessPolicyState>(() => {
      const value = { ...CLOSED_ACCESS_POLICY };
      for (const key of VISIBLE_ACCESS_POLICY_KEYS) {
        value[key] = states[key] === "on";
      }
      return value;
    }, [states]);
    const mixedKeys = useMemo(
      () =>
        new Set(
          VISIBLE_ACCESS_POLICY_KEYS.filter((key) => states[key] === "mixed")
        ),
      [states]
    );

    const onChipsChange = useCallback((next: AccessPolicyState) => {
      setStates((prev) => {
        const out = { ...prev };
        for (const key of VISIBLE_ACCESS_POLICY_KEYS) {
          if (prev[key] === "mixed" && !next[key]) {
            continue; // untouched
          }
          out[key] = next[key] ? "on" : "off";
        }
        return out;
      });
    }, []);

    const changes = useMemo<BulkAccessChanges>(() => {
      const out: BulkAccessChanges = {};
      for (const key of VISIBLE_ACCESS_POLICY_KEYS) {
        if (states[key] !== "mixed" && states[key] !== initial[key]) {
          out[key] = states[key] === "on";
        }
      }
      return out;
    }, [states, initial]);
    const hasChanges = Object.keys(changes).length > 0;

    const turningOff = useMemo(
      () => VISIBLE_ACCESS_POLICY_KEYS.filter((key) => changes[key] === false),
      [changes]
    );

    const onOpenChange = useCallback(
      (nextOpen: boolean) => {
        if (!nextOpen && !isApplying) {
          onDismiss();
        }
      },
      [isApplying, onDismiss]
    );

    const onSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        if (!hasChanges || isApplying) {
          return;
        }
        onApply(changes);
      },
      [hasChanges, isApplying, onApply, changes]
    );

    return (
      <Dialog.Root open={open} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth="520px" size="3">
          <Dialog.Title>
            <FormattedMessage
              id="ScopeBulkAccessDialog.title"
              values={{ count: scopes.length }}
            />
          </Dialog.Title>
          <Dialog.Description size="2" color="gray">
            <FormattedMessage id="ScopeBulkAccessDialog.description" />
          </Dialog.Description>
          <form className={styles.form} onSubmit={onSubmit}>
            <AccessPolicyChips
              value={chipValue}
              mixedKeys={mixedKeys}
              disabled={isApplying}
              onChange={onChipsChange}
            />
            {turningOff.length > 0 ? (
              <Callout
                type="warning"
                size="1"
                showCloseButton={false}
                text={
                  <FormattedMessage
                    id="ScopeBulkAccessDialog.turning-off"
                    values={{ categories: formatCategories(turningOff) }}
                  />
                }
              />
            ) : null}
            <Flex gap="3" mt="4" justify="end">
              <SecondaryButton
                size="2"
                text={<FormattedMessage id="cancel" />}
                onClick={onDismiss}
                disabled={isApplying}
              />
              <PrimaryButton
                type="submit"
                size="2"
                text={
                  <FormattedMessage
                    id="ScopeBulkAccessDialog.apply"
                    values={{ count: scopes.length }}
                  />
                }
                loading={isApplying}
                disabled={!hasChanges || isApplying}
              />
            </Flex>
          </form>
        </Dialog.Content>
      </Dialog.Root>
    );
  };
