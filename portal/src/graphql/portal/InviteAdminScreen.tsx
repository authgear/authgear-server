import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";
import FormTextField from "../../FormTextField";
import { GenericErrorHandlingRule } from "../../error/useGenericError";
import { useCreateCollaboratorInvitationMutation } from "./mutations/createCollaboratorInvitationMutation";

import styles from "./InviteAdminScreen.module.scss";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

interface FormState {
  email: string;
}

const defaultFormState: FormState = {
  email: "",
};

interface InviteAdminContentProps {
  form: SimpleFormModel<FormState>;
}

const InviteAdminContent: React.FC<InviteAdminContentProps> = function InviteAdminContent(
  props: InviteAdminContentProps
) {
  const { state, setState } = props.form;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "..", label: <FormattedMessage id="PortalAdminSettings.title" /> },
      { to: ".", label: <FormattedMessage id="InviteAdminScreen.title" /> },
    ];
  }, []);

  const onEmailChange = useCallback(
    (_, value?: string) => {
      setState((s) => ({ ...s, email: value ?? "" }));
    },
    [setState]
  );

  return (
    <div className={styles.root}>
      <NavBreadcrumb className={styles.breadcrumb} items={navBreadcrumbItems} />
      <FormTextField
        jsonPointer="/inviteeEmail"
        parentJSONPointer=""
        fieldName="inviteeEmail"
        className={styles.emailField}
        type="text"
        label={renderToString("InviteAdminScreen.email.label")}
        value={state.email}
        onChange={onEmailChange}
      />
    </div>
  );
};

const InviteAdminScreen: React.FC = function InviteAdminScreen() {
  const { appID } = useParams();
  const navigate = useNavigate();
  const {
    createCollaboratorInvitation,
  } = useCreateCollaboratorInvitationMutation(appID);

  const submit = useCallback(
    async (state: FormState) => {
      await createCollaboratorInvitation(state.email);
    },
    [createCollaboratorInvitation]
  );

  const form = useSimpleForm(defaultFormState, submit);

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("..");
    }
  }, [form.isSubmitted, navigate]);

  const errorRules: GenericErrorHandlingRule[] = useMemo(
    () => [
      {
        reason: "CollaboratorInvitationDuplicate",
        errorMessageID: "InviteAdminScreen.duplicated-error",
      },
    ],
    []
  );

  const saveButtonProps = useMemo(
    () => ({
      labelId: "InviteAdminScreen.add-user.label",
      iconName: "Add",
    }),
    []
  );

  return (
    <FormContainer
      form={form}
      saveButtonProps={saveButtonProps}
      errorParseRules={errorRules}
    >
      <InviteAdminContent form={form} />
    </FormContainer>
  );
};

export default InviteAdminScreen;
