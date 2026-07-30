import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
} from "react";
import { Dialog, Flex, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import {
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import { useUpdateScopeMutationMutation } from "../../graphql/adminapi/mutations/updateScopeMutation.generated";
import { ResourceScopesQueryDocument } from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { TextField } from "../v2/TextField/TextField";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import ErrorRenderer from "../../ErrorRenderer";
import styles from "./EditScopeDialog.module.css";

interface EditScopeFormState {
  description: string;
}

export interface EditScopeDialogProps {
  resourceURI: string;
  scope: Scope | null;
  onDismiss: () => void;
  onSaved?: () => void;
}

export const EditScopeDialog: React.VFC<EditScopeDialogProps> =
  function EditScopeDialog({ resourceURI, scope, onDismiss, onSaved }) {
    const { renderToString } = useContext(Context);
    const [updateScope] = useUpdateScopeMutationMutation();
    const open = scope != null;
    const scopeName = scope?.scope ?? "";

    const form = useSimpleForm<EditScopeFormState, Scope>({
      defaultState: { description: "" },
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
            },
          },
          refetchQueries: [ResourceScopesQueryDocument],
          awaitRefetchQueries: true,
        });
        if (result.data == null) {
          throw new Error("unexpected null data");
        }
        return result.data.updateScope.scope;
      }
    });

    const {
      state,
      setState,
      save,
      isUpdating,
      updateError,
      reset,
    } = form;

    useEffect(() => {
      if (scope == null) {
        reset();
        return;
      }
      setState(() => ({
        description: scope.description ?? "",
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
        setState((s) => ({ ...s, description: e.target.value }));
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
              label={<FormattedMessage id="ScopeForm.description.label" />}
              type="text"
              value={state.description}
              onChange={onDescriptionChange}
              placeholder={renderToString(
                "CreateScopeForm.description.placeholder"
              )}
            />
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
