import React, { useCallback, useEffect, useMemo } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import { useCreateCollaboratorInvitationMutation } from "./mutations/createCollaboratorInvitationMutation";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormProvider } from "../../form";
import { useErrorMessage } from "../../formbinding";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";

interface FormState {
  email: string;
}

const defaultState: FormState = { email: "" };

const emailFieldErrorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "CollaboratorInvitationDuplicate",
    "InviteAdminScreen.duplicated-error"
  ),
];

export interface InviteAdminDialogProps {
  open: boolean;
  onDismiss: () => void;
}

const InviteAdminEmailField: React.VFC<{
  value: string;
  onChange: React.ChangeEventHandler<HTMLInputElement>;
}> = function InviteAdminEmailField({ value, onChange }) {
  // Register the field so ValidationFailed causes for /inviteeEmail and the
  // CollaboratorInvitationDuplicate rule surface as an input error message.
  const field = useMemo(
    () => ({
      parentJSONPointer: "",
      fieldName: "inviteeEmail",
      rules: emailFieldErrorRules,
    }),
    []
  );
  const { errorMessage } = useErrorMessage(field);

  return (
    <TextField
      size="2"
      type="email"
      required={true}
      label={<FormattedMessage id="InviteAdminScreen.email.label" />}
      hint={
        errorMessage == null ? (
          <FormattedMessage id="InviteAdminScreen.email.description" />
        ) : undefined
      }
      error={errorMessage}
      value={value}
      onChange={onChange}
    />
  );
};

const InviteAdminDialog: React.VFC<InviteAdminDialogProps> =
  function InviteAdminDialog({ open, onDismiss }) {
    const { appID } = useParams() as { appID: string };
    const { createCollaboratorInvitation } =
      useCreateCollaboratorInvitationMutation(appID);

    const submit = useCallback(
      async (state: FormState) => {
        await createCollaboratorInvitation(state.email);
      },
      [createCollaboratorInvitation]
    );

    const form = useSimpleForm({
      defaultState,
      submit,
    });

    const { setState, save, isUpdating, updateError, state, reset } = form;

    useEffect(() => {
      if (!open) {
        reset();
      }
    }, [open, reset]);

    const onEmailChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((s) => ({ ...s, email: value }));
      },
      [setState]
    );

    const onSubmit = useCallback(
      async (e: React.FormEvent) => {
        e.preventDefault();
        if (isUpdating) {
          return;
        }
        try {
          await save();
          onDismiss();
        } catch {
          // Error is rendered via FormProvider / field error message.
        }
      },
      [isUpdating, onDismiss, save]
    );

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

    return (
      <FormProvider loading={isUpdating} error={updateError}>
        <Dialog.Root open={open} onOpenChange={onOpenChange}>
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="InviteAdminScreen.title" />
            </Dialog.Title>
            {/* Skip browser-native validation so empty / invalid email is
                reported as a FormField error message from the API instead. */}
            <form noValidate={true} onSubmit={onSubmit}>
              <InviteAdminEmailField
                value={state.email}
                onChange={onEmailChange}
              />
              <Flex gap="3" mt="4" justify="end">
                <SecondaryButton
                  size="2"
                  type="button"
                  text={<FormattedMessage id="cancel" />}
                  onClick={onCancel}
                  disabled={isUpdating}
                />
                <PrimaryButton
                  size="2"
                  type="submit"
                  text={
                    <FormattedMessage id="InviteAdminScreen.add-user.label" />
                  }
                  loading={isUpdating}
                  disabled={isUpdating}
                />
              </Flex>
            </form>
          </Dialog.Content>
        </Dialog.Root>
      </FormProvider>
    );
  };

export default InviteAdminDialog;
