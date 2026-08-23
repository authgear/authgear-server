import React, { useCallback, useContext, useState } from "react";
import { Dialog, DialogFooter, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import Toggle from "../../Toggle";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
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
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
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
      (_event: React.MouseEvent<HTMLElement>, checked?: boolean) => {
        if (checked == null) {
          return;
        }
        if (!checked) {
          setIsConfirmDialogVisible(true);
          return;
        }
        setAllowed(true).finally(() => {});
      },
      [setAllowed]
    );

    const onDismissConfirmDialog = useCallback(() => {
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
          <Text block={true} styles={{ root: { fontWeight: 600 } }}>
            <FormattedMessage id="DynamicClientsAccessRow.title" />
          </Text>
          <Text
            block={true}
            variant="small"
            styles={{ root: { color: themes.main.palette.neutralSecondary } }}
          >
            <FormattedMessage id="DynamicClientsAccessRow.description" />
          </Text>
        </div>
        <Toggle
          checked={allowed}
          disabled={isUpdating}
          onChange={onToggleChange}
        />
        <Dialog
          hidden={!isConfirmDialogVisible}
          dialogContentProps={{
            title: renderToString("DynamicClientsAccessRow.confirm.title"),
            subText: renderToString(
              "DynamicClientsAccessRow.confirm.description"
            ),
          }}
          modalProps={{ isBlocking: isUpdating }}
          onDismiss={onDismissConfirmDialog}
        >
          <DialogFooter>
            <PrimaryButton
              theme={themes.destructive}
              disabled={isUpdating}
              onClick={onConfirmDisallow}
              text={renderToString("DynamicClientsAccessRow.confirm.confirm")}
            />
            <DefaultButton
              onClick={onDismissConfirmDialog}
              disabled={isUpdating}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
      </div>
    );
  };
