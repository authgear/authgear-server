import React, { useCallback } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import produce from "immer";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import UserProfileAttributesList from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  CustomAttributesAttributeConfig,
} from "../../types";
import styles from "./CustomAttributesConfigurationScreen.module.scss";

interface FormState {
  items: CustomAttributesAttributeConfig[];
}

interface CustomAttributesConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const items = config.user_profile?.custom_attributes?.attributes ?? [];
  return {
    items,
  };
}

function constructConfig(
  rawConfig: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(rawConfig, (rawConfig) => {
    rawConfig.user_profile ??= {};
    rawConfig.user_profile.custom_attributes ??= {};
    rawConfig.user_profile.custom_attributes.attributes = currentState.items;
  });
}

const CustomAttributesConfigurationScreenContent: React.FC<CustomAttributesConfigurationScreenContentProps> =
  function CustomAttributesConfigurationScreenContent(props) {
    const { state, setState } = props.form;
    const onChangeItems = useCallback(
      (newItems: CustomAttributesAttributeConfig[]) => {
        setState((prev) => {
          return {
            ...prev,
            items: newItems,
          };
        });
      },
      [setState]
    );
    return (
      <>
        <ScreenContent>
          <ScreenTitle className={styles.widget}>
            <FormattedMessage id="CustomAttributesConfigurationScreen.title" />
          </ScreenTitle>
          <div className={styles.widget}>
            <UserProfileAttributesList
              items={state.items}
              onChangeItems={onChangeItems}
            />
          </div>
        </ScreenContent>
      </>
    );
  };

const CustomAttributesConfigurationScreen: React.FC =
  function CustomAttributesConfigurationScreen() {
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
        <CustomAttributesConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default CustomAttributesConfigurationScreen;
