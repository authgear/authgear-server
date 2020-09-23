import React from "react";
import { useLocation, useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { useQuery, gql } from "@apollo/client";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import {
  UserDetailsScreenQuery,
  UserDetailsScreenQueryVariables,
  UserDetailsScreenQuery_node_User,
} from "./__generated__/UserDetailsScreenQuery";
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

import styles from "./UserDetailsScreen.module.scss";

interface UserDetailsProps {
  data: UserDetailsScreenQuery_node_User | null;
  loading: boolean;
  refetch: () => void;
}

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { data, loading, refetch } = props;
  const location = useLocation();
  const hash = location.hash.slice(1);
  const { renderToString } = React.useContext(Context);

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
        <Pivot initialSelectedKey={hash}>
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
              refetchUserDetail={refetch}
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
  const { userID } = useParams();
  const { loading, error, data, refetch } = useQuery<
    UserDetailsScreenQuery,
    UserDetailsScreenQueryVariables
  >(query, {
    variables: {
      userID,
    },
    fetchPolicy: "network-only",
  });

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

  return (
    <main className={styles.root}>
      <UserDetailCommandBar />
      <div className={styles.screenContent}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <UserDetails data={userDetails} loading={loading} refetch={refetch} />
      </div>
    </main>
  );
};

export default UserDetailsScreen;
