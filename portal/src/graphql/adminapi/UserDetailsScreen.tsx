import React, { useMemo } from "react";
import { useLocation, useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import UserDetailCommandBar from "./UserDetailCommandBar";
import UserDetailSummary from "./UserDetailSummary";
import UserDetailsAccountSecurity from "./UserDetailsAccountSecurity";
import UserDetailsConnectedIdentities from "./UserDetailsConnectedIdentities";
import UserDetailsSession from "./UserDetailsSession";

import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { nonNullable } from "../../util/types";
import { extractUserInfoFromIdentities } from "../../util/user";
import { PortalAPIAppConfig } from "../../types";

import styles from "./UserDetailsScreen.module.scss";

interface UserDetailsProps {
  data: UserQuery_node_User | null;
  appConfig: PortalAPIAppConfig | null;
  loading: boolean;
}

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { data, loading, appConfig } = props;
  const location = useLocation();
  const hash = location.hash.slice(1);
  const { renderToString } = React.useContext(Context);

  const availableLoginIdIdentities = useMemo(() => {
    const authenticationIdentities =
      appConfig?.authentication?.identities ?? [];
    const loginIdIdentityEnabled = authenticationIdentities.includes(
      "login_id"
    );
    if (!loginIdIdentityEnabled) {
      return [];
    }
    const rawLoginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
    return rawLoginIdKeys.map((loginIdKey) => loginIdKey.key);
  }, [appConfig]);

  if (loading) {
    return <ShowLoading />;
  }

  const identities =
    data?.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
    [];
  const userInfo = extractUserInfoFromIdentities(identities);

  const authenticators =
    data?.authenticators?.edges
      ?.map((edge) => edge?.node)
      .filter(nonNullable) ?? [];

  return (
    <div className={styles.userDetails}>
      <UserDetailSummary
        userInfo={userInfo}
        createdAtISO={data?.createdAt ?? null}
        lastLoginAtISO={data?.lastLoginAt ?? null}
      />
      <div className={styles.userDetailsTab}>
        <Pivot defaultSelectedKey={hash}>
          <PivotItem
            itemKey={"account-security"}
            headerText={renderToString("UserDetails.account-security.header")}
          >
            <UserDetailsAccountSecurity authenticators={authenticators} />
          </PivotItem>
          <PivotItem
            itemKey={"connected-identities"}
            headerText={renderToString(
              "UserDetails.connected-identities.header"
            )}
          >
            <UserDetailsConnectedIdentities
              identities={identities}
              availableLoginIdIdentities={availableLoginIdIdentities}
            />
          </PivotItem>
          <PivotItem
            itemKey={"session"}
            headerText={renderToString("UserDetails.session.header")}
          >
            <UserDetailsSession />
          </PivotItem>
        </Pivot>
      </div>
    </div>
  );
};

const UserDetailsScreen: React.FC = function UserDetailsScreen() {
  const { appID, userID } = useParams();
  const { user, loading, error, refetch } = useUserQuery(userID);
  const {
    loading: loadingAppConfig,
    error: appConfigError,
    data: appConfigData,
    refetch: refetchAppConfig,
  } = useAppConfigQuery(appID);

  const appConfig =
    appConfigData?.node?.__typename === "App"
      ? appConfigData.node.effectiveAppConfig
      : null;

  const navBreadcrumbItems = React.useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="UserDetailsScreen.title" /> },
    ];
  }, []);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (appConfigError != null) {
    return <ShowError error={error} onRetry={refetchAppConfig} />;
  }

  return (
    <main className={styles.root}>
      <UserDetailCommandBar />
      <div className={styles.screenContent}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <UserDetails
          data={user}
          loading={loading || loadingAppConfig}
          appConfig={appConfig}
        />
      </div>
    </main>
  );
};

export default UserDetailsScreen;
