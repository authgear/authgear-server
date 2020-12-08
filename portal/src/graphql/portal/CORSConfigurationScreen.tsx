import React, { useCallback, useMemo } from "react";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import produce from "immer";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormTextFieldList from "../../FormTextFieldList";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

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

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="CORSConfigurationScreen.title" />,
      },
    ];
  }, []);

  const onAllowedOriginsChange = useCallback(
    (allowedOrigins: string[]) => {
      setState((state) => ({ ...state, allowedOrigins }));
    },
    [setState]
  );

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <Text className={styles.description}>
        <FormattedMessage id="CORSConfigurationScreen.desc" />
      </Text>
      <FormTextFieldList
        className={styles.fieldList}
        parentJSONPointer="/http"
        fieldName="allowed_origins"
        list={state.allowedOrigins}
        onListChange={onAllowedOriginsChange}
        addButtonLabelMessageID="add"
      />
    </div>
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
