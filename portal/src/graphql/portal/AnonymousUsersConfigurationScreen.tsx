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
import { AppConfigQuery } from "./query/__generated__/AppConfigQuery";
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
  promotionConflictBehaviour: PromotionConflictBehaviour;
}

interface AnonymousUsersConfigurationProps {
  data?: AppConfigQuery;
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

    if (state.enabled) {
      authenticationIdentitiesSet.add("anonymous");
      authentication.identities = Array.from(authenticationIdentitiesSet);
      onConflict.promotion = state.promotionConflictBehaviour;
    } else {
      authenticationIdentitiesSet.delete("anonymous");
      authentication.identities = Array.from(authenticationIdentitiesSet);
    }
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
  const { data } = props;
  const { renderToString } = useContext(Context);

  const appConfig: PortalAPIAppConfig | null =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;

  const [state, setState] = useState(constructStateFromAppConfig(appConfig));
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
    if (appConfig == null) {
      return;
    }
    constructNewAppConfigFromState(state, appConfig);
    // TODO: call mutation to save config
  }, [appConfig, state]);

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
      <PrimaryButton className={styles.saveButton} onClick={onSaveClicked}>
        <FormattedMessage id="save" />
      </PrimaryButton>
    </section>
  );
};

const AnonymousUserConfigurationScreen: React.FC = function AnonymousUserConfigurationScreen() {
  const { appID } = useParams();
  const { loading, error, data, refetch } = useAppConfigQuery(appID);

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
      <AnonymousUsersConfiguration data={data} />
    </main>
  );
};

export default AnonymousUserConfigurationScreen;
