import React, { useCallback, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, Text } from "@fluentui/react";
import ScreenHeader from "../../ScreenHeader";
import { useCreateAppMutation } from "./mutations/createAppMutation";
import { useTextField } from "../../hook/useInput";
import { useSystemConfig } from "../../context/SystemConfigContext";

import styles from "./CreateAppScreen.module.scss";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import { ErrorParseRule } from "../../error/parse";
import FormTextField from "../../FormTextField";
import OnboardingFormContainer from "./OnboardingFormContainer";

interface FormState {
  appID: string;
}

const defaultFormState: FormState = {
  appID: "",
};

const APP_ID_SCHEME = "https://";

const errorRules: ErrorParseRule[] = [
  {
    reason: "DuplicatedAppID",
    errorMessageID: "CreateAppScreen.error.duplicated-app-id",
  },
  {
    reason: "AppIDReserved",
    errorMessageID: "CreateAppScreen.error.reserved-app-id",
  },
  {
    reason: "InvalidAppID",
    errorMessageID: "CreateAppScreen.error.invalid-app-id",
  },
];

interface CreateAppContentProps {
  form: SimpleFormModel<FormState, string | null>;
}

const CreateAppContent: React.FC<CreateAppContentProps> = function CreateAppContent(
  props
) {
  const { state, setState } = props.form;
  const systemConfig = useSystemConfig();

  const { onChange: onAppIDChange } = useTextField((value) =>
    setState((prev) => ({ ...prev, appID: value }))
  );

  return (
    <main>
      <Text className={styles.pageTitle} block={true} variant="xLarge">
        <FormattedMessage id="CreateAppScreen.title" />
      </Text>
      <Text className={styles.pageDesc} block={true} variant="small">
        <FormattedMessage id="CreateAppScreen.desc" />
      </Text>
      <Label className={styles.fieldLabel}>
        <FormattedMessage id="CreateAppScreen.app-id.label" />
      </Label>
      <FormTextField
        className={styles.appIDField}
        parentJSONPointer="/"
        fieldName="app_id"
        value={state.appID}
        errorRules={errorRules}
        disabled={props.form.isUpdating}
        onChange={onAppIDChange}
        prefix={APP_ID_SCHEME}
        suffix={systemConfig.appHostSuffix}
      />
    </main>
  );
};

const CreateAppScreen: React.FC = function CreateAppScreen() {
  const navigate = useNavigate();
  const { createApp } = useCreateAppMutation();

  const submit = useCallback(
    async (state: FormState) => {
      return createApp(state.appID);
    },
    [createApp]
  );
  const form = useSimpleForm(defaultFormState, submit);

  useEffect(() => {
    if (form.submissionResult) {
      const appID = form.submissionResult;
      navigate(`/project/${encodeURIComponent(appID)}/onboarding`);
    }
  }, [form, navigate]);

  const saveButtonProps = useMemo(
    () => ({
      labelId: "create",
    }),
    []
  );

  return (
    <div className={styles.root}>
      <ScreenHeader />
      <OnboardingFormContainer form={form} saveButtonProps={saveButtonProps}>
        <CreateAppContent form={form} />
      </OnboardingFormContainer>
    </div>
  );
};

export default CreateAppScreen;
