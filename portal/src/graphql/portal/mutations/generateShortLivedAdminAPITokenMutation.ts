import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import { usePortalClient } from "../../portal/apollo";
import {
  GenerateShortLivedAdminApiTokenDocument,
  GenerateShortLivedAdminApiTokenMutation,
  GenerateShortLivedAdminApiTokenMutationVariables,
} from "./generateShortLivedAdminAPITokenMutation.generated";

export function useGenerateShortLivedAdminAPITokenMutation(appID: string): {
  generateShortLivedAdminAPIToken: (
    appSecretVisitToken: string
  ) => Promise<string>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    GenerateShortLivedAdminApiTokenMutation,
    GenerateShortLivedAdminApiTokenMutationVariables
  >(GenerateShortLivedAdminApiTokenDocument, {
    client,
  });

  const generateShortLivedAdminAPIToken = useCallback(
    async (appSecretVisitToken: string) => {
      const result = await mutationFunction({
        variables: { appID, appSecretVisitToken },
      });

      if (result.errors != null && result.errors.length > 0) {
        // eslint-disable-next-line @typescript-eslint/only-throw-error
        throw result.errors;
      }

      return result.data!.generateShortLivedAdminAPIToken!.token;
    },
    [mutationFunction, appID]
  );

  return { generateShortLivedAdminAPIToken, error, loading };
}
