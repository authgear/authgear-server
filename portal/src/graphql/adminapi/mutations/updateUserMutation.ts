import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { StandardAttributes, CustomAttributes } from "../../../types";
import { UpdateUserMutation } from "./__generated__/UpdateUserMutation";

const updateUserMutation = gql`
  mutation UpdateUserMutation(
    $userID: ID!
    $standardAttributes: UserStandardAttributes!
    $customAttributes: UserCustomAttributes!
  ) {
    updateUser(
      input: {
        userID: $userID
        standardAttributes: $standardAttributes
        customAttributes: $customAttributes
      }
    ) {
      user {
        id
        updatedAt
        standardAttributes
        customAttributes
      }
    }
  }
`;

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
    useMutation<UpdateUserMutation>(updateUserMutation);

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
