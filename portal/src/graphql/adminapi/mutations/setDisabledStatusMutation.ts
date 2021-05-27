import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { SetDisabledStatusMutation } from "./__generated__/SetDisabledStatusMutation";

const setDisabledStatusMutation = gql`
  mutation SetDisabledStatusMutation($userID: ID!, $isDisabled: Boolean!) {
    setDisabledStatus(input: { userID: $userID, isDisabled: $isDisabled }) {
      user {
        id
        isDisabled
      }
    }
  }
`;

export function useSetDisabledStatusMutation(userID: string): {
  setDisabledStatus: (isDisabled: boolean) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<SetDisabledStatusMutation>(setDisabledStatusMutation);

  const setDisabledStatus = useCallback(
    async (isDisabled: boolean) => {
      const result = await mutationFunction({
        variables: {
          userID,
          isDisabled,
        },
      });

      return !!result.data;
    },
    [mutationFunction, userID]
  );

  return { setDisabledStatus, loading, error };
}
