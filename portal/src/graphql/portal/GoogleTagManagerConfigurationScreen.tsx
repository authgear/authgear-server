import React, { useMemo, useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { produce } from "immer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb from "../../NavBreadcrumb";
import { PortalAPIAppConfig } from "../../types";
import Widget from "../../Widget";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import Toggle from "../../Toggle";
import styles from "./GoogleTagManagerConfigurationScreen.module.css";
import { clearEmptyObject } from "../../util/misc";

interface FormState {
  enabled: boolean;
  googleTagManagerContainerID: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const googleTagManagerContainerID =
    config.google_tag_manager?.container_id ?? "";
  const enabled = googleTagManagerContainerID !== "";
  return {
    enabled,
    googleTagManagerContainerID,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.google_tag_manager ??= {};
    if (currentState.enabled) {
      config.google_tag_manager.container_id =
        currentState.googleTagManagerContainerID;
    } else {
      delete config.google_tag_manager.container_id;
    }
    clearEmptyObject(config);
  });
}

interface GoogleTagManagerConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const GoogleTagManagerConfigurationContent: React.VFC<GoogleTagManagerConfigurationContentProps> =
  function GoogleTagManagerConfigurationContent(props) {
    const { renderToString } = useContext(Context);
    const {
      form: {
        state: { googleTagManagerContainerID, enabled },
        setState,
      },
    } = props;

    const navBreadcrumbItems = useMemo(() => {
      return [
        {
          to: "~/integrations",
          label: (
            <FormattedMessage id="IntegrationsConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: (
            <FormattedMessage id="GoogleTagManagerConfigurationScreen.title" />
          ),
        },
      ];
    }, []);

    const onChangeEnabled = useCallback(
      (_e, checked?: boolean) => {
        if (checked == null) {
          return;
        }
        setState((prev) => {
          return {
            ...prev,
            enabled: checked,
          };
        });
      },
      [setState]
    );

    const onChangeContainerID = useCallback(
      (_e, newValue?: string) => {
        if (newValue == null) {
          return;
        }

        setState((prev) => {
          return {
            ...prev,
            googleTagManagerContainerID: newValue,
          };
        });
      },
      [setState]
    );

    return (
      <ScreenContent>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <Widget className={styles.widget}>
          <WidgetDescription>
            <FormattedMessage id="GoogleTagManagerConfigurationScreen.description" />
          </WidgetDescription>
        </Widget>
        <Widget className={styles.widget}>
          <Toggle
            className={styles.control}
            checked={enabled}
            label={
              <FormattedMessage id="GoogleTagManagerConfigurationScreen.toggle.label" />
            }
            inlineLabel={true}
            onChange={onChangeEnabled}
          />
          <FormTextField
            className={styles.control}
            parentJSONPointer="/google_tag_manager"
            fieldName="container_id"
            value={googleTagManagerContainerID}
            onChange={onChangeContainerID}
            required={true}
            label={renderToString(
              "GoogleTagManagerConfigurationScreen.container-id.label"
            )}
            placeholder={renderToString(
              "GoogleTagManagerConfigurationScreen.container-id.placeholder"
            )}
            disabled={!enabled}
          />
        </Widget>
      </ScreenContent>
    );
  };

const GoogleTagManagerConfigurationScreen: React.VFC =
  function GoogleTagManagerConfigurationScreen() {
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
        <GoogleTagManagerConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default GoogleTagManagerConfigurationScreen;
