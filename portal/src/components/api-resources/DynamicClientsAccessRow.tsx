import React, { useCallback, useState } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { Toggle } from "../v2/Toggle/Toggle";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useUpdateResourceMutationMutation } from "../../graphql/adminapi/mutations/updateResourceMutation.generated";
import { ResourceQueryDocument } from "../../graphql/adminapi/query/resourceQuery.generated";
import { parseRawError } from "../../error/parse";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import styles from "./DynamicClientsAccessRow.module.css";

export interface DynamicClientsAccessRowProps {
  resource: Resource;
}

export const DynamicClientsAccessRow: React.VFC<DynamicClientsAccessRowProps> =
  function DynamicClientsAccessRow({ resource }) {
    const { setErrors } = useErrorMessageBarContext();

    const [updateResource, { loading: isUpdating }] =
      useUpdateResourceMutationMutation();
    const [isConfirmDialogVisible, setIsConfirmDialogVisible] = useState(false);

    const allowed = resource.accessPolicy.allowDynamicThirdPartyClientAccess;

    const setAllowed = useCallback(
      async (checked: boolean) => {
        try {
          await updateResource({
            variables: {
              input: {
                resourceURI: resource.resourceURI,
                accessPolicy: {
                  allowDynamicThirdPartyClientAccess: checked,
                },
              },
            },
            refetchQueries: [ResourceQueryDocument],
            awaitRefetchQueries: true,
          });
        } catch (e: unknown) {
          setErrors(parseRawError(e));
        }
      },
      [updateResource, resource.resourceURI, setErrors]
    );

    const onToggleChange = useCallback(
      (checked: boolean) => {
        if (!checked) {
          setIsConfirmDialogVisible(true);
          return;
        }
        // setAllowed swallows its own errors, so there is nothing to chain;
        // void it rather than an empty .finally (which would not even mark a
        // rejection as handled). onConfirmDisallow's .finally below is real.
        void setAllowed(true);
      },
      [setAllowed]
    );

    const onConfirmDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open && !isUpdating) {
          setIsConfirmDialogVisible(false);
        }
      },
      [isUpdating]
    );

    const onCancelDisallow = useCallback(() => {
      if (!isUpdating) {
        setIsConfirmDialogVisible(false);
      }
    }, [isUpdating]);

    const onConfirmDisallow = useCallback(() => {
      setAllowed(false).finally(() => {
        setIsConfirmDialogVisible(false);
      });
    }, [setAllowed]);

    return (
      <div className={styles.row}>
        <div className={styles.rowText}>
          <Text size="2" weight="medium">
            <FormattedMessage id="DynamicClientsAccessRow.title" />
          </Text>
          <Text size="1" color="gray">
            <FormattedMessage id="DynamicClientsAccessRow.description" />
          </Text>
        </div>
        <Toggle
          checked={allowed}
          disabled={isUpdating}
          onCheckedChange={onToggleChange}
        />
        <ConfirmationDialog
          open={isConfirmDialogVisible}
          onOpenChange={onConfirmDialogOpenChange}
          title={
            <FormattedMessage id="DynamicClientsAccessRow.confirm.title" />
          }
          description={
            <FormattedMessage id="DynamicClientsAccessRow.confirm.description" />
          }
          confirmText={
            <FormattedMessage id="DynamicClientsAccessRow.confirm.confirm" />
          }
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={onConfirmDisallow}
          onCancel={onCancelDisallow}
          loading={isUpdating}
          confirmColor="red"
        />
      </div>
    );
  };
