import React, { useCallback } from "react";
import cn from "classnames";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import FormTextFieldList from "../../FormTextFieldList";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";

import styles from "./CORSConfigurationScreen.module.scss";

interface FormState {
  allowedOrigins: string[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    allowedOrigins: config.http?.allowed_origins ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.http ??= {};
    config.http.allowed_origins = currentState.allowedOrigins;
    clearEmptyObject(config);
  });
}

interface CORSConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const CORSConfigurationContent: React.FC<CORSConfigurationContentProps> = function CORSConfigurationContent(
  props
) {
  const { state, setState } = props.form;

  const onAllowedOriginsChange = useCallback(
    (allowedOrigins: string[]) => {
      setState((state) => ({ ...state, allowedOrigins }));
    },
    [setState]
  );

  return (
    <ScreenContent className={styles.root}>
      <ScreenTitle>
        <FormattedMessage id="CORSConfigurationScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="CORSConfigurationScreen.desc" />
      </ScreenDescription>
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle>
          <FormattedMessage id="CORSConfigurationScreen.title" />
        </WidgetTitle>
        <FormTextFieldList
          className={styles.control}
          parentJSONPointer="/http"
          fieldName="allowed_origins"
          list={state.allowedOrigins}
          onListChange={onAllowedOriginsChange}
          addButtonLabelMessageID="add"
        />
      </Widget>
    </ScreenContent>
  );
};

const CORSConfigurationScreen: React.FC = function CORSConfigurationScreen() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <CORSConfigurationContent form={form} />
    </FormContainer>
  );
};

export default CORSConfigurationScreen;
