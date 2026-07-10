import React, { useCallback, useEffect, useMemo } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";
import {
  ErrorParseRule,
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import { useCreateCollaboratorInvitationMutation } from "./mutations/createCollaboratorInvitationMutation";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import ErrorRenderer from "../../ErrorRenderer";

interface FormState {
  email: string;
}

const defaultState: FormState = { email: "" };

const errorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "CollaboratorInvitationDuplicate",
    "InviteAdminScreen.duplicated-error"
  ),
];

export interface InviteAdminDialogProps {
  open: boolean;
  onDismiss: () => void;
}

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
      stateMode:
        "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
      defaultState,
      submit,
    });

    const {
      setState,
      save,
      isUpdating,
      updateError,
      state,
      isSubmitted,
      reset,
    } = form;

    useEffect(() => {
      if (isSubmitted) {
        onDismiss();
      }
    }, [isSubmitted, onDismiss]);

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

    const onSubmit = useCallback(() => {
      save().catch(() => {});
    }, [save]);

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

    const formError = useMemo(() => {
      if (updateError == null) {
        return null;
      }
      const apiErrors = parseRawError(updateError);
      const { topErrors } = parseAPIErrors(apiErrors, [], errorRules);
      return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
    }, [updateError]);

    return (
      <Dialog.Root open={open} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="InviteAdminScreen.title" />
          </Dialog.Title>
          <TextField
            size="2"
            type="email"
            required={true}
            label={<FormattedMessage id="InviteAdminScreen.email.label" />}
            hint={
              formError == null ? (
                <FormattedMessage id="InviteAdminScreen.email.description" />
              ) : undefined
            }
            error={formError}
            value={state.email}
            onChange={onEmailChange}
          />
          <Flex gap="3" mt="4" justify="end">
            <SecondaryButton
              size="2"
              text={<FormattedMessage id="cancel" />}
              onClick={onCancel}
              disabled={isUpdating}
            />
            <PrimaryButton
              size="2"
              text={<FormattedMessage id="InviteAdminScreen.add-user.label" />}
              onClick={onSubmit}
              loading={isUpdating}
              disabled={isUpdating}
            />
          </Flex>
        </Dialog.Content>
      </Dialog.Root>
    );
  };

export default InviteAdminDialog;
