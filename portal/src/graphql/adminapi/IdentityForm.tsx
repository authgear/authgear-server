import React, { useCallback, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { LoginIDKeyType, PortalAPIAppConfig } from "../../types";
import { AuthenticatorKind, AuthenticatorType } from "./globalTypes.generated";

import styles from "./IdentityForm.module.css";
import { useSimpleForm } from "../../hook/useSimpleForm";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import { ErrorParseRule } from "../../error/parse";
import { canCreateLoginIDIdentity } from "../../util/loginID";
import { Text } from "@fluentui/react";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import { useUpdateLoginIDIdentityMutation } from "./mutations/updateIdentityMutation";

interface FormState {
  loginID: string;
}

const defaultState: FormState = {
  loginID: "",
};

interface User {
  id: string;
  primaryAuthenticators: AuthenticatorType[];
}

export interface LoginIDFieldProps {
  value: string;
  onChange: (value: string) => void;
}

interface IdentityFormProps {
  originalIdentityID: string | null;
  currentValueMessage?: React.ReactNode;
  appConfig: PortalAPIAppConfig | null;
  rawUser: UserQueryNodeFragment | null;
  loginIDType: LoginIDKeyType;
  title: React.ReactNode;
  loginIDField: React.ComponentType<LoginIDFieldProps>;
  errorRules?: ErrorParseRule[];
  onReset?: () => void;
}

const IdentityForm: React.VFC<IdentityFormProps> = function IdentityForm(
  props: IdentityFormProps
) {
  const {
    originalIdentityID,
    currentValueMessage,
    appConfig,
    rawUser,
    loginIDType,
    title,

    loginIDField: LoginIDField,
    onReset,
  } = props;

  const navigate = useNavigate();

  const user: User = useMemo(() => {
    if (!rawUser) {
      return { id: "", primaryAuthenticators: [] };
    }
    const authenticators =
      rawUser.authenticators?.edges?.map((e) => e?.node) ?? [];
    return {
      id: rawUser.id,
      primaryAuthenticators: authenticators
        .filter((a) => a?.kind === AuthenticatorKind.Primary)
        .map((a) => a!.type),
    };
  }, [rawUser]);

  const { createIdentity } = useCreateLoginIDIdentityMutation(user.id);
  const { updateIdentity } = useUpdateLoginIDIdentityMutation(user.id);

  const validate = useCallback(() => {
    return null;
  }, []);

  const submit = useCallback(
    async (state: FormState) => {
      if (originalIdentityID) {
        await updateIdentity(originalIdentityID, {
          key: loginIDType,
          value: state.loginID,
        });
      } else {
        await createIdentity(
          { key: loginIDType, value: state.loginID },
          undefined
        );
      }
    },
    [originalIdentityID, updateIdentity, loginIDType, createIdentity]
  );

  const rawForm = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
  });
  const form = useMemo(
    () => ({
      ...rawForm,
      reset: () => {
        rawForm.reset();
        onReset?.();
      },
    }),
    [rawForm, onReset]
  );

  useEffect(() => {
    if (form.isSubmitted) {
      if (originalIdentityID == null) {
        navigate("./..#connected-identities");
      } else {
        navigate("./../..#connected-identities");
      }
    }
  }, [form.isSubmitted, navigate, originalIdentityID]);

  const onLoginIDChange = useCallback(
    (value: string) => form.setState((state) => ({ ...state, loginID: value })),
    [form]
  );

  const canSave = form.state.loginID.length > 0;

  if (!canCreateLoginIDIdentity(appConfig)) {
    return (
      <Text className={styles.helpText}>
        <FormattedMessage id="CreateIdentity.require-login-id" />
      </Text>
    );
  }

  return (
    <FormContainer form={form} canSave={canSave}>
      <ScreenContent>
        {title}
        {currentValueMessage != null ? (
          <div className={styles.currentValue}>
            <Text>{currentValueMessage}</Text>
          </div>
        ) : null}
        <LoginIDField value={form.state.loginID} onChange={onLoginIDChange} />
      </ScreenContent>
    </FormContainer>
  );
};

export default IdentityForm;
