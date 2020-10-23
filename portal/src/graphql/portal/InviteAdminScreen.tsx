import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";

import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import FormTextField from "../../FormTextField";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import { useValidationError } from "../../error/useValidationError";
import { useGenericError } from "../../error/useGenericError";
import { FormContext } from "../../error/FormContext";
import { useCreateCollaboratorInvitationMutation } from "./mutations/createCollaboratorInvitationMutation";

import styles from "./InviteAdminScreen.module.scss";

const InviteAdminContent: React.FC = function InviteAdminContent() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const navigate = useNavigate();

  const {
    createCollaboratorInvitation,
    loading: creatingCollaboratorInvitation,
    error: createCollaboratorInvitationError,
  } = useCreateCollaboratorInvitationMutation(appID);

  const {
    unhandledCauses: rawUnhandledCauses,
    otherError,
    value: formContextValue,
  } = useValidationError(createCollaboratorInvitationError);

  const { errorMessage: otherErrorMessage, unhandledCauses } = useGenericError(
    otherError,
    rawUnhandledCauses,
    [
      {
        reason: "CollaboratorInvitationDuplicate",
        errorMessageID: "InviteAdminScreen.duplicated-error",
      },
    ]
  );

  const [email, setEmail] = useState("");
  const [submittedForm, setSubmittedForm] = useState(false);

  const isFormModified = useMemo(() => {
    return email !== "";
  }, [email]);

  const resetForm = useCallback(() => {
    setEmail("");
  }, []);

  const onEmailChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setEmail(value);
  }, []);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      createCollaboratorInvitation(email)
        .then((invitationID) => {
          if (invitationID !== null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [createCollaboratorInvitation, email]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("../");
    }
  }, [submittedForm, navigate]);

  return (
    <FormContext.Provider value={formContextValue}>
      <form className={styles.content} onSubmit={onFormSubmit}>
        <ModifiedIndicatorPortal
          resetForm={resetForm}
          isModified={isFormModified}
        />
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        <FormTextField
          jsonPointer="/inviteeEmail"
          parentJSONPointer=""
          fieldName="inviteeEmail"
          className={styles.emailField}
          type="text"
          label={renderToString("InviteAdminScreen.email.label")}
          errorMessage={otherErrorMessage}
          value={email}
          onChange={onEmailChange}
        />
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified}
          labelId="InviteAdminScreen.add-user.label"
          loading={creatingCollaboratorInvitation}
        />
        <NavigationBlockerDialog
          blockNavigation={!submittedForm && isFormModified}
        />
      </form>
    </FormContext.Provider>
  );
};

const InviteAdminScreen: React.FC = function InviteAdminScreen() {
  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "../", label: <FormattedMessage id="SettingsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="InviteAdminScreen.title" /> },
    ];
  }, []);

  return (
    <main className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.wrapper}>
        <NavBreadcrumb
          className={styles.breadcrumb}
          items={navBreadcrumbItems}
        />
        <InviteAdminContent />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default InviteAdminScreen;
