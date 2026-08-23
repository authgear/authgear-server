import React, { useCallback, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { Text } from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import styles from "./DynamicClientsEmptyView.module.css";

export interface DynamicClientsEmptyViewProps {
  // Whether oauth.dynamic_client_registration.enabled is on for the project.
  registrationEnabled: boolean;
}

export const DynamicClientsEmptyView: React.VFC<DynamicClientsEmptyViewProps> =
  function DynamicClientsEmptyView({ registrationEnabled }) {
    const navigate = useNavigate();
    const { renderToString } = useContext(Context);

    const onEnableClick = useCallback(() => {
      navigate("./dcr");
    }, [navigate]);

    if (!registrationEnabled) {
      return (
        <div className={styles.container}>
          <Text
            variant="mediumPlus"
            className={styles.title}
            block={true}
            styles={{ root: { fontWeight: 600, color: "var(--gray-12)" } }}
          >
            <FormattedMessage id="DynamicClientsEmptyView.disabled.title" />
          </Text>
          <Text
            variant="medium"
            className={styles.description}
            block={true}
            styles={{ root: { color: "var(--gray-11)" } }}
          >
            <FormattedMessage id="DynamicClientsEmptyView.disabled.description" />
          </Text>
          <PrimaryButton
            text={renderToString("DynamicClientsEmptyView.disabled.cta")}
            onClick={onEnableClick}
          />
        </div>
      );
    }

    return (
      <div className={styles.container}>
        <Text
          variant="mediumPlus"
          className={styles.title}
          block={true}
          styles={{ root: { fontWeight: 600, color: "var(--gray-12)" } }}
        >
          <FormattedMessage id="DynamicClientsEmptyView.empty.title" />
        </Text>
        <Text
          variant="medium"
          className={styles.description}
          block={true}
          styles={{ root: { color: "var(--gray-11)" } }}
        >
          <FormattedMessage id="DynamicClientsEmptyView.empty.description" />
        </Text>
      </div>
    );
  };
