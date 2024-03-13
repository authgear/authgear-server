import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
  ITag,
  Label,
} from "@fluentui/react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo, useState } from "react";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ErrorDialog from "../../error/ErrorDialog";
import { useAddRoleToGroupsMutation } from "../../graphql/adminapi/mutations/addRoleToGroupsMutation";
import { useRoleQuery } from "../../graphql/adminapi/query/roleQuery";
import StyledTagPicker from "../../StyledTagPicker";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "../../graphql/adminapi/query/groupsListQuery.generated";
import { useQuery } from "@apollo/client";
import { Group } from "../../graphql/adminapi/globalTypes.generated";

import styles from "./AddRoleGroupsDialog.module.css";

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

const dialogStyles = { main: { minHeight: 0 } };

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

    const [groupTags, setGroupTags] = useState<GroupTag[]>([]);

    const { addRoleToGroups, loading, error } = useAddRoleToGroupsMutation();
    const { refetch: refetchRole } = useRoleQuery(roleID, {
      skip: true,
    });
    const { refetch } = useQuery<
      GroupsListQueryQuery,
      GroupsListQueryQueryVariables
    >(GroupsListQueryDocument);

    const onChangeGroupTags = useCallback((tags?: ITag[]) => {
      if (tags === undefined) {
        setGroupTags([]);
      } else {
        setGroupTags(tags as GroupTag[]);
      }
    }, []);
    const [searchGroupKeyword, setSearchGroupKeyword] = useState<string>("");
    const onSearchInputChange = useCallback((value: string): string => {
      setSearchGroupKeyword(value);
      return value;
    }, []);
    const onClearGroupTags = useCallback(() => setGroupTags([]), []);

    const onResolveGroupSuggestions = useCallback(
      async (filter: string, selectedTags?: ITag[]): Promise<ITag[]> => {
        const selectedGroupIDs = new Set(
          selectedTags?.map((tag) => (tag as GroupTag).group.id)
        );
        const result = await refetch({ searchKeyword: filter, pageSize: 100 });

        if (result.data.groups?.edges == null) {
          return [];
        }
        return result.data.groups.edges.flatMap<GroupTag>((edge) => {
          const node = edge?.node;
          if (node == null) {
            return [];
          }
          // Filter out existing groups
          if (existingGroupIDs.has(node.id)) {
            return [];
          }

          // Filter out selected groups
          if (selectedGroupIDs.has(node.id)) {
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

    const onSubmit = useCallback(() => {
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
    }, [
      loading,
      isHidden,
      groupTags,
      addRoleToGroups,
      roleKey,
      refetchRole,
      roleID,
      onDismiss,
    ]);

    const modalProps = useMemo((): IModalProps => {
      return {
        onDismissed: () => {
          // Reset states on dismiss
          setGroupTags([]);

          propsOnDismissed?.();
        },
      };
    }, [propsOnDismissed]);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      return {
        title: renderToString("AddRoleGroupsDialog.title", {
          roleName: roleName ?? roleKey,
        }),
      };
    }, [renderToString, roleKey, roleName]);

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          modalProps={modalProps}
          dialogContentProps={dialogContentProps}
          styles={dialogStyles}
          maxWidth="560px"
        >
          <div className={styles.content}>
            <div className={styles.field}>
              <Label>
                <FormattedMessage id="AddRoleGroupsDialog.selectGroups" />
              </Label>
              <StyledTagPicker
                autoFocus={true}
                value={searchGroupKeyword}
                onInputChange={onSearchInputChange}
                selectedItems={groupTags}
                onChange={onChangeGroupTags}
                onResolveSuggestions={onResolveGroupSuggestions}
                onClearTags={onClearGroupTags}
              />
            </div>
          </div>
          <DialogFooter>
            <PrimaryButton
              disabled={loading}
              onClick={onSubmit}
              text={<FormattedMessage id="add" />}
            />
            <DefaultButton
              onClick={onDialogDismiss}
              disabled={loading}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
        <ErrorDialog error={error} />
      </>
    );
  };
