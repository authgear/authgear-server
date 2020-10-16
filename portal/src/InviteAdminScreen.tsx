import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";

import NavBreadcrumb, { BreadcrumbItem } from "./NavBreadcrumb";
import ButtonWithLoading from "./ButtonWithLoading";
import NavigationBlockerDialog from "./NavigationBlockerDialog";
import FormTextField from "./FormTextField";
import ShowUnhandledValidationErrorCause from "./error/ShowUnhandledValidationErrorCauses";
import { useValidationError } from "./error/useValidationError";
import { useGenericError } from "./error/useGenericError";
import { FormContext } from "./error/FormContext";
import { useCreateCollaboratorInvitationMutation } from "./graphql/portal/mutations/createCollaboratorInvitationMutation";

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
    unhandledCauses,
    otherError,
    value: formContextValue,
  } = useValidationError(createCollaboratorInvitationError);

  const otherErrorMessage = useGenericError(otherError, [
    {
      reason: "CollaboratorInvitationDuplicate",
      errorMessageID: "InviteAdminScreen.duplicated-error",
    },
  ]);

  const [email, setEmail] = useState("");
  const [submittedForm, setSubmittedForm] = useState(false);

  const isFormModified = useMemo(() => {
    return email !== "";
  }, [email]);

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
      <NavBreadcrumb className={styles.breadcrumb} items={navBreadcrumbItems} />
      <InviteAdminContent />
    </main>
  );
};

export default InviteAdminScreen;
