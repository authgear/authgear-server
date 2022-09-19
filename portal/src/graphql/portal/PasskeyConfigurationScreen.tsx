import React, { useCallback, useContext, useMemo } from "react";
import { TooltipHost } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import Toggle from "../../Toggle";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useTooltipTargetElement } from "../../Tooltip";
import FormContainer from "../../FormContainer";
import styles from "./PasskeyConfigurationScreen.module.css";

interface FormState {
  passkeyChecked: boolean;
  passkeyDisabled: boolean;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const passkeyIndex =
    config.authentication?.primary_authenticators?.indexOf("passkey");
  const passkeyChecked = passkeyIndex != null && passkeyIndex >= 0;

  const passkeyDisabled = !(
    config.authentication?.identities?.includes("login_id") ?? true
  );

  return {
    passkeyChecked,
    passkeyDisabled,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    function setEnable<T extends string>(
      arr: T[],
      value: T,
      enabled: boolean
    ): T[] {
      const index = arr.indexOf(value);

      if (enabled) {
        if (index >= 0) {
          return arr;
        }
        return [...arr, value];
      }

      if (index < 0) {
        return arr;
      }
      return [...arr.slice(0, index), ...arr.slice(index + 1)];
    }

    // Construct primary_authenticators and identities
    // We do not offer any control to modify identities o primary_authenticators  this screen,
    // so we read from effectiveConfig
    config.authentication ??= {};
    if (currentState.passkeyChecked) {
      config.authentication.primary_authenticators = setEnable(
        effectiveConfig.authentication?.primary_authenticators ?? [],
        "passkey",
        true
      );
      config.authentication.identities = setEnable(
        effectiveConfig.authentication?.identities ?? [],
        "passkey",
        true
      );
    } else {
      config.authentication.primary_authenticators = setEnable(
        effectiveConfig.authentication?.primary_authenticators ?? [],
        "passkey",
        false
      );
      config.authentication.identities = setEnable(
        effectiveConfig.authentication?.identities ?? [],
        "passkey",
        false
      );
    }
    clearEmptyObject(config);
  });
}

interface PasskeyConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const PasskeyConfigurationContent: React.VFC<PasskeyConfigurationContentProps> =
  function PasskeyConfigurationContent(props) {
    const { state, setState } = props.form;

    const tooltipResult = useTooltipTargetElement();
    const passkeyTooltipProps = useMemo(() => {
      return {
        targetElement: tooltipResult.targetElement,
      };
    }, [tooltipResult.targetElement]);

    const { passkeyChecked, passkeyDisabled } = state;

    const { renderToString } = useContext(Context);

    const onChangePasskeyChecked = useCallback(
      (_event, checked?: boolean) =>
        setState((state) => ({
          ...state,
          passkeyChecked: checked ?? false,
        })),
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="PasskeyConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="PasskeyConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          {passkeyDisabled ? (
            <TooltipHost
              content={<FormattedMessage id="errors.validation.passkey" />}
              tooltipProps={passkeyTooltipProps}
            >
              <Toggle
                id={tooltipResult.id}
                ref={tooltipResult.setRef}
                label={renderToString(
                  "PasskeyConfigurationScreen.toggle.title"
                )}
                description={renderToString(
                  "PasskeyConfigurationScreen.toggle.description"
                )}
                disabled={passkeyDisabled}
                checked={passkeyChecked}
                onChange={onChangePasskeyChecked}
                inlineLabel={false}
              />
            </TooltipHost>
          ) : (
            <div>
              <Toggle
                label={renderToString(
                  "PasskeyConfigurationScreen.toggle.title"
                )}
                description={renderToString(
                  "PasskeyConfigurationScreen.toggle.description"
                )}
                disabled={passkeyDisabled}
                checked={passkeyChecked}
                onChange={onChangePasskeyChecked}
                inlineLabel={false}
              />
            </div>
          )}
        </Widget>
      </ScreenContent>
    );
  };

const PasskeyConfigurationScreen: React.VFC =
  function PasskeyConfigurationScreen() {
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
      <FormContainer form={form}>
        <PasskeyConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default PasskeyConfigurationScreen;
