import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Dropdown, IDropdownOption, Toggle } from "@fluentui/react";
import cn from "classnames";
import {
  isPromotionConflictBehaviour,
  PortalAPIAppConfig,
  PromotionConflictBehaviour,
  promotionConflictBehaviours,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import styles from "./AnonymousUsersConfigurationScreen.module.scss";

const dropDownStyles = {
  dropdown: {
    width: "300px",
  },
};

interface FormState {
  enabled: boolean;
  promotionConflictBehaviour: PromotionConflictBehaviour;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const enabled =
    config.authentication?.identities?.includes("anonymous") ?? false;
  const promotionConflictBehaviour =
    config.identity?.on_conflict?.promotion ?? "error";
  return { enabled, promotionConflictBehaviour };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    if (initialState.enabled !== currentState.enabled) {
      const identities = (
        effectiveConfig.authentication?.identities ?? []
      ).slice();
      const index = identities.indexOf("anonymous");
      if (currentState.enabled && index === -1) {
        identities.push("anonymous");
      } else if (!currentState.enabled && index >= 0) {
        identities.splice(index, 1);
      }
      config.authentication ??= {};
      config.authentication.identities = identities;
    }
    if (
      currentState.enabled &&
      initialState.promotionConflictBehaviour !==
        currentState.promotionConflictBehaviour
    ) {
      config.identity ??= {};
      config.identity.on_conflict ??= {};
      config.identity.on_conflict.promotion =
        currentState.promotionConflictBehaviour;
    }
    clearEmptyObject(config);
  });
}

const conflictBehaviourMessageId: Record<PromotionConflictBehaviour, string> = {
  login: "AnonymousIdentityConflictBehaviour.login",
  error: "AnonymousIdentityConflictBehaviour.error",
};

interface AnonymousUserConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const AnonymousUserConfigurationContent: React.FC<AnonymousUserConfigurationContentProps> =
  function AnonymousUserConfigurationContent(props) {
    const { state, setState } = props.form;

    const { renderToString } = useContext(Context);

    const conflictBehaviourOptions = useMemo(
      () =>
        promotionConflictBehaviours.map((behaviour) => {
          const selectedBehaviour = state.promotionConflictBehaviour;
          return {
            key: behaviour,
            text: renderToString(conflictBehaviourMessageId[behaviour]),
            isSelected: selectedBehaviour === behaviour,
          };
        }),
      [state, renderToString]
    );

    const onEnableChange = useCallback(
      (_event, checked?: boolean) =>
        setState((state) => ({
          ...state,
          enabled: checked ?? false,
        })),
      [setState]
    );

    const onConflictOptionChange = useCallback(
      (_event, option?: IDropdownOption) => {
        const key = option?.key;
        if (key && isPromotionConflictBehaviour(key)) {
          setState((state) => ({
            ...state,
            promotionConflictBehaviour: key,
          }));
        }
      },
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="AnonymousUsersConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={cn(styles.widget, styles.controlGroup)}>
          <WidgetTitle>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
          </WidgetTitle>
          <Toggle
            className={styles.control}
            checked={state.enabled}
            onChange={onEnableChange}
            label={renderToString(
              "AnonymousUsersConfigurationScreen.enable.label"
            )}
            inlineLabel={true}
          />
          <Dropdown
            className={styles.control}
            styles={dropDownStyles}
            label={renderToString(
              "AnonymousUsersConfigurationScreen.conflict-droplist.label"
            )}
            disabled={!state.enabled}
            options={conflictBehaviourOptions}
            selectedKey={state.promotionConflictBehaviour}
            onChange={onConflictOptionChange}
          />
        </Widget>
      </ScreenContent>
    );
  };

const AnonymousUserConfigurationScreen: React.FC =
  function AnonymousUserConfigurationScreen() {
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
        <AnonymousUserConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default AnonymousUserConfigurationScreen;
