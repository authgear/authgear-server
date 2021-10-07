import React, { useMemo, useState, useCallback } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import UserDetailCommandBarContainer from "./UserDetailCommandBarContainer";
import UserDetailSummary from "./UserDetailSummary";
import UserDetailsStandardAttributes from "./UserDetailsStandardAttributes";
import UserDetailsAccountSecurity from "./UserDetailsAccountSecurity";
import UserDetailsConnectedIdentities from "./UserDetailsConnectedIdentities";
import UserDetailsSession from "./UserDetailsSession";

import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { usePivotNavigation } from "../../hook/usePivot";
import { nonNullable } from "../../util/types";
import { extractUserInfoFromIdentities } from "../../util/user";
import { PortalAPIAppConfig, StandardAttributes } from "../../types";

import styles from "./UserDetailsScreen.module.scss";

interface UserDetailsProps {
  data: UserQuery_node_User | null;
  appConfig: PortalAPIAppConfig | null;
}

const STANDARD_ATTRIBUTES_KEY = "standard-attributes";
const ACCOUNT_SECURITY_PIVOT_KEY = "account-security";
const CONNECTED_IDENTITIES_PIVOT_KEY = "connected-identities";
const SESSION_PIVOT_KEY = "session";

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { data, appConfig } = props;
  const { renderToString } = React.useContext(Context);
  const { selectedKey, onLinkClick } = usePivotNavigation([
    STANDARD_ATTRIBUTES_KEY,
    ACCOUNT_SECURITY_PIVOT_KEY,
    CONNECTED_IDENTITIES_PIVOT_KEY,
    SESSION_PIVOT_KEY,
  ]);

  const availableLoginIdIdentities = useMemo(() => {
    const authenticationIdentities =
      appConfig?.authentication?.identities ?? [];
    const loginIdIdentityEnabled =
      authenticationIdentities.includes("login_id");
    if (!loginIdIdentityEnabled) {
      return [];
    }
    const rawLoginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
    return rawLoginIdKeys.map((loginIdKey) => loginIdKey.key);
  }, [appConfig]);

  const [standardAttributes, setStandardAttributes] = useState(
    data?.standardAttributes ?? {}
  );

  const verifiedClaims = data?.verifiedClaims ?? [];

  const identities =
    data?.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
    [];
  const userInfo = extractUserInfoFromIdentities(identities);

  const authenticators =
    data?.authenticators?.edges
      ?.map((edge) => edge?.node)
      .filter(nonNullable) ?? [];

  const sessions =
    data?.sessions?.edges?.map((edge) => edge?.node).filter(nonNullable) ?? [];

  const onChangeStandardAttributes = useCallback(
    (attrs: StandardAttributes) => {
      setStandardAttributes(attrs);
    },
    []
  );

  return (
    <div className={styles.userDetails}>
      <UserDetailSummary
        userInfo={userInfo}
        createdAtISO={data?.createdAt ?? null}
        lastLoginAtISO={data?.lastLoginAt ?? null}
      />
      <div className={styles.userDetailsTab}>
        <Pivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
          <PivotItem
            itemKey={STANDARD_ATTRIBUTES_KEY}
            headerText={renderToString(
              "UserDetails.standard-attributes.header"
            )}
          >
            <UserDetailsStandardAttributes
              identities={identities}
              standardAttributes={standardAttributes}
              onChangeStandardAttributes={onChangeStandardAttributes}
            />
          </PivotItem>
          <PivotItem
            itemKey={ACCOUNT_SECURITY_PIVOT_KEY}
            headerText={renderToString("UserDetails.account-security.header")}
          >
            <UserDetailsAccountSecurity authenticators={authenticators} />
          </PivotItem>
          <PivotItem
            itemKey={CONNECTED_IDENTITIES_PIVOT_KEY}
            headerText={renderToString(
              "UserDetails.connected-identities.header"
            )}
          >
            <UserDetailsConnectedIdentities
              identities={identities}
              verifiedClaims={verifiedClaims}
              availableLoginIdIdentities={availableLoginIdIdentities}
            />
          </PivotItem>
          <PivotItem
            itemKey={SESSION_PIVOT_KEY}
            headerText={renderToString("UserDetails.session.header")}
          >
            <UserDetailsSession sessions={sessions} />
          </PivotItem>
        </Pivot>
      </div>
    </div>
  );
};

const UserDetailsScreen: React.FC = function UserDetailsScreen() {
  const { appID, userID } = useParams();
  const { user, loading: loadingUser, error, refetch } = useUserQuery(userID);
  const {
    effectiveAppConfig,
    loading: loadingAppConfig,
    error: appConfigError,
    refetch: refetchAppConfig,
  } = useAppAndSecretConfigQuery(appID);

  const navBreadcrumbItems = React.useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="UserDetailsScreen.title" /> },
    ];
  }, []);

  const loading = loadingUser || loadingAppConfig;

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (appConfigError != null) {
    return <ShowError error={appConfigError} onRetry={refetchAppConfig} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  const identities =
    user?.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
    [];

  return (
    <UserDetailCommandBarContainer user={user} identities={identities}>
      <main className={styles.root}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <UserDetails data={user} appConfig={effectiveAppConfig} />
      </main>
    </UserDetailCommandBarContainer>
  );
};

export default UserDetailsScreen;
