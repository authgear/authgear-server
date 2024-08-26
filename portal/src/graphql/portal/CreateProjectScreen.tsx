import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import WizardScreenLayout from "../../WizardScreenLayout";
import WizardContentLayout, { WizardTitle } from "../../WizardContentLayout";
import FormTextField from "../../FormTextField";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import { FormProvider } from "../../form";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useCreateAppMutation } from "./mutations/createAppMutation";
import { useAppListQuery } from "./query/appListQuery";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import { useSimpleForm } from "../../hook/useSimpleForm";
import {
  randomProjectName,
  projectNameWithCompanyName,
} from "../../util/projectname";
import PrimaryButton from "../../PrimaryButton";

interface FormState {
  appID: string;
}

function makeDefaultState(): FormState {
  return {
    appID: randomProjectName(),
  };
}

const FORM_TEXT_FIELD_STYLES = {
  description: {
    display: "block",
    marginTop: "10px",
    fontSize: "12px",
  },
};

const APP_ID_SCHEME = "https://";

const errorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "DuplicatedAppID",
    "CreateProjectScreen.error.duplicated-app-id"
  ),
  makeReasonErrorParseRule(
    "AppIDReserved",
    "CreateProjectScreen.error.reserved-app-id"
  ),
  makeReasonErrorParseRule(
    "InvalidAppID",
    "CreateProjectScreen.error.invalid-app-id"
  ),
];

interface CreateProjectScreenContentProps {
  numberOfApps: number;
}

interface LocationState {
  company_name: string;
}

function processCompanyName(companyName: string): string {
  return companyName
    .trim()
    .split("")
    .filter((char) => /[a-zA-Z\s]/.exec(char))
    .join("")
    .split(" ")
    .filter((word) => word !== "")
    .join("-")
    .toLowerCase();
}

function CreateProjectScreenContent(props: CreateProjectScreenContentProps) {
  const { numberOfApps } = props;
  const navigate = useNavigate();
  const { appHostSuffix } = useSystemConfig();
  const { createApp } = useCreateAppMutation();
  const { renderToString } = useContext(Context);
  const { state } = useLocation();

  const submit = useCallback(
    async (state: FormState) => {
      return createApp(state.appID);
    },
    [createApp]
  );

  const defaultState = useMemo(() => {
    const typedState: LocationState | null = state as LocationState | null;
    const defaultState = makeDefaultState();
    if (typedState) {
      const intermediateName = processCompanyName(typedState.company_name);
      if (intermediateName !== "")
        defaultState.appID = projectNameWithCompanyName(intermediateName);
    }
    return defaultState;
  }, [state]);

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState: defaultState,
    submit,
  });

  const {
    updateError,
    save,
    isUpdating,
    state: { appID },
    setState,
  } = form;

  const onChangeAppID = useCallback(
    (_e, newValue) => {
      if (newValue != null) {
        setState((prev) => ({ ...prev, appID: newValue }));
      }
    },
    [setState]
  );

  const onSubmitForm = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      save().catch(() => {});
    },
    [save]
  );

  useEffect(() => {
    if (form.submissionResult) {
      const appID = form.submissionResult;
      // assign app id to the event after creating the app
      navigate(`/project/${encodeURIComponent(appID)}/wizard`);
    }
  }, [form, navigate]);

  return (
    <FormProvider loading={isUpdating} error={updateError}>
      <WizardScreenLayout>
        <FormErrorMessageBar />
        <WizardContentLayout
          backButtonDisabled={true}
          primaryButton={
            <PrimaryButton
              disabled={isUpdating}
              onClick={onSubmitForm}
              text={
                <FormattedMessage id="CreateProjectScreen.create-project.label" />
              }
            />
          }
        >
          <WizardTitle>
            <FormattedMessage
              id="CreateProjectScreen.title"
              values={{
                apps: numberOfApps,
              }}
            />
          </WizardTitle>
          <form onSubmit={onSubmitForm}>
            <FormTextField
              styles={FORM_TEXT_FIELD_STYLES}
              parentJSONPointer=""
              fieldName="app_id"
              value={appID}
              onChange={onChangeAppID}
              errorRules={errorRules}
              prefix={APP_ID_SCHEME}
              suffix={appHostSuffix}
              label={renderToString("CreateProjectScreen.app-id.label")}
              description={renderToString(
                "CreateProjectScreen.app-id.description"
              )}
            />
          </form>
        </WizardContentLayout>
      </WizardScreenLayout>
    </FormProvider>
  );
}

const CreateProjectScreen: React.VFC = function CreateProjectScreen() {
  const { loading, error, apps, refetch } = useAppListQuery();

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <CreateProjectScreenContent numberOfApps={apps?.length ?? 0} />;
};

export default CreateProjectScreen;
