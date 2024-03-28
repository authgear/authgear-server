import { ITag } from "@fluentui/react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo } from "react";
import ErrorDialog from "../../../error/ErrorDialog";
import { useAddRoleToGroupsMutation } from "../../../graphql/adminapi/mutations/addRoleToGroupsMutation";
import { useRoleQuery } from "../../../graphql/adminapi/query/roleQuery";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "../../../graphql/adminapi/query/groupsListQuery.generated";
import { useQuery } from "@apollo/client";
import { Group } from "../../../graphql/adminapi/globalTypes.generated";
import AddTagsDialog from "./AddTagsDialog";

interface AddRoleGroupsDialogGroup extends Pick<Group, "id" | "key" | "name"> {}

interface AddRoleGroupsDialogProps {
  roleID: string;
  roleKey: string;
  roleName: string | null;
  roleGroups: AddRoleGroupsDialogGroup[];
  isHidden: boolean;

  onDismiss: () => void;
  onDismissed?: () => void;
}

interface GroupTag extends ITag {
  group: AddRoleGroupsDialogGroup;
}

function toGroupTag(group: AddRoleGroupsDialogGroup): GroupTag {
  return {
    key: group.id,
    name: group.name ?? group.key,
    group: group,
  };
}

export const AddRoleGroupsDialog: React.VFC<AddRoleGroupsDialogProps> =
  function AddRoleGroupsDialog({
    isHidden,
    onDismiss,
    onDismissed: propsOnDismissed,
    roleID,
    roleKey,
    roleName,
    roleGroups,
  }) {
    const { renderToString } = useContext(MessageContext);
    const existingGroupIDs = useMemo(() => {
      return new Set(roleGroups.map((group) => group.id));
    }, [roleGroups]);

    const { addRoleToGroups, loading, error } = useAddRoleToGroupsMutation();
    const { refetch: refetchRole } = useRoleQuery(roleID, {
      skip: true,
    });
    const { refetch } = useQuery<
      GroupsListQueryQuery,
      GroupsListQueryQueryVariables
    >(GroupsListQueryDocument);

    const onResolveGroupSuggestions = useCallback(
      async (filter: string, selectedTags?: ITag[]): Promise<ITag[]> => {
        const selectedGroupIDs = new Set(
          selectedTags?.map((tag) => (tag as GroupTag).group.id)
        );
        const excludedIDs = new Set([...selectedGroupIDs, ...existingGroupIDs]);
        const result = await refetch({
          searchKeyword: filter,
          excludedIDs: [...excludedIDs],
          pageSize: 100,
        });

        if (result.data.groups?.edges == null) {
          return [];
        }
        return result.data.groups.edges.flatMap<GroupTag>((edge) => {
          const node = edge?.node;
          if (node == null) {
            return [];
          }

          return [toGroupTag(node)];
        });
      },
      [existingGroupIDs, refetch]
    );

    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss();
    }, [isHidden, loading, onDismiss]);

    const onSubmit = useCallback(
      (tags: ITag[]) => {
        const groupTags = tags as GroupTag[];
        if (loading || isHidden || groupTags.length === 0) {
          return;
        }
        addRoleToGroups(
          roleKey,
          groupTags.map((tag) => tag.group.key)
        )
          .then(async () => {
            // Update the cache
            return refetchRole({ roleID: roleID });
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
        addRoleToGroups,
        roleKey,
        refetchRole,
        roleID,
        onDismiss,
      ]
    );

    const dialogTitle = renderToString("AddRoleGroupsDialog.title", {
      roleName: roleName ?? roleKey,
    });
    const tagPickerLabel = renderToString("AddRoleGroupsDialog.selectGroups");

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
