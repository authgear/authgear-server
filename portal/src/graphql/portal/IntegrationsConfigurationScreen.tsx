import React, { useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { Text } from "@radix-ui/themes";
import { FormattedMessage, Context } from "../../intl";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import Link from "../../Link";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import { Badge } from "../../components/v2/Badge/Badge";
import { PortalAPIAppConfig } from "../../types";
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
  editPath: string;
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
  const { item } = props;
  return (
    <div className={styles.addon}>
      <div className={styles.addonLogo}>
        <img className={styles.addonLogoImage} src={item.iconURL} alt="" />
      </div>
      <Text as="div" size="2" weight="medium" className={styles.addonName}>
        {item.name}
      </Text>
      <Text as="div" size="2" className={styles.addonDescription}>
        {item.description}
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

    const { renderToString } = useContext(Context);

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
          editPath: "./google-tag-manager",
        },
      ];
    }, [renderToString, googleTagManagerContainerID]);

    return (
      <ScreenLayoutScrollView>
        <ScreenContent layout="list">
          <div className={styles.widget}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="IntegrationsConfigurationScreen.title" />
            </Text>
          </div>
          <div className={styles.widget}>
            <div className={styles.tableWrapper}>
              <div className={styles.table}>
                <div className={styles.tableHeader}>
                  <div className={styles.headerCellAddon}>
                    <FormattedMessage id="IntegrationsConfigurationScreen.add-on" />
                  </div>
                  <div className={styles.headerCellStatus} aria-hidden={true} />
                  <div className={styles.headerCellAction}>
                    <FormattedMessage id="IntegrationsConfigurationScreen.action" />
                  </div>
                </div>
                {items.map((item) => (
                  <div key={item.name} className={styles.tableRow}>
                    <div className={styles.cellAddon}>
                      <Addon item={item} />
                    </div>
                    <div className={styles.cellStatus}>
                      {item.connected ? (
                        <Badge
                          size="1"
                          variant="success"
                          text={
                            <FormattedMessage id="IntegrationsConfigurationScreen.status.connected" />
                          }
                        />
                      ) : null}
                    </div>
                    <div className={styles.cellAction}>
                      <Link to={item.editPath} className={styles.action}>
                        {item.connected ? (
                          <FormattedMessage id="edit" />
                        ) : (
                          <FormattedMessage id="connect" />
                        )}
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </ScreenContent>
      </ScreenLayoutScrollView>
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
