import React from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  useAppAndSecretConfigQuery,
  AppAndSecretConfigQueryResult,
} from "./query/appAndSecretConfigQuery";
import styles from "./AdminAPIConfigurationScreen.module.scss";

interface AdminAPIConfigurationScreenContentProps {
  queryResult: AppAndSecretConfigQueryResult;
}

const AdminAPIConfigurationScreenContent: React.FC<AdminAPIConfigurationScreenContentProps> =
  function AdminAPIConfigurationScreenContent(props) {
    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="AdminAPIConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="AdminAPIConfigurationScreen.description" />
        </ScreenDescription>
      </ScreenContent>
    );
  };

const AdminAPIConfigurationScreen: React.FC =
  function AdminAPIConfigurationScreen() {
    const { appID } = useParams();
    const queryResult = useAppAndSecretConfigQuery(appID);

    if (queryResult.loading) {
      return <ShowLoading />;
    }

    if (queryResult.error) {
      return (
        <ShowError error={queryResult.error} onRetry={queryResult.refetch} />
      );
    }

    return <AdminAPIConfigurationScreenContent queryResult={queryResult} />;
  };

export default AdminAPIConfigurationScreen;
