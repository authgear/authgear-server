import { QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";

import { usePortalClient } from "../../portal/apollo";
import { Collaborator, CollaboratorInvitation } from "../globalTypes.generated";
import {
  CollaboratorsAndInvitationsQueryQuery,
  CollaboratorsAndInvitationsQueryQueryVariables,
  CollaboratorsAndInvitationsQueryDocument,
} from "./collaboratorsAndInvitationsQuery.generated";

interface CollaboratorsAndInvitationsQueryResult
  extends Pick<
    QueryResult<
      CollaboratorsAndInvitationsQueryQuery,
      CollaboratorsAndInvitationsQueryQueryVariables
    >,
    "loading" | "error" | "refetch"
  > {
  collaborators: Collaborator[] | null;
  collaboratorInvitations: CollaboratorInvitation[] | null;
}

export function useCollaboratorsAndInvitationsQuery(
  appID: string
): CollaboratorsAndInvitationsQueryResult {
  const client = usePortalClient();
  const { data, loading, error, refetch } =
    useQuery<CollaboratorsAndInvitationsQueryQuery>(
      CollaboratorsAndInvitationsQueryDocument,
      {
        client,
        variables: {
          appID,
        },
      }
    );

  const { collaborators, collaboratorInvitations } = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      collaborators: appNode?.collaborators ?? null,
      collaboratorInvitations: appNode?.collaboratorInvitations ?? null,
    };
  }, [data]);

  return { collaborators, collaboratorInvitations, loading, error, refetch };
}
