import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { Dialog, Flex, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { parseAPIErrors, parseRawError } from "../../error/parse";
import { useUpdateScopeMutationMutation } from "../../graphql/adminapi/mutations/updateScopeMutation.generated";
import { ResourceScopesQueryDocument } from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import {
  AccessPolicy,
  Scope,
} from "../../graphql/adminapi/globalTypes.generated";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { TextField } from "../v2/TextField/TextField";
import { FieldLabelWithTooltip } from "../v2/FieldLabelWithTooltip/FieldLabelWithTooltip";
import { AccessPolicyChips } from "./AccessPolicyChips";
import { FormField } from "../v2/FormField/FormField";
import { Callout } from "../v2/Callout/Callout";
import {
  AccessPolicyState,
  CLOSED_ACCESS_POLICY,
  VISIBLE_ACCESS_POLICY_KEYS,
  accessPolicyLabelIDs,
  accessPolicyStateFromAccessPolicy,
} from "./accessPolicy";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import ErrorRenderer from "../../ErrorRenderer";
import styles from "./EditScopeDialog.module.css";

interface EditScopeFormState {
  description: string;
  accessPolicy: AccessPolicyState;
}

export interface EditScopeDialogProps {
  resourceURI: string;
  // The parent resource's policy: a category allowed here but not there is
  // unreachable, and is warned about.
  resourceAccessPolicy: AccessPolicy;
  scope: Scope | null;
  onDismiss: () => void;
  onSaved?: () => void;
}

export const EditScopeDialog: React.VFC<EditScopeDialogProps> =
  function EditScopeDialog({
    resourceURI,
    resourceAccessPolicy,
    scope,
    onDismiss,
    onSaved,
  }) {
    const { renderToString } = useContext(Context);
    const [updateScope] = useUpdateScopeMutationMutation();
    const open = scope != null;
    const scopeName = scope?.scope ?? "";

    const form = useSimpleForm<EditScopeFormState, Scope>({
      defaultState: {
        description: "",
        accessPolicy: CLOSED_ACCESS_POLICY,
      },
      submit: async (state) => {
        if (scope == null) {
          throw new Error("unexpected null scope");
        }
        const result = await updateScope({
          variables: {
            input: {
              resourceURI,
              scope: scope.scope,
              description: state.description.trim(),
              accessPolicy: state.accessPolicy,
            },
          },
          refetchQueries: [ResourceScopesQueryDocument],
          awaitRefetchQueries: true,
        });
        if (result.data == null) {
          throw new Error("unexpected null data");
        }
        return result.data.updateScope.scope;
      },
    });

    const { state, setState, save, isUpdating, updateError, reset } = form;

    useEffect(() => {
      if (scope == null) {
        reset();
        return;
      }
      setState(() => ({
        description: scope.description ?? "",
        accessPolicy: accessPolicyStateFromAccessPolicy(scope.accessPolicy),
      }));
    }, [scope, reset, setState]);

    const onCancel = useCallback(() => {
      if (!isUpdating) {
        onDismiss();
      }
    }, [isUpdating, onDismiss]);

    const onOpenChange = useCallback(
      (nextOpen: boolean) => {
        if (!nextOpen) {
          onCancel();
        }
      },
      [onCancel]
    );

    const onDescriptionChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const description = e.target.value;
        setState((s) => ({ ...s, description }));
      },
      [setState]
    );

    const onAccessPolicyChange = useCallback(
      (accessPolicy: AccessPolicyState) => {
        setState((s) => ({ ...s, accessPolicy }));
      },
      [setState]
    );

    const onSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        if (isUpdating || scope == null) {
          return;
        }
        save()
          .then(() => {
            onSaved?.();
            onDismiss();
          })
          .catch(() => {});
      },
      [isUpdating, save, scope, onSaved, onDismiss]
    );

    const formError = useMemo(() => {
      if (updateError == null) {
        return null;
      }
      const apiErrors = parseRawError(updateError);
      const { topErrors } = parseAPIErrors(apiErrors, [], []);
      return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
    }, [updateError]);

    const unreachableCategories = useMemo(
      () =>
        VISIBLE_ACCESS_POLICY_KEYS.filter(
          (key) => state.accessPolicy[key] && !resourceAccessPolicy[key]
        ).map((key) => renderToString(accessPolicyLabelIDs[key])),
      [state.accessPolicy, resourceAccessPolicy, renderToString]
    );

    const descriptionLabel = (
      <FieldLabelWithTooltip
        tooltip={<FormattedMessage id="ScopeForm.description.tooltip" />}
        tooltipLabel={renderToString("ScopeForm.description.tooltip")}
      >
        <FormattedMessage id="ScopeForm.description.label" />
      </FieldLabelWithTooltip>
    );

    return (
      <Dialog.Root open={open} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth="480px" size="3">
          <Dialog.Title>
            <FormattedMessage id="EditScopeScreen.title" />
          </Dialog.Title>
          <form className={styles.form} onSubmit={onSubmit}>
            {formError != null ? (
              <Text as="p" size="2" color="red" className={styles.formError}>
                {formError}
              </Text>
            ) : null}
            <TextField
              size="2"
              label={<FormattedMessage id="ScopeForm.scope.label" />}
              type="text"
              value={scopeName}
              readOnly={true}
            />
            <TextField
              size="2"
              label={descriptionLabel}
              type="text"
              value={state.description}
              onChange={onDescriptionChange}
              placeholder={renderToString(
                "CreateScopeForm.description.placeholder"
              )}
            />
            <FormField
              size="2"
              labelSpace="1"
              label={
                <FieldLabelWithTooltip
                  tooltip={
                    <FormattedMessage id="AccessPolicyCheckboxes.scope.hint" />
                  }
                  tooltipLabel={renderToString(
                    "AccessPolicyCheckboxes.scope.hint"
                  )}
                >
                  <FormattedMessage id="AccessPolicyCheckboxes.scope.title" />
                </FieldLabelWithTooltip>
              }
            >
              <AccessPolicyChips
                value={state.accessPolicy}
                disabled={isUpdating}
                onChange={onAccessPolicyChange}
              />
            </FormField>
            {unreachableCategories.length > 0 ? (
              <Callout
                type="warning"
                size="1"
                showCloseButton={false}
                text={
                  <FormattedMessage
                    id="ScopeForm.access-policy.unreachable"
                    values={{ categories: unreachableCategories.join(", ") }}
                  />
                }
              />
            ) : null}
            <Flex gap="3" mt="4" justify="end">
              <SecondaryButton
                size="2"
                text={<FormattedMessage id="cancel" />}
                onClick={onCancel}
                disabled={isUpdating}
              />
              <PrimaryButton
                type="submit"
                size="2"
                text={<FormattedMessage id="save" />}
                loading={isUpdating}
                disabled={isUpdating}
              />
            </Flex>
          </form>
        </Dialog.Content>
      </Dialog.Root>
    );
  };
