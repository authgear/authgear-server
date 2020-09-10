import React, { useContext, useCallback, useState, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  Text,
  Toggle,
  Dropdown,
  PrimaryButton,
  IDropdownOption,
} from "@fluentui/react";

import { useAppConfigQuery } from "./query/appConfigQuery";
import {
  PortalAPIAppConfig,
  PromotionConflictBehaviour,
  promotionConflictBehaviours,
  isPromotionConflictBehaviour,
} from "../../types";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

import styles from "./AnonymousUsersConfigurationScreen.module.scss";

interface AnonymousUsersConfigurationScreenState {
  enabled: boolean;
  promotionConflictBehaviour?: PromotionConflictBehaviour;
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): AnonymousUsersConfigurationScreenState {
  if (appConfig == null) {
    return {
      enabled: false,
    };
  }
  const promotionConflictBehaviour = appConfig.identity?.on_conflict?.promotion;
  return {
    enabled: promotionConflictBehaviour != null,
    promotionConflictBehaviour,
  };
}

function constructNewAppConfigFromState(
  state: AnonymousUsersConfigurationScreenState,
  appConfig: PortalAPIAppConfig
) {
  return produce(appConfig, (draftConfig) => {
    draftConfig.identity = draftConfig.identity ?? {};
    draftConfig.identity.on_conflict = draftConfig.identity.on_conflict ?? {};

    const onConflict = draftConfig.identity.on_conflict;
    if (state.enabled && state.promotionConflictBehaviour != null) {
      onConflict.promotion = state.promotionConflictBehaviour;
    } else {
      delete onConflict.promotion;
    }
  });
}

function constructConflictBehaviourOptions(
  state: AnonymousUsersConfigurationScreenState
): IDropdownOption[] {
  return promotionConflictBehaviours.map((behaviour: string) => {
    const selectedBehaviour = state.enabled
      ? state.promotionConflictBehaviour
      : null;
    return {
      key: behaviour,
      text: behaviour,
      isSelected: selectedBehaviour === behaviour,
    };
  });
}

const AnonymousUserConfigurationScreen: React.FC = function AnonymousUserConfigurationScreen() {
  const { appID } = useParams();
  const { loading, error, data, refetch } = useAppConfigQuery(appID);
  const { renderToString } = useContext(Context);

  const appConfig: PortalAPIAppConfig | null =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;

  const [state, setState] = useState(constructStateFromAppConfig(appConfig));

  const conflictBehaviourOptions = useMemo(() => {
    return constructConflictBehaviourOptions(state);
  }, [state]);

  const onSwitchToggled = useCallback(
    (checked: boolean) => {
      setState({
        ...state,
        enabled: checked,
      });
    },
    [state]
  );

  const onConflictOptionChanged = useCallback(
    (option: IDropdownOption) => {
      setState({
        ...state,
        promotionConflictBehaviour: isPromotionConflictBehaviour(option.key)
          ? option.key
          : undefined,
      });
    },
    [state]
  );

  const onSaveClicked = useCallback(() => {
    if (appConfig == null) {
      return;
    }
    constructNewAppConfigFromState(state, appConfig);
    // TODO: call mutation to save config
  }, [appConfig, state]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      <Text as="h1" className={styles.title}>
        <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
      </Text>
      <section className={styles.screenContent}>
        <Toggle
          className={styles.enableToggle}
          checked={state.enabled}
          onChanged={onSwitchToggled}
          label={renderToString(
            "AnonymousUsersConfigurationScreen.enable.label"
          )}
          inlineLabel={true}
        />
        <Dropdown
          className={styles.conflictDropdown}
          label={renderToString(
            "AnonymousUsersConfigurationScreen.conflict-droplist.label"
          )}
          disabled={!state.enabled}
          options={conflictBehaviourOptions}
          onChanged={onConflictOptionChanged}
        />
        <PrimaryButton className={styles.saveButton} onClick={onSaveClicked}>
          <FormattedMessage id="save" />
        </PrimaryButton>
      </section>
    </main>
  );
};

export default AnonymousUserConfigurationScreen;
