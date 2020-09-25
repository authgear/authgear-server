import React, { useMemo } from "react";
import { useLocation, useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { useQuery, gql } from "@apollo/client";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import {
  UserDetailsScreenQuery,
  UserDetailsScreenQueryVariables,
  UserDetailsScreenQuery_node_User,
} from "./__generated__/UserDetailsScreenQuery";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import UserDetailCommandBar from "./UserDetailCommandBar";
import UserDetailSummary from "./UserDetailSummary";
import UserDetailsAccountSecurity from "./UserDetailsAccountSecurity";
import UserDetailsConnectedIdentities from "./UserDetailsConnectedIdentities";
import UserDetailsSession from "./UserDetailsSession";

import { nonNullable } from "../../util/types";
import { extractUserInfoFromIdentities } from "../../util/user";
import { PortalAPIAppConfig } from "../../types";

import styles from "./UserDetailsScreen.module.scss";

interface UserDetailsProps {
  data: UserDetailsScreenQuery_node_User | null;
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

const query = gql`
  query UserDetailsScreenQuery($userID: ID!) {
    node(id: $userID) {
      __typename
      ... on User {
        id
        authenticators {
          edges {
            node {
              id
              type
              kind
              isDefault
              claims
              createdAt
              updatedAt
            }
          }
        }
        identities {
          edges {
            node {
              id
              type
              claims
              createdAt
              updatedAt
            }
          }
        }
        lastLoginAt
        createdAt
        updatedAt
      }
    }
  }
`;

const UserDetailsScreen: React.FC = function UserDetailsScreen() {
  const { appID, userID } = useParams();
  const { loading, error, data, refetch } = useQuery<
    UserDetailsScreenQuery,
    UserDetailsScreenQueryVariables
  >(query, {
    variables: {
      userID,
    },
    fetchPolicy: "network-only",
  });
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

  const userDetails = React.useMemo(() => {
    const node = data?.node;
    return node?.__typename === "User" ? node : null;
  }, [data]);

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
          data={userDetails}
          loading={loading || loadingAppConfig}
          appConfig={appConfig}
        />
      </div>
    </main>
  );
};

export default UserDetailsScreen;
