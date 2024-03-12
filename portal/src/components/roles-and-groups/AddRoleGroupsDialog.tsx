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
    onDismissed,
    roleID,
    roleKey,
    roleName,
    roleGroups,
  }) {
    const { renderToString } = useContext(MessageContext);

    const [groupTags, setGroupTags] = useState<GroupTag[]>([]);

    const { addRoleToGroups, loading, error } = useAddRoleToGroupsMutation();
    const { refetch: refetchRole } = useRoleQuery(roleID, {
      skip: true,
    });
    const { refetch } = useQuery<
      GroupsListQueryQuery,
      GroupsListQueryQueryVariables
    >(GroupsListQueryDocument, {
      variables: {
        pageSize: roleGroups.length + groupTags.length + 4,
        searchKeyword: "",
      },
      fetchPolicy: "network-only",
      skip: true,
    });

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
      async (filter: string): Promise<ITag[]> => {
        const result = await refetch({ searchKeyword: filter });

        if (result.data.groups?.edges == null) {
          return [];
        }
        return result.data.groups.edges.flatMap<GroupTag>((edge) => {
          if (edge?.node == null) {
            return [];
          }
          return [toGroupTag(edge.node)];
        });
      },
      [refetch]
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
        onDismissed,
      };
    }, [onDismissed]);

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
