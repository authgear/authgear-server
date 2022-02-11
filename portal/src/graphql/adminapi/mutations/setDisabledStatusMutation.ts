import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { SetDisabledStatusMutation } from "./__generated__/SetDisabledStatusMutation";

const setDisabledStatusMutation = gql`
  mutation SetDisabledStatusMutation($userID: ID!, $isDisabled: Boolean!) {
    setDisabledStatus(input: { userID: $userID, isDisabled: $isDisabled }) {
      user {
        id
        isDisabled
        disableReason
        isDeactivated
        deleteAt
      }
    }
  }
`;

export function useSetDisabledStatusMutation(): {
  setDisabledStatus: (userID: string, isDisabled: boolean) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<SetDisabledStatusMutation>(setDisabledStatusMutation, {
      // Disabling a user will terminate all sessions.
      // So we have to refetch queries that fetch sessions.
      refetchQueries: ["UserQuery"],
    });

  const setDisabledStatus = useCallback(
    async (userID: string, isDisabled: boolean) => {
      const result = await mutationFunction({
        variables: {
          userID,
          isDisabled,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { setDisabledStatus, loading, error };
}
