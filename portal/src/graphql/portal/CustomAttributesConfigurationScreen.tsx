import React, { useCallback } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import produce from "immer";
import cn from "classnames";
import { Text, PrimaryButton } from "@fluentui/react";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import UserProfileAttributesList, {
  ItemComponentProps,
} from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  CustomAttributesAttributeConfig,
} from "../../types";
import { parseJSONPointer } from "../../util/jsonpointer";
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

function EmptyState() {
  return (
    <div className={styles.emptyState}>
      <Text className={styles.emptyStateMessage} block={true}>
        <FormattedMessage id="CustomAttributesConfigurationScreen.empty-message" />
      </Text>
      <PrimaryButton className={styles.addNewAttributeButton}>
        <FormattedMessage id="CustomAttributesConfigurationScreen.label.add-new-attribute" />
      </PrimaryButton>
    </div>
  );
}

function ItemComponent(
  props: ItemComponentProps<CustomAttributesAttributeConfig>
) {
  const { className, item } = props;
  const { pointer } = item;
  const fieldName = parseJSONPointer(pointer)[0];
  return (
    <Text className={cn(className, styles.fieldName)} block={true}>
      {fieldName}
    </Text>
  );
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
              ItemComponent={ItemComponent}
            />
            {state.items.length <= 0 ? <EmptyState /> : null}
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
