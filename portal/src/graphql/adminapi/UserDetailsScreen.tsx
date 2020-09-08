import React from "react";
import { useParams } from "react-router-dom";
import { useQuery, gql } from "@apollo/client";
import { FormattedMessage } from "@oursky/react-messageformat";

import {
  UserDetailsScreenQuery,
  UserDetailsScreenQueryVariables,
  UserDetailsScreenQuery_node_User,
} from "./__generated__/UserDetailsScreenQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { isUserDetails } from "../../util/types";

import styles from "./UserDetailsScreen.module.scss";
interface UserDetailsProps {
  data: UserDetailsScreenQuery_node_User | null;
  loading: boolean;
}

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { data, loading } = props;
  if (loading) {
    return <ShowLoading />;
  }
  return <div className={styles.root}></div>;
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
    return isUserDetails(node) ? node : null;
  }, [data]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumItems} />
      <UserDetails data={userDetails} loading={loading} />
    </div>
  );
};

export default UserDetailsScreen;
