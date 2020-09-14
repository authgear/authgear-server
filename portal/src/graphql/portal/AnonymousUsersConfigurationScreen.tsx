import React, { useContext, useCallback, useState, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Text, Toggle, Dropdown, IDropdownOption } from "@fluentui/react";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { AppConfigQuery } from "./query/__generated__/AppConfigQuery";
import {
  PortalAPIApp,
  PortalAPIAppConfig,
  PromotionConflictBehaviour,
  promotionConflictBehaviours,
  isPromotionConflictBehaviour,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ButtonWithLoading from "../../ButtonWithLoading";

import styles from "./AnonymousUsersConfigurationScreen.module.scss";

interface AnonymousUsersConfigurationScreenState {
  enabled: boolean;
  promotionConflictBehaviour: PromotionConflictBehaviour;
}

interface AnonymousUsersConfigurationProps {
  data?: AppConfigQuery;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
}

const DEFAULT_CONFLICT_BEHAVIOUR: PromotionConflictBehaviour = "error";

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): AnonymousUsersConfigurationScreenState {
  if (appConfig == null) {
    return {
      enabled: false,
      promotionConflictBehaviour: DEFAULT_CONFLICT_BEHAVIOUR,
    };
  }
  const anonymousUserEnabled =
    appConfig.authentication?.identities?.find(
      (identity) => identity === "anonymous"
    ) != null;
  const promotionConflictBehaviour =
    appConfig.identity?.on_conflict?.promotion ?? DEFAULT_CONFLICT_BEHAVIOUR;
  return {
    enabled: anonymousUserEnabled,
    promotionConflictBehaviour,
  };
}

function constructNewAppConfigFromState(
  state: AnonymousUsersConfigurationScreenState,
  initialState: AnonymousUsersConfigurationScreenState,
  appConfig: PortalAPIAppConfig
) {
  return produce(appConfig, (draftConfig) => {
    draftConfig.identity = draftConfig.identity ?? {};
    draftConfig.identity.on_conflict = draftConfig.identity.on_conflict ?? {};
    const onConflict = draftConfig.identity.on_conflict;

    draftConfig.authentication = draftConfig.authentication ?? {};
    draftConfig.authentication.identities =
      draftConfig.authentication.identities ?? [];
    const { authentication } = draftConfig;
    const authenticationIdentitiesSet = new Set(authentication.identities);

    const enabledStateChanged = state.enabled !== initialState.enabled;
    const behaviourStateChanged =
      state.promotionConflictBehaviour !==
      initialState.promotionConflictBehaviour;

    if (state.enabled) {
      if (enabledStateChanged) {
        authenticationIdentitiesSet.add("anonymous");
        authentication.identities = Array.from(authenticationIdentitiesSet);
      }
      if (behaviourStateChanged) {
        onConflict.promotion = state.promotionConflictBehaviour;
      }
    } else {
      if (enabledStateChanged) {
        authenticationIdentitiesSet.delete("anonymous");
        authentication.identities = Array.from(authenticationIdentitiesSet);
      }
    }

    clearEmptyObject(draftConfig);
  });
}

function constructConflictBehaviourOptions(
  state: AnonymousUsersConfigurationScreenState
): IDropdownOption[] {
  return promotionConflictBehaviours.map((behaviour: string) => {
    const selectedBehaviour = state.promotionConflictBehaviour;
    return {
      key: behaviour,
      text: behaviour,
      isSelected: selectedBehaviour === behaviour,
    };
  });
}

const AnonymousUsersConfiguration: React.FC<AnonymousUsersConfigurationProps> = function AnonymousUsersConfiguration(
  props: AnonymousUsersConfigurationProps
) {
  const { data, updateAppConfig, updatingAppConfig } = props;
  const { renderToString } = useContext(Context);

  const effectiveAppConfig: PortalAPIAppConfig | null =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;
  const rawAppConfig: PortalAPIAppConfig | null =
    data?.node?.__typename === "App" ? data.node.rawAppConfig : null;

  const initialState = useMemo(
    () => constructStateFromAppConfig(effectiveAppConfig),
    [effectiveAppConfig]
  );

  const [state, setState] = useState(initialState);
  const conflictBehaviourOptions = useMemo(() => {
    return constructConflictBehaviourOptions(state);
  }, [state]);

  const onSwitchToggled = useCallback(
    (_event, checked?: boolean) => {
      if (checked == null) {
        return;
      }
      setState({
        ...state,
        enabled: checked,
      });
    },
    [state]
  );

  const onConflictOptionChange = useCallback(
    (_event, option?: IDropdownOption) => {
      if (option != null && isPromotionConflictBehaviour(option.key)) {
        setState({
          ...state,
          promotionConflictBehaviour: option.key,
        });
      }
    },
    [state]
  );

  const onSaveClicked = useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }
    const newAppConfig = constructNewAppConfigFromState(
      state,
      initialState,
      rawAppConfig
    );
    // TODO: handle error
    updateAppConfig(newAppConfig).catch(() => {});
  }, [updateAppConfig, rawAppConfig, initialState, state]);

  return (
    <section className={styles.screenContent}>
      <Toggle
        className={styles.enableToggle}
        checked={state.enabled}
        onChange={onSwitchToggled}
        label={renderToString("AnonymousUsersConfigurationScreen.enable.label")}
        inlineLabel={true}
      />
      <Dropdown
        className={styles.conflictDropdown}
        label={renderToString(
          "AnonymousUsersConfigurationScreen.conflict-droplist.label"
        )}
        disabled={!state.enabled}
        options={conflictBehaviourOptions}
        onChange={onConflictOptionChange}
      />
      <ButtonWithLoading
        className={styles.saveButton}
        onClick={onSaveClicked}
        loading={updatingAppConfig}
        labelId="save"
        loadingLabelId="saving"
      />
    </section>
  );
};

const AnonymousUserConfigurationScreen: React.FC = function AnonymousUserConfigurationScreen() {
  const { appID } = useParams();
  const { loading, error, data, refetch } = useAppConfigQuery(appID);
  const {
    loading: updatingAppConfig,
    error: updateAppConfigError,
    updateAppConfig,
  } = useUpdateAppConfigMutation(appID);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <Text as="h1" className={styles.title}>
        <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
      </Text>
      <AnonymousUsersConfiguration
        data={data}
        updateAppConfig={updateAppConfig}
        updatingAppConfig={updatingAppConfig}
      />
    </main>
  );
};

export default AnonymousUserConfigurationScreen;
