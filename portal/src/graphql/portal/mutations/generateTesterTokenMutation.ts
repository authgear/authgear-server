import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import { usePortalClient } from "../../portal/apollo";
import {
  GenerateTesterTokenMutationDocument,
  GenerateTesterTokenMutationMutation,
  GenerateTesterTokenMutationMutationVariables,
} from "./generateTesterTokenMutation.generated";

export function useGenerateTesterTokenMutation(appID: string): {
  generateTesterToken: (returnUri: string) => Promise<string>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    GenerateTesterTokenMutationMutation,
    GenerateTesterTokenMutationMutationVariables
  >(GenerateTesterTokenMutationDocument, {
    client,
  });

  const generateTesterToken = useCallback(
    async (returnUri: string) => {
      const result = await mutationFunction({
        variables: { appID, returnUri },
      });

      if (result.errors != null && result.errors.length > 0) {
        // eslint-disable-next-line @typescript-eslint/only-throw-error
        throw result.errors;
      }

      return result.data!.generateTesterToken.token;
    },
    [mutationFunction, appID]
  );

  return { generateTesterToken, error, loading };
}
