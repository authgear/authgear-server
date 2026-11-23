import * as Apollo from "@apollo/client";
import {
  AuthFlowsScreenQueryQuery,
  AuthFlowsScreenQueryQueryVariables,
  AuthFlowsScreenQueryDocument,
  AuthFlowsScreenQueryQueryHookResult,
} from "./authFlowsScreenQuery.generated";
import { usePortalClient } from "../apollo";

export function useAuthFlowsScreenQueryQuery(
  baseOptions?: Apollo.QueryHookOptions<
    AuthFlowsScreenQueryQuery,
    AuthFlowsScreenQueryQueryVariables
  >
): AuthFlowsScreenQueryQueryHookResult {
  const client = usePortalClient();
  const options = { ...{ client }, ...baseOptions };
  return Apollo.useQuery<
    AuthFlowsScreenQueryQuery,
    AuthFlowsScreenQueryQueryVariables
  >(AuthFlowsScreenQueryDocument, options);
}
