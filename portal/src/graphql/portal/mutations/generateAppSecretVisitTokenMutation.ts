import { useCallback, useState, useEffect, useMemo } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
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
  const client = usePortalClient();
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

class SecretVisitTokenStore {
  private store: Storage;
  private appID: string;
  constructor(store: Storage, appID: string) {
    this.store = store;
    this.appID = appID;
  }

  private formatTokenKey(secrets: AppSecretKey[]): string {
    return `app.secrettoken.${this.appID}.${secrets.sort(undefined).join("+")}`;
  }

  public get(secrets: AppSecretKey[]): string | null {
    const key = this.formatTokenKey(secrets);
    return this.store.getItem(key);
  }

  public set(secrets: AppSecretKey[], token: string): void {
    const key = this.formatTokenKey(secrets);
    this.store.setItem(key, token);
  }
}

export function useAppSecretVisitToken(
  appID: string,
  secrets: AppSecretKey[],
  refresh: boolean
): {
  token: string | null | undefined;
  loading: boolean;
  error: unknown;
  retry: () => void;
} {
  const store = useMemo(
    () => new SecretVisitTokenStore(window.sessionStorage, appID),
    [appID]
  );

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
    const cachedToken = store.get(secrets);
    if (cachedToken && !refresh) {
      setToken(cachedToken);
      return cleanupFn;
    }
    if (!refresh) {
      setToken(null);
      return cleanupFn;
    }
    generateAppSecretVisitToken(secrets)
      .then((token) => {
        if (!isMounted) {
          return;
        }
        if (token != null) {
          store.set(secrets, token);
        }
        setToken(token);
      })
      .catch((e) => {
        if (!isMounted) {
          return;
        }
        console.error("Failed to generate secret visit token", e);
      });

    return cleanupFn;
  }, [generateAppSecretVisitToken, secrets, retryCounter, store, refresh]);

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
