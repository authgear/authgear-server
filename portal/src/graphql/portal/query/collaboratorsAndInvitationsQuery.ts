import { gql, QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";

import { client } from "../../portal/apollo";
import {
  CollaboratorsAndInvitationsQuery,
  CollaboratorsAndInvitationsQueryVariables,
  CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations,
  CollaboratorsAndInvitationsQuery_node_App_collaborators,
} from "./__generated__/CollaboratorsAndInvitationsQuery";

export const collaboratorsAndInvitationsQuery = gql`
  query CollaboratorsAndInvitationsQuery($appID: ID!) {
    node(id: $appID) {
      __typename
      ... on App {
        id
        collaborators {
          id
          createdAt
          user {
            id
            email
          }
        }
        collaboratorInvitations {
          id
          createdAt
          expireAt
          invitedBy {
            id
            email
          }
          inviteeEmail
        }
      }
    }
  }
`;

export type Collaborator = CollaboratorsAndInvitationsQuery_node_App_collaborators;
export type CollaboratorInvitation = CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations;

interface CollaboratorsAndInvitationsQueryResult
  extends Pick<
    QueryResult<
      CollaboratorsAndInvitationsQuery,
      CollaboratorsAndInvitationsQueryVariables
    >,
    "loading" | "error" | "refetch"
  > {
  collaborators: Collaborator[] | null;
  collaboratorInvitations: CollaboratorInvitation[] | null;
}

export function useCollaboratorsAndInvitationsQuery(
  appID: string
): CollaboratorsAndInvitationsQueryResult {
  const { data, loading, error, refetch } = useQuery<
    CollaboratorsAndInvitationsQuery
  >(collaboratorsAndInvitationsQuery, {
    client,
    variables: {
      appID,
    },
  });

  const { collaborators, collaboratorInvitations } = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      collaborators: appNode?.collaborators ?? null,
      collaboratorInvitations: appNode?.collaboratorInvitations ?? null,
    };
  }, [data]);

  return { collaborators, collaboratorInvitations, loading, error, refetch };
}
