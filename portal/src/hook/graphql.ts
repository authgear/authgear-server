import {
  ApolloError,
  DocumentNode,
  MutationHookOptions,
  MutationTuple,
  OperationVariables,
  useMutation,
} from "@apollo/client";
import { useCallback, useEffect, useState } from "react";

export function useGraphqlMutation<
  TData = never,
  TVariables = OperationVariables
>(
  mutation: DocumentNode,
  options?: MutationHookOptions<TData, TVariables>
): [...MutationTuple<TData, TVariables>, () => void] {
  const [mutationFunction, mutationResult] = useMutation<TData, TVariables>(
    mutation,
    options
  );
  const { error: errorResult } = mutationResult;

  const [error, setError] = useState<ApolloError | undefined>();

  const resetError = useCallback(() => {
    setError(undefined);
  }, []);

  useEffect(() => {
    setError(errorResult);
  }, [errorResult]);

  return [mutationFunction, { ...mutationResult, error }, resetError];
}
