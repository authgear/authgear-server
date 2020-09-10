import React, { useContext } from "react";
import { useParams } from "react-router-dom";
import {
  Pivot,
  PivotItem,
  CommandBar,
  ICommandBarItemProps,
} from "@fluentui/react";
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
}

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { data, loading } = props;
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
        <Pivot>
          <PivotItem
            headerText={renderToString("UserDetails.account-security.header")}
          >
            <UserDetailsAccountSecurity authenticators={authenticators} />
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "UserDetails.connected-identities.header"
            )}
          >
            <UserDetailsConnectedIdentities identities={identities} />
          </PivotItem>
          <PivotItem headerText={renderToString("UserDetails.session.header")}>
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
  const { renderToString } = useContext(Context);
  const { loading, error, data, refetch } = useQuery<
    UserDetailsScreenQuery,
    UserDetailsScreenQueryVariables
  >(query, {
    variables: {
      userID,
    },
  });

  const navBreadcrumItems = React.useMemo(() => {
    return [
      { to: "../../", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="UserDetailsScreen.title" /> },
    ];
  }, []);

  const userDetails = React.useMemo(() => {
    const node = data?.node;
    return node?.__typename === "User" ? node : null;
  }, [data]);

  const commandBarItems: ICommandBarItemProps[] = [
    {
      key: "remove",
      text: renderToString("remove"),
      iconProps: { iconName: "Delete" },
    },

    {
      key: "loginAsUser",
      text: renderToString("UserDetails.command-bar.login-as-user"),
      iconProps: { iconName: "FollowUser" },
    },

    {
      key: "invalidateSessions",
      text: renderToString("UserDetails.command-bar.invalidate-sessions"),
      iconProps: {
        iconName: "CircleAddition",
        className: styles.invalidateIcon,
      },
    },
    {
      key: "disable",
      text: renderToString("disable"),
      iconProps: { iconName: "CircleStop" },
    },
  ];

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root} role="main">
      <CommandBar
        className={styles.commandBar}
        items={[]}
        farItems={commandBarItems}
      />
      <div className={styles.screenContent}>
        <NavBreadcrumb items={navBreadcrumItems} />
        <UserDetails data={userDetails} loading={loading} />
      </div>
    </main>
  );
};

export default UserDetailsScreen;
