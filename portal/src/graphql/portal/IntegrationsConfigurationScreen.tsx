import React, { useContext, useMemo, useCallback } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  DetailsList,
  IColumn,
  ColumnActionsMode,
  SelectionMode,
  Text,
  Image,
  ImageFit,
} from "@fluentui/react";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import Link from "../../Link";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { PortalAPIAppConfig } from "../../types";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./IntegrationsConfigurationScreen.module.css";

import gtmLogoURL from "../../images/gtm_logo.png";

interface FormState {
  googleTagManagerContainerID: string;
}

interface Item {
  iconURL: string;
  name: string;
  description: string;
  connected: boolean;
}

export interface IntegrationsConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    googleTagManagerContainerID: config.google_tag_manager?.container_id ?? "",
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  _currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return config;
}

interface AddonProps {
  item: Item;
}

function Addon(props: AddonProps) {
  const {
    themes: {
      main: {
        palette: { neutralTertiary },
      },
    },
  } = useSystemConfig();
  return (
    <div className={styles.addon}>
      <div className={styles.addonLogo}>
        <Image
          className={styles.addonLogoImage}
          src={props.item.iconURL}
          imageFit={ImageFit.cover}
        />
      </div>
      <Text className={styles.addonName}>{props.item.name}</Text>
      <Text
        className={styles.addonDescription}
        styles={{
          root: {
            color: neutralTertiary,
          },
        }}
      >
        {props.item.description}
      </Text>
    </div>
  );
}

const IntegrationsConfigurationContent: React.VFC<IntegrationsConfigurationContentProps> =
  function IntegrationsConfigurationContent(props) {
    const {
      form: {
        state: { googleTagManagerContainerID },
      },
    } = props;
    const {
      themes: {
        main: {
          palette: { neutralSecondary },
        },
      },
    } = useSystemConfig();

    const { renderToString } = useContext(Context);
    const columns: IColumn[] = [
      {
        key: "add-on",
        name: renderToString("IntegrationsConfigurationScreen.add-on"),
        minWidth: 250,
        columnActionsMode: ColumnActionsMode.disabled,
      },
      {
        key: "status",
        // Empty string here is intentional.
        name: "",
        minWidth: 100,
        columnActionsMode: ColumnActionsMode.disabled,
      },
      {
        key: "action",
        name: renderToString("IntegrationsConfigurationScreen.action"),
        minWidth: 100,
        columnActionsMode: ColumnActionsMode.disabled,
      },
    ];

    const items: Item[] = useMemo(() => {
      return [
        {
          iconURL: gtmLogoURL,
          name: renderToString(
            "IntegrationsConfigurationScreen.add-on.gtm.name"
          ),
          description: renderToString(
            "IntegrationsConfigurationScreen.add-on.gtm.description"
          ),
          connected: googleTagManagerContainerID !== "",
        },
      ];
    }, [renderToString, googleTagManagerContainerID]);

    const onRenderItemColumn = useCallback(
      (item?: Item, _index?: number, column?: IColumn) => {
        if (item == null || column == null) {
          return null;
        }
        switch (column.key) {
          case "add-on": {
            return <Addon item={item} />;
          }
          case "status": {
            if (item.connected) {
              return (
                <div className={styles.cell}>
                  <Text
                    styles={{
                      root: {
                        color: neutralSecondary,
                      },
                    }}
                  >
                    <FormattedMessage id="IntegrationsConfigurationScreen.status.connected" />
                  </Text>
                </div>
              );
            }
            return null;
          }
          case "action": {
            return (
              <div className={styles.cell}>
                <Link to="./google-tag-manager" className={styles.action}>
                  {item.connected ? (
                    <FormattedMessage id="edit" />
                  ) : (
                    <FormattedMessage id="connect" />
                  )}
                </Link>
              </div>
            );
          }
        }
        return null;
      },
      [neutralSecondary]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="IntegrationsConfigurationScreen.title" />
        </ScreenTitle>
        <div className={styles.widget}>
          <DetailsList
            styles={{}}
            columns={columns}
            items={items}
            selectionMode={SelectionMode.none}
            onRenderItemColumn={onRenderItemColumn}
          />
        </div>
      </ScreenContent>
    );
  };

const IntegrationsConfigurationScreen: React.VFC =
  function IntegrationsConfigurationScreen() {
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

    return <IntegrationsConfigurationContent form={form} />;
  };

export default IntegrationsConfigurationScreen;
