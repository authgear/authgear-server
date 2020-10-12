import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import produce from "immer";
import deepEqual from "deep-equal";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import { clearEmptyObject } from "../../util/misc";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { PortalAPIApp, PortalAPIAppConfig } from "../../types";

import styles from "./UserInterfaceScreen.module.scss";

interface UserInterfaceScreenState {
  customCss: string;
}

interface UserInterfaceProps {
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): UserInterfaceScreenState {
  return {
    customCss: appConfig?.ui?.custom_css ?? "",
  };
}

function constructNewAppConfigFromState(
  state: UserInterfaceScreenState,
  initialState: UserInterfaceScreenState,
  appConfig: PortalAPIAppConfig
) {
  return produce(appConfig, (draftConfig) => {
    draftConfig.ui = draftConfig.ui ?? {};

    if (state.customCss !== initialState.customCss) {
      draftConfig.ui.custom_css = state.customCss;
    }

    clearEmptyObject(draftConfig);
  });
}

const UserInterface: React.FC<UserInterfaceProps> = function UserInterface(
  props: UserInterfaceProps
) {
  const {
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
  } = props;

  const initialState = useMemo(
    () => constructStateFromAppConfig(effectiveAppConfig),
    [effectiveAppConfig]
  );

  const [state, setState] = useState<UserInterfaceScreenState>(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onCustomCssChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        customCss: value,
      }));
    },
    []
  );

  const onSaveButtonClicked = useCallback(() => {
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
  }, [state, rawAppConfig, initialState, updateAppConfig]);

  return (
    <div className={styles.form}>
      <Label className={styles.label}>
        <FormattedMessage id="UserInterfaceScreen.custom-css.label" />
      </Label>
      <CodeEditor
        className={styles.codeEditor}
        language="css"
        value={state.customCss}
        onChange={onCustomCssChange}
      />

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          disabled={!isFormModified}
          onClick={onSaveButtonClicked}
          loading={updatingAppConfig}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>

      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </div>
  );
};

const UserInterfaceScreen: React.FC = function UserInterfaceScreen() {
  const { appID } = useParams();

  const {
    effectiveAppConfig,
    rawAppConfig,
    loading,
    error,
    refetch,
  } = useAppConfigQuery(appID);
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
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingAppConfig,
      })}
    >
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="UserInterfaceScreen.title" />
        </Text>
        <UserInterface
          effectiveAppConfig={effectiveAppConfig}
          rawAppConfig={rawAppConfig}
          updateAppConfig={updateAppConfig}
          updatingAppConfig={updatingAppConfig}
        />
      </div>
    </main>
  );
};

export default UserInterfaceScreen;
