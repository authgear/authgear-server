import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Dropdown, IDropdownOption, Toggle } from "@fluentui/react";
import {
  isPromotionConflictBehaviour,
  PortalAPIAppConfig,
  PromotionConflictBehaviour,
  promotionConflictBehaviours,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

import styles from "./AnonymousUsersConfigurationScreen.module.scss";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import FormContainer from "../../FormContainer";

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

const AnonymousUserConfigurationContent: React.FC<AnonymousUserConfigurationContentProps> = function AnonymousUserConfigurationContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: (
          <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
        ),
      },
    ];
  }, []);

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
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <Toggle
        className={styles.toggle}
        checked={state.enabled}
        onChange={onEnableChange}
        label={renderToString("AnonymousUsersConfigurationScreen.enable.label")}
        inlineLabel={true}
      />
      <Dropdown
        className={styles.dropdown}
        styles={dropDownStyles}
        label={renderToString(
          "AnonymousUsersConfigurationScreen.conflict-droplist.label"
        )}
        disabled={!state.enabled}
        options={conflictBehaviourOptions}
        selectedKey={state.promotionConflictBehaviour}
        onChange={onConflictOptionChange}
      />
    </div>
  );
};

const AnonymousUserConfigurationScreen: React.FC = function AnonymousUserConfigurationScreen() {
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
