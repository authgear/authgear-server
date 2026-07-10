import React, { useCallback, useMemo, useRef } from "react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { produce } from "immer";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import UserProfileAttributesList, {
  UserProfileAttributesListItem,
  UserProfileAttributesListSection,
  ItemComponentProps,
} from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  StandardAttributesAccessControlConfig,
} from "../../types";
import { parseJSONPointer } from "../../util/jsonpointer";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import styles from "./StandardAttributesConfigurationScreen.module.css";
import ExternalLink from "../../ExternalLink";

interface FormState {
  standardAttributesItems: StandardAttributesAccessControlConfig[];
}

interface StandardAttributesConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const standardAttributeSections: UserProfileAttributesListSection[] = [
  {
    key: "identity",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.identity-attributes",
    pointers: ["/email", "/phone_number", "/preferred_username"],
  },
  {
    key: "name",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.name-attributes",
    pointers: [
      "/name",
      "/given_name",
      "/family_name",
      "/middle_name",
      "/nickname",
    ],
  },
  {
    key: "profile",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.profile-attributes",
    pointers: [
      "/profile",
      "/picture",
      "/website",
      "/gender",
      "/birthdate",
      "/address",
    ],
  },
  {
    key: "local-preferences",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.local-preferences-attributes",
    pointers: ["/zoneinfo", "/locale"],
  },
];

const naturalOrder = standardAttributeSections.flatMap(
  (section) => section.pointers
);

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
  const messageId = "standard-attribute.description." + fieldName;

  const renderExternalLink = useCallback(
    (children: React.ReactNode) => (
      <ExternalLink href="https://en.wikipedia.org/wiki/List_of_tz_database_time_zones">
        {children}
      </ExternalLink>
    ),
    []
  );

  const values = useMemo(() => {
    if (messageId === "standard-attribute.description.zoneinfo") {
      return {
        externalLink: renderExternalLink,
      };
    }
    return {};
  }, [messageId, renderExternalLink]);

  return (
    <div className={className}>
      <Text as="p" size="2" weight="medium">
        <FormattedMessage id={"standard-attribute." + fieldName} />
      </Text>
      <Text as="p" size="1" color="gray">
        <FormattedMessage id={messageId} values={values} />
      </Text>
    </div>
  );
}

const StandardAttributesConfigurationScreenContent: React.VFC<StandardAttributesConfigurationScreenContentProps> =
  function StandardAttributesConfigurationScreenContent(props) {
    const { state, setState } = props.form;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);
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
      <ScreenContent
        layout="list"
        className={cn(
          styles.screenContent,
          isDirty ? styles.contentWithSaveBar : null
        )}
      >
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="StandardAttributesConfigurationScreen.title" />
          </Text>
        </div>
        <div className={cn(styles.widget, styles.tableWidget)}>
          <UserProfileAttributesList
            items={state.standardAttributesItems}
            onChangeItems={onChangeItems}
            ItemComponent={ItemComponent}
            sections={standardAttributeSections}
          />
        </div>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
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
        hideFooterComponent={true}
        canSave={true}
      >
        <StandardAttributesConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default StandardAttributesConfigurationScreen;
