import { ITag } from "@fluentui/react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo } from "react";
import ErrorDialog from "../../../error/ErrorDialog";
import { useQuery } from "@apollo/client";
import { Role } from "../../../graphql/adminapi/globalTypes.generated";
import AddTagsDialog from "./AddTagsDialog";
import { useAddGroupToRolesMutation } from "../../../graphql/adminapi/mutations/addGroupToRolesMutation";
import { useGroupQuery } from "../../../graphql/adminapi/query/groupQuery";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "../../../graphql/adminapi/query/rolesListQuery.generated";

interface AddGroupRolesDialogGroup extends Pick<Role, "id" | "key" | "name"> {}

interface AddGroupRolesDialogProps {
  groupID: string;
  groupKey: string;
  groupName: string | null;
  groupRoles: AddGroupRolesDialogGroup[];
  isHidden: boolean;

  onDismiss: () => void;
  onDismissed?: () => void;
}

interface RoleTag extends ITag {
  role: AddGroupRolesDialogGroup;
}

function toRoleTag(role: AddGroupRolesDialogGroup): RoleTag {
  return {
    key: role.id,
    name: role.name ?? role.key,
    role: role,
  };
}

export const AddGroupRolesDialog: React.VFC<AddGroupRolesDialogProps> =
  function AddGroupRolesDialog({
    isHidden,
    onDismiss,
    onDismissed: propsOnDismissed,
    groupID,
    groupKey,
    groupName,
    groupRoles,
  }) {
    const { renderToString } = useContext(MessageContext);
    const existingRoleIDs = useMemo(() => {
      return new Set(groupRoles.map((group) => group.id));
    }, [groupRoles]);

    const { addGroupToRoles, loading, error } = useAddGroupToRolesMutation();
    const { refetch: refetchGroup } = useGroupQuery(groupID, {
      skip: true,
    });
    const { refetch } = useQuery<
      RolesListQueryQuery,
      RolesListQueryQueryVariables
    >(RolesListQueryDocument);

    const onResolveGroupSuggestions = useCallback(
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
        addGroupToRoles(
          groupKey,
          roleTags.map((tag) => tag.role.key)
        )
          .then(async () => {
            // Update the cache
            return refetchGroup({ groupID: groupID });
          })
          .then(
            () => onDismiss(),
            (e: unknown) => {
              onDismiss();
              throw e;
            }
          );
      },
      [
        loading,
        isHidden,
        addGroupToRoles,
        groupKey,
        refetchGroup,
        groupID,
        onDismiss,
      ]
    );

    const dialogTitle = renderToString("AddGroupRolesDialog.title", {
      groupName: groupName ?? groupKey,
    });
    const tagPickerLabel = renderToString("AddGroupRolesDialog.selectGroups");

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
          onResolveSuggestions={onResolveGroupSuggestions}
        />
        <ErrorDialog error={error} />
      </>
    );
  };
