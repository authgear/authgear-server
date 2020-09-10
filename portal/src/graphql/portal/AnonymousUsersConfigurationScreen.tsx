import React, { useContext, useCallback } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Text, Toggle, Dropdown, PrimaryButton } from "@fluentui/react";

import { useAppConfigQuery } from "./query/appConfigQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

import styles from "./AnonymousUsersConfigurationScreen.module.scss";

const AnonymousUserConfigurationScreen: React.FC = function AnonymousUserConfigurationScreen() {
  const { appID } = useParams();
  const { loading, error, data, refetch } = useAppConfigQuery(appID);
  const { renderToString } = useContext(Context);

  const appConfig =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;

  const onSaveClicked = useCallback(() => {
    produce(appConfig, (_draftConfig) => {
      // TODO: to be implemented
    });
    // TODO: call mutation to save config
  }, [appConfig]);

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
          options={[]}
        />
        <PrimaryButton className={styles.saveButton} onClick={onSaveClicked}>
          <FormattedMessage id="save" />
        </PrimaryButton>
      </section>
    </main>
  );
};

export default AnonymousUserConfigurationScreen;
