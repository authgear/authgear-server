import { useCallback, useState, useEffect } from "react";
import { useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import {
  GenerateAppSecretVisitTokenMutationDocument,
  GenerateAppSecretVisitTokenMutationMutation,
  GenerateAppSecretVisitTokenMutationMutationVariables,
} from "./generateAppSecretVisitTokenMutation.generated";
import { AppSecretKey } from "../globalTypes.generated";

export function useGenerateAppSecretVisitTokenMutation(appID: string): {
  generateAppSecretVisitToken: (
    secrets: AppSecretKey[]
  ) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    GenerateAppSecretVisitTokenMutationMutation,
    GenerateAppSecretVisitTokenMutationMutationVariables
  >(GenerateAppSecretVisitTokenMutationDocument, {
    client,
  });

  const generateAppSecretVisitToken = useCallback(
    async (secrets: AppSecretKey[]) => {
      const result = await mutationFunction({
        variables: { appID, secrets },
      });

      return result.data?.generateAppSecretVisitToken.token ?? null;
    },
    [mutationFunction, appID]
  );

  return { generateAppSecretVisitToken, error, loading };
}

export function useAppSecretVisitToken(
  appID: string,
  secrets: AppSecretKey[]
): {
  token: string | null | undefined;
  loading: boolean;
  error: unknown;
  retry: () => void;
} {
  const [token, setToken] = useState<string | null | undefined>(undefined);
  const [retryCounter, setRetryCounter] = useState(0);
  const { generateAppSecretVisitToken, loading, error } =
    useGenerateAppSecretVisitTokenMutation(appID);

  useEffect(() => {
    let isMounted = true;
    const cleanupFn = () => {
      isMounted = false;
    };
    if (secrets.length === 0) {
      setToken(null);
      return cleanupFn;
    }
    setToken(undefined);
    generateAppSecretVisitToken(secrets)
      .then((token) => {
        if (isMounted) {
          setToken(token);
        }
      })
      .catch((e) => console.error("Failed to generate secret visit token", e));

    return cleanupFn;
  }, [generateAppSecretVisitToken, secrets, retryCounter]);

  const retry = useCallback(() => {
    setRetryCounter((prev) => (prev + 1) % 100000);
  }, []);

  return {
    token,
    loading,
    error,
    retry,
  };
}
