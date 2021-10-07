import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
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
  const output: StandardAttributes = {
    ...attrs,
  };

  delete output.updated_at;
  delete output.email_verified;
  delete output.phone_number_verified;

  for (const key of Object.keys(output)) {
    // @ts-expect-error
    const value = output[key];
    if (value === "") {
      // @ts-expect-error
      delete output[key];
    }
  }

  if (output.address != null) {
    for (const key of Object.keys(output.address)) {
      // @ts-expect-error
      const value = output.address[key];
      if (value === "") {
        // @ts-expect-error
        delete output.address[key];
      }
    }
    if (Object.keys(output.address).length === 0) {
      delete output.address;
    }
  }

  return output;
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
