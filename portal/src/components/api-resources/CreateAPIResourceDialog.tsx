import React, {
  useCallback,
  useContext,
  useEffect,
  useId,
  useMemo,
} from "react";
import { Dialog, Flex, IconButton, Text, TextField as RadixTextField } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import {
  ErrorParseRule,
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import { useCreateResourceMutationMutation } from "../../graphql/adminapi/mutations/createResourceMutation.generated";
import { useSimpleForm } from "../../hook/useSimpleForm";
import {
  ResourceFormState,
  sanitizeFormState,
} from "./ResourceForm";
import { TextField } from "../v2/TextField/TextField";
import { FormField } from "../v2/FormField/FormField";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import ErrorRenderer from "../../ErrorRenderer";
import styles from "./CreateAPIResourceDialog.module.css";

const defaultState: ResourceFormState = {
  name: "",
  resourceURI: "",
};

const errorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "ResourceDuplicateURI",
    "errors.resources.duplicateURI"
  ),
];

function isFormComplete(state: ResourceFormState): boolean {
  const s = sanitizeFormState(state);
  return Boolean(s.name && s.resourceURI);
}

export interface CreateAPIResourceDialogProps {
  open: boolean;
  onDismiss: () => void;
  onCreated?: (resourceID: string) => void;
}

export const CreateAPIResourceDialog: React.VFC<CreateAPIResourceDialogProps> =
  function CreateAPIResourceDialog({ open, onDismiss, onCreated }) {
    const { renderToString } = useContext(Context);
    const { appID } = useParams<{ appID: string }>();
    const navigate = useNavigate();
    const resourceURIId = useId();
    const [createResource] = useCreateResourceMutationMutation();

    const form = useSimpleForm<ResourceFormState, string>({
      defaultState,
      submit: async (s) => {
        const state = sanitizeFormState(s);
        const result = await createResource({
          variables: {
            input: {
              name: state.name,
              resourceURI: state.resourceURI,
            },
          },
        });
        if (result.data == null) {
          throw new Error("unexpected null data");
        }
        return result.data.createResource.resource.id;
      },
      stateMode:
        "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    });

    const {
      state,
      setState,
      save,
      isUpdating,
      updateError,
      isSubmitted,
      submissionResult,
      reset,
    } = form;

    useEffect(() => {
      if (!open) {
        reset();
      }
    }, [open, reset]);

    useEffect(() => {
      if (!isSubmitted || submissionResult == null || appID == null) {
        return;
      }
      onCreated?.(submissionResult);
      onDismiss();
      navigate(
        `/project/${appID}/api-resources/${encodeURIComponent(submissionResult)}`
      );
    }, [
      isSubmitted,
      submissionResult,
      appID,
      navigate,
      onCreated,
      onDismiss,
    ]);

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

    const onNameChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setState((s) => ({ ...s, name: e.target.value }));
      },
      [setState]
    );

    const onResourceURIChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setState((s) => {
          let resourceURI = e.target.value;
          resourceURI = resourceURI.replace(/^(\s*)https:\/\//, "");
          return { ...s, resourceURI };
        });
      },
      [setState]
    );

    const onSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        if (!isFormComplete(state) || isUpdating) {
          return;
        }
        save().catch(() => {});
      },
      [isUpdating, save, state]
    );

    const formError = useMemo(() => {
      if (updateError == null) {
        return null;
      }
      const apiErrors = parseRawError(updateError);
      const { topErrors } = parseAPIErrors(apiErrors, [], errorRules);
      return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
    }, [updateError]);

    const resourceURILabel = (
      <Flex display="inline-flex" align="center" gap="1">
        <span>
          <FormattedMessage id="ResourceForm.resourceURI.label" />
        </span>
        <span className={styles.requiredMark} aria-hidden="true">
          *
        </span>
        <Tooltip
          side="top"
          content={<FormattedMessage id="ResourceForm.resourceURI.tooltip" />}
        >
          <IconButton
            type="button"
            variant="ghost"
            color="gray"
            size="1"
            aria-label={renderToString("ResourceForm.resourceURI.tooltip")}
          >
            <InfoCircledIcon width="1rem" height="1rem" />
          </IconButton>
        </Tooltip>
      </Flex>
    );

    return (
      <Dialog.Root open={open} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth="480px" size="3">
          <Dialog.Title>
            <FormattedMessage id="CreateAPIResourceScreen.title" />
          </Dialog.Title>
          <form className={styles.form} onSubmit={onSubmit}>
            {formError != null ? (
              <Text as="p" size="2" color="red" className={styles.formError}>
                {formError}
              </Text>
            ) : null}
            <TextField
              size="2"
              required={true}
              label={<FormattedMessage id="ResourceForm.name.label" />}
              hint={<FormattedMessage id="ResourceForm.name.description" />}
              type="text"
              value={state.name}
              onChange={onNameChange}
            />
            <FormField
              size="2"
              label={resourceURILabel}
              htmlFor={resourceURIId}
              hint={
                <FormattedMessage id="ResourceForm.resourceURI.description" />
              }
            >
              <RadixTextField.Root
                id={resourceURIId}
                size="2"
                variant="surface"
                type="text"
                value={state.resourceURI}
                onChange={onResourceURIChange}
              >
                <RadixTextField.Slot side="left">
                  <Text size="2" color="gray">
                    https://
                  </Text>
                </RadixTextField.Slot>
              </RadixTextField.Root>
            </FormField>
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
                text={<FormattedMessage id="create" />}
                loading={isUpdating}
                disabled={!isFormComplete(state) || isUpdating}
              />
            </Flex>
          </form>
        </Dialog.Content>
      </Dialog.Root>
    );
  };
