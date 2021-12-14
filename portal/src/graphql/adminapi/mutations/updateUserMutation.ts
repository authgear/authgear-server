import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { produce } from "immer";
import { StandardAttributes } from "../../../types";
import { UpdateUserMutation } from "./__generated__/UpdateUserMutation";

const updateUserMutation = gql`
  mutation UpdateUserMutation(
    $userID: ID!
    $standardAttributes: UserStandardAttributes!
  ) {
    updateUser(
      input: { userID: $userID, standardAttributes: $standardAttributes }
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
    standardAttributes: StandardAttributes
  ) => Promise<StandardAttributes>;
  loading: boolean;
  error: unknown;
}

function sanitize(attrs: StandardAttributes): StandardAttributes {
  return produce(attrs, (attrs) => {
    delete attrs.updated_at;
    delete attrs.email_verified;
    delete attrs.phone_number_verified;

    for (const key of Object.keys(attrs)) {
      // @ts-expect-error
      const value = attrs[key];
      if (value === "") {
        // @ts-expect-error
        delete attrs[key];
      }
    }

    if (attrs.address != null) {
      for (const key of Object.keys(attrs.address)) {
        // @ts-expect-error
        const value = attrs.address[key];
        if (value === "") {
          // @ts-expect-error
          delete attrs.address[key];
        }
      }
      if (Object.keys(attrs.address).length === 0) {
        delete attrs.address;
      }
    }
  });
}

export function useUpdateUserMutation(): UseUpdateUserMutationReturnType {
  const [mutationFunction, { error, loading }] =
    useMutation<UpdateUserMutation>(updateUserMutation);

  const updateUser = useCallback(
    async (userID: string, standardAttributes: StandardAttributes) => {
      const result = await mutationFunction({
        variables: {
          userID,
          standardAttributes: sanitize(standardAttributes),
        },
      });
      return result.data?.updateUser.user.standardAttributes ?? {};
    },
    [mutationFunction]
  );

  return {
    updateUser,
    error,
    loading,
  };
}
