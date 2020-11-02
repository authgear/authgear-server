import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { ITextFieldProps } from "@fluentui/react";
import { useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";

import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import PasswordField, { localValidatePassword } from "../../PasswordField";
import ButtonWithLoading from "../../ButtonWithLoading";
import { CreateIdentityFunction } from "./mutations/createIdentityMutation";
import { nonNullable } from "../../util/types";
import { AuthenticatorType } from "./__generated__/globalTypes";
import { PortalAPIAppConfig } from "../../types";

import styles from "./AddIdentityForm.module.scss";

interface AddIdentityFormProps {
  className?: string;
  appConfig: PortalAPIAppConfig | null;
  user: UserQuery_node_User | null;
  loginIDKey: "username" | "email" | "phone";
  loginID: string;
  loginIDField: React.ReactNode;
  password: string;
  onPasswordChange: ITextFieldProps["onChange"];
  passwordFieldErrorMessage?: string;
  isFormModified: boolean;
  createIdentity: CreateIdentityFunction;
  creatingIdentity: boolean;
}

function determineIsPasswordRequired(user: UserQuery_node_User | null) {
  const authenticators =
    user?.authenticators?.edges
      ?.map((edge) => edge?.node?.type)
      .filter(nonNullable) ?? [];
  const hasPasswordAuthenticator = authenticators.includes(
    AuthenticatorType.PASSWORD
  );
  return !hasPasswordAuthenticator;
}

const AddIdentityForm: React.FC<AddIdentityFormProps> = function AddIdentityForm(
  props: AddIdentityFormProps
) {
  const {
    className,
    appConfig,
    user,
    loginIDKey,
    loginID,
    loginIDField,
    password,
    onPasswordChange,
    passwordFieldErrorMessage,
    isFormModified,
    createIdentity,
    creatingIdentity,
  } = props;

  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const isPasswordRequired = useMemo(() => {
    return determineIsPasswordRequired(user);
  }, [user]);

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  const [localValidationErrorMessage, setLocalViolationErrorMessage] = useState<
    string | undefined
  >(undefined);
  const [submittedForm, setSubmittedForm] = useState(false);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      if (isPasswordRequired) {
        const localErrorMessageMap = localValidatePassword(
          renderToString,
          passwordPolicy,
          password
        );
        setLocalViolationErrorMessage(localErrorMessageMap?.password);

        if (localErrorMessageMap != null) {
          return;
        }
      }

      const requestPassword = isPasswordRequired ? password : undefined;
      createIdentity({ key: loginIDKey, value: loginID }, requestPassword)
        .then((identity) => {
          if (identity != null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [
      renderToString,
      loginIDKey,
      loginID,
      createIdentity,
      isPasswordRequired,
      password,
      passwordPolicy,
    ]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("..#connected-identities");
    }
  }, [submittedForm, navigate]);

  return (
    <form className={className} onSubmit={onFormSubmit}>
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
      {loginIDField}
      {isPasswordRequired && (
        <PasswordField
          className={styles.password}
          textFieldClassName={styles.passwordField}
          passwordPolicy={passwordPolicy}
          label={renderToString("AddUsernameScreen.password.label")}
          value={password}
          onChange={onPasswordChange}
          errorMessage={
            localValidationErrorMessage ?? passwordFieldErrorMessage
          }
        />
      )}
      <ButtonWithLoading
        type="submit"
        disabled={!isFormModified || submittedForm}
        labelId="add"
        loading={creatingIdentity}
      />
    </form>
  );
};

export default AddIdentityForm;
