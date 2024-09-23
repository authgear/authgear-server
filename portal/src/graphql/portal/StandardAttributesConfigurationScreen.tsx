import React, { useCallback } from "react";
import { useParams } from "react-router-dom";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { produce } from "immer";
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
  UserProfileAttributesListItem,
  ItemComponentProps,
} from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  StandardAttributesAccessControlConfig,
} from "../../types";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { parseJSONPointer } from "../../util/jsonpointer";
import styles from "./StandardAttributesConfigurationScreen.module.css";

interface FormState {
  standardAttributesItems: StandardAttributesAccessControlConfig[];
}

interface StandardAttributesConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const naturalOrder = [
  "/email",
  "/phone_number",
  "/preferred_username",
  "/name",
  "/given_name",
  "/family_name",
  "/middle_name",
  "/nickname",
  "/profile",
  "/picture",
  "/website",
  "/gender",
  "/birthdate",
  "/zoneinfo",
  "/locale",
  "/address",
];

function constructFormState(config: PortalAPIAppConfig): FormState {
  const items = config.user_profile?.standard_attributes?.access_control ?? [];
  const listedItems = items.filter((a) => naturalOrder.indexOf(a.pointer) >= 0);
  listedItems.sort((a, b) => {
    const ia = naturalOrder.indexOf(a.pointer);
    const ib = naturalOrder.indexOf(b.pointer);
    return ia - ib;
  });
  return {
    standardAttributesItems: listedItems,
  };
}

function constructConfig(
  rawConfig: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  const modifiedEffectiveConfig = produce(
    effectiveConfig,
    (effectiveConfig) => {
      effectiveConfig.user_profile ??= {};
      effectiveConfig.user_profile.standard_attributes ??= {};
      for (const accessControl of effectiveConfig.user_profile
        .standard_attributes.access_control ?? []) {
        for (const item of currentState.standardAttributesItems) {
          if (accessControl.pointer === item.pointer) {
            accessControl.access_control = item.access_control;
          }
        }
      }
    }
  );

  const accessControl =
    modifiedEffectiveConfig.user_profile?.standard_attributes?.access_control;
  return produce(rawConfig, (rawConfig) => {
    rawConfig.user_profile ??= {};
    rawConfig.user_profile.standard_attributes ??= {};
    rawConfig.user_profile.standard_attributes.access_control = accessControl;
  });
}

function ItemComponent(
  props: ItemComponentProps<StandardAttributesAccessControlConfig>
) {
  const { className, item } = props;
  const { pointer } = item;
  const fieldName = parseJSONPointer(pointer)[0];
  const { themes } = useSystemConfig();
  const descriptionColor = themes.main.palette.neutralTertiary;
  return (
    <div className={className}>
      <Text className={styles.fieldName} block={true}>
        <FormattedMessage id={"standard-attribute." + fieldName} />
      </Text>
      <Text
        variant="small"
        block={true}
        style={{
          color: descriptionColor,
        }}
      >
        <FormattedMessage id={"standard-attribute.description." + fieldName} />
      </Text>
    </div>
  );
}

const StandardAttributesConfigurationScreenContent: React.VFC<StandardAttributesConfigurationScreenContentProps> =
  function StandardAttributesConfigurationScreenContent(props) {
    const { state, setState } = props.form;
    const onChangeItems = useCallback(
      (newItems: UserProfileAttributesListItem[]) => {
        setState((prev) => {
          return {
            ...prev,
            standardAttributesItems: newItems,
          };
        });
      },
      [setState]
    );
    return (
      <>
        <ScreenContent layout="list">
          <ScreenTitle className={styles.widget}>
            <FormattedMessage id="StandardAttributesConfigurationScreen.title" />
          </ScreenTitle>
          <div className={styles.widget}>
            <UserProfileAttributesList
              items={state.standardAttributesItems}
              onChangeItems={onChangeItems}
              ItemComponent={ItemComponent}
            />
          </div>
        </ScreenContent>
      </>
    );
  };

const StandardAttributesConfigurationScreen: React.VFC =
  function StandardAttributesConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer
        form={form}
        stickyFooterComponent={true}
        showDiscardButton={true}
      >
        <StandardAttributesConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default StandardAttributesConfigurationScreen;
