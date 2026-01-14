import { ITag } from "@fluentui/react";
import { Context as MessageContext } from "../../../intl";
import React, { useCallback, useContext, useMemo } from "react";
import ErrorDialog from "../../../error/ErrorDialog";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "../../../graphql/adminapi/query/groupsListQuery.generated";
import { useQuery } from "@apollo/client";
import { Group } from "../../../graphql/adminapi/globalTypes.generated";
import AddTagsDialog from "./AddTagsDialog";
import { useAddUserToGroupsMutation } from "../../../graphql/adminapi/mutations/addUserToGroupsMutation";
import { useUserQuery } from "../../../graphql/adminapi/query/userQuery";
import { extractRawID } from "../../../util/graphql";

interface AddUserGroupsDialogGroup extends Pick<Group, "id" | "key" | "name"> {}

interface AddUserGroupsDialogProps {
  userID: string;
  userFormattedName: string | null;
  userEndUserAccountID: string | null;
  userGroups: AddUserGroupsDialogGroup[];
  isHidden: boolean;

  onDismiss: () => void;
  onDismissed?: () => void;
}

interface GroupTag extends ITag {
  group: AddUserGroupsDialogGroup;
}

function toGroupTag(group: AddUserGroupsDialogGroup): GroupTag {
  return {
    key: group.id,
    name: group.name ?? group.key,
    group: group,
  };
}

export const AddUserGroupsDialog: React.VFC<AddUserGroupsDialogProps> =
  function AddUserGroupsDialog({
    isHidden,
    onDismiss,
    onDismissed: propsOnDismissed,
    userID,
    userFormattedName,
    userEndUserAccountID,
    userGroups,
  }) {
    const { renderToString } = useContext(MessageContext);
    const existingGroupIDs = useMemo(() => {
      return new Set(userGroups.map((group) => group.id));
    }, [userGroups]);

    const { addUserToGroups, loading, error } = useAddUserToGroupsMutation();
    const { refetch: refetchUser } = useUserQuery(userID, {
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
        addUserToGroups(
          userID,
          groupTags.map((tag) => tag.group.key)
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
      [loading, isHidden, addUserToGroups, userID, refetchUser, onDismiss]
    );

    const dialogTitle = renderToString("AddUserGroupsDialog.title", {
      userName:
        userFormattedName ?? userEndUserAccountID ?? extractRawID(userID),
    });
    const tagPickerLabel = renderToString("AddUserGroupsDialog.selectGroups");

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
