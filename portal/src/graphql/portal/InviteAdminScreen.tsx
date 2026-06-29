import React, { useCallback, useEffect, useMemo } from "react";
import { FormattedMessage } from "../../intl";
import { useNavigate, useParams } from "react-router-dom";
import { ChevronRightIcon } from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";
import {
  ErrorParseRule,
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import { useCreateCollaboratorInvitationMutation } from "./mutations/createCollaboratorInvitationMutation";
import { useSimpleForm } from "../../hook/useSimpleForm";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import ErrorRenderer from "../../ErrorRenderer";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import styles from "./InviteAdminScreen.module.css";

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

const InviteAdminScreen: React.VFC = function InviteAdminScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
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

  const {
    setState,
    save,
    isUpdating,
    updateError,
    state,
    isSubmitted,
    isDirty,
    reset,
  } = form;

  useEffect(() => {
    if (isSubmitted) {
      navigate("./..");
    }
  }, [isSubmitted, navigate]);

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
    navigate("./..");
  }, [navigate]);

  const onConfirmDiscardNavigation = useCallback(() => {
    reset();
  }, [reset]);

  const formError = useMemo(() => {
    if (updateError == null) return null;
    const apiErrors = parseRawError(updateError);
    const { topErrors } = parseAPIErrors(apiErrors, [], errorRules);
    return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
  }, [updateError]);

  return (
    <ScreenLayoutScrollView>
      <ScreenContent>
        <div className={styles.pageHeader}>
          <div className={styles.breadcrumb}>
            <button
              type="button"
              className={styles.breadcrumbParent}
              onClick={onCancel}
            >
              <Text
                as="span"
                size="5"
                weight="bold"
                color="gray"
                className={styles.breadcrumbText}
              >
                <FormattedMessage id="PortalAdminSettings.title" />
              </Text>
            </button>
            <ChevronRightIcon
              className={styles.breadcrumbSeparator}
              width={20}
              height={20}
            />
            <Text
              as="span"
              size="5"
              weight="bold"
              className={styles.breadcrumbText}
            >
              <FormattedMessage id="InviteAdminScreen.title" />
            </Text>
          </div>
        </div>

        <div className={styles.formContent}>
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
          <div className={styles.formActions}>
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
          </div>
        </div>
      </ScreenContent>
      <NavigationBlockerDialog
        blockNavigation={isDirty}
        onConfirmNavigation={onConfirmDiscardNavigation}
      />
    </ScreenLayoutScrollView>
  );
};

export default InviteAdminScreen;
