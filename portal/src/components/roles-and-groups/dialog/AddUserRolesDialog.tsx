import { ITag } from "@fluentui/react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo } from "react";
import ErrorDialog from "../../../error/ErrorDialog";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "../../../graphql/adminapi/query/rolesListQuery.generated";
import { useQuery } from "@apollo/client";
import { Role } from "../../../graphql/adminapi/globalTypes.generated";
import AddTagsDialog from "./AddTagsDialog";
import { useAddUserToRolesMutation } from "../../../graphql/adminapi/mutations/addUserToRolesMutation";
import { useUserQuery } from "../../../graphql/adminapi/query/userQuery";
import { extractRawID } from "../../../util/graphql";

interface AddUserRolesDialogRole extends Pick<Role, "id" | "key" | "name"> {}

interface AddUserRolesDialogProps {
  userID: string;
  userFormattedName: string | null;
  userEndUserAccountID: string | null;
  userRoles: AddUserRolesDialogRole[];
  isHidden: boolean;

  onDismiss: () => void;
  onDismissed?: () => void;
}

interface RoleTag extends ITag {
  role: AddUserRolesDialogRole;
}

function toRoleTag(role: AddUserRolesDialogRole): RoleTag {
  return {
    key: role.id,
    name: role.name ?? role.key,
    role: role,
  };
}

export const AddUserRolesDialog: React.VFC<AddUserRolesDialogProps> =
  function AddUserRolesDialog({
    isHidden,
    onDismiss,
    onDismissed: propsOnDismissed,
    userID,
    userFormattedName,
    userEndUserAccountID,
    userRoles,
  }) {
    const { renderToString } = useContext(MessageContext);
    const existingRoleIDs = useMemo(() => {
      return new Set(userRoles.map((role) => role.id));
    }, [userRoles]);

    const { addUserToRoles, loading, error } = useAddUserToRolesMutation();
    const { refetch: refetchUser } = useUserQuery(userID, {
      skip: true,
    });
    const { refetch } = useQuery<
      RolesListQueryQuery,
      RolesListQueryQueryVariables
    >(RolesListQueryDocument);

    const onResolveRoleSuggestions = useCallback(
      async (filter: string, selectedTags?: ITag[]): Promise<ITag[]> => {
        const selectedRoleIDs = new Set(
          selectedTags?.map((tag) => (tag as RoleTag).role.id)
        );
        const excludedIDs = new Set([...selectedRoleIDs, ...existingRoleIDs]);
        const result = await refetch({
          searchKeyword: filter,
          excludedIDs: [...excludedIDs],
          pageSize: 100,
        });

        if (result.data.roles?.edges == null) {
          return [];
        }
        return result.data.roles.edges.flatMap<RoleTag>((edge) => {
          const node = edge?.node;
          if (node == null) {
            return [];
          }

          return [toRoleTag(node)];
        });
      },
      [existingRoleIDs, refetch]
    );

    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss();
    }, [isHidden, loading, onDismiss]);

    const onSubmit = useCallback(
      (tags: ITag[]) => {
        const roleTags = tags as RoleTag[];
        if (loading || isHidden || roleTags.length === 0) {
          return;
        }
        addUserToRoles(
          userID,
          roleTags.map((tag) => tag.role.key)
        )
          .then(async () => {
            // Update the cache
            return refetchUser({ userID });
          })
          .then(
            () => onDismiss(),
            (e: unknown) => {
              onDismiss();
              throw e;
            }
          );
      },
      [loading, isHidden, addUserToRoles, userID, refetchUser, onDismiss]
    );

    const dialogTitle = renderToString("AddUserRolesDialog.title", {
      userName:
        userFormattedName ?? userEndUserAccountID ?? extractRawID(userID),
    });
    const tagPickerLabel = renderToString("AddUserRolesDialog.selectRoles");

    return (
      <>
        <AddTagsDialog
          isHidden={isHidden}
          isLoading={loading}
          title={dialogTitle}
          tagPickerLabel={tagPickerLabel}
          onDismiss={onDialogDismiss}
          onDismissed={propsOnDismissed}
          onSubmit={onSubmit}
          onResolveSuggestions={onResolveRoleSuggestions}
        />
        <ErrorDialog error={error} />
      </>
    );
  };
