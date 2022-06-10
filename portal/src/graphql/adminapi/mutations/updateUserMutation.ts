import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { StandardAttributes, CustomAttributes } from "../../../types";
import {
  UpdateUserMutationMutation,
  UpdateUserMutationDocument,
} from "./updateUserMutation.generated";

export interface UseUpdateUserMutationReturnType {
  updateUser: (
    userID: string,
    standardAttributes: StandardAttributes,
    customAttributes: CustomAttributes
  ) => Promise<void>;
  loading: boolean;
  error: unknown;
}

export function useUpdateUserMutation(): UseUpdateUserMutationReturnType {
  const [mutationFunction, { error, loading }] =
    useMutation<UpdateUserMutationMutation>(UpdateUserMutationDocument);

  const updateUser = useCallback(
    async (
      userID: string,
      standardAttributes: StandardAttributes,
      customAttributes: CustomAttributes
    ) => {
      await mutationFunction({
        variables: {
          userID,
          standardAttributes,
          customAttributes,
        },
      });
    },
    [mutationFunction]
  );

  return {
    updateUser,
    error,
    loading,
  };
}
