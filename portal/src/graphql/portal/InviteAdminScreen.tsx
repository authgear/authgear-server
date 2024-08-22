import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";
import FormTextField from "../../FormTextField";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import { useCreateCollaboratorInvitationMutation } from "./mutations/createCollaboratorInvitationMutation";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import styles from "./InviteAdminScreen.module.css";

interface FormState {
  email: string;
}

const defaultState: FormState = {
  email: "",
};

interface InviteAdminContentProps {
  form: SimpleFormModel<FormState>;
}

const InviteAdminContent: React.VFC<InviteAdminContentProps> =
  function InviteAdminContent(props: InviteAdminContentProps) {
    const { state, setState } = props.form;
    const { renderToString } = useContext(Context);

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "~/portal-admins",
          label: <FormattedMessage id="PortalAdminSettings.title" />,
        },
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
      <ScreenContent>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <FormTextField
          parentJSONPointer=""
          fieldName="inviteeEmail"
          className={styles.widget}
          type="text"
          label={renderToString("InviteAdminScreen.email.label")}
          value={state.email}
          onChange={onEmailChange}
        />
      </ScreenContent>
    );
  };

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
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
  });

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("./..");
    }
  }, [form.isSubmitted, navigate]);

  const errorRules: ErrorParseRule[] = useMemo(
    () => [
      makeReasonErrorParseRule(
        "CollaboratorInvitationDuplicate",
        "InviteAdminScreen.duplicated-error"
      ),
    ],
    []
  );

  const saveButtonProps = useMemo(
    () => ({
      labelId: "InviteAdminScreen.add-user.label",
      iconProps: {
        iconName: "Add",
      },
    }),
    []
  );

  return (
    <FormContainer
      form={form}
      saveButtonProps={saveButtonProps}
      errorRules={errorRules}
    >
      <InviteAdminContent form={form} />
    </FormContainer>
  );
};

export default InviteAdminScreen;
