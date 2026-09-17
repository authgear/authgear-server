import React, { useCallback, useContext, useState } from "react";
import { Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { CardTable } from "../v2/CardTable/CardTable";
import { Toggle } from "../v2/Toggle/Toggle";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { useCalloutToast } from "../v2/Callout/Callout";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useUpdateResourceMutationMutation } from "../../graphql/adminapi/mutations/updateResourceMutation.generated";
import { ResourceQueryDocument } from "../../graphql/adminapi/query/resourceQuery.generated";
import { parseRawError } from "../../error/parse";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import {
  VISIBLE_ACCESS_POLICY_KEYS,
  VisibleAccessPolicyKey,
  accessPolicyDescriptionIDs,
  accessPolicyInputForKey,
  accessPolicyLabelIDs,
} from "./accessPolicy";
import styles from "./ResourceAccessPolicyTable.module.css";

const SAVED_TOAST_DURATION_MS = 2000;

export interface ResourceAccessPolicyTableProps {
  resource: Resource;
}

// One row per client category that may request the resource on a user's
// behalf. Each toggle saves on change; turning one off is confirmed first
// because it cuts those clients off on their next token refresh.
export const ResourceAccessPolicyTable: React.VFC<ResourceAccessPolicyTableProps> =
  function ResourceAccessPolicyTable({ resource }) {
    const { renderToString } = useContext(Context);
    const { setErrors } = useErrorMessageBarContext();
    const { showToast } = useCalloutToast();
    const [updateResource, { loading: isUpdating }] =
      useUpdateResourceMutationMutation();
    const [pendingKey, setPendingKey] = useState<VisibleAccessPolicyKey | null>(
      null
    );

    const save = useCallback(
      async (key: VisibleAccessPolicyKey, allowed: boolean) => {
        try {
          await updateResource({
            variables: {
              input: {
                resourceURI: resource.resourceURI,
                accessPolicy: accessPolicyInputForKey(key, allowed),
              },
            },
            refetchQueries: [ResourceQueryDocument],
            awaitRefetchQueries: true,
          });
          // Names the one setting that was written.
          showToast({
            type: "success",
            text: (
              <FormattedMessage
                id={
                  allowed
                    ? "ResourceAccessPolicyTable.toast.on"
                    : "ResourceAccessPolicyTable.toast.off"
                }
                values={{ category: renderToString(accessPolicyLabelIDs[key]) }}
              />
            ),
            duration: SAVED_TOAST_DURATION_MS,
          });
        } catch (e: unknown) {
          setErrors(parseRawError(e));
        }
      },
      [
        updateResource,
        resource.resourceURI,
        setErrors,
        showToast,
        renderToString,
      ]
    );

    const onToggle = useCallback(
      (key: VisibleAccessPolicyKey, allowed: boolean) => {
        if (!allowed) {
          setPendingKey(key);
          return;
        }
        void save(key, true);
      },
      [save]
    );

    const onConfirmDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open && !isUpdating) {
          setPendingKey(null);
        }
      },
      [isUpdating]
    );

    const onCancel = useCallback(() => {
      if (!isUpdating) {
        setPendingKey(null);
      }
    }, [isUpdating]);

    const onConfirm = useCallback(() => {
      if (pendingKey == null) {
        return;
      }
      save(pendingKey, false).finally(() => {
        setPendingKey(null);
      });
    }, [pendingKey, save]);

    return (
      <>
        <CardTable>
          <CardTable.Header>
            <CardTable.HeaderCell className={styles.colClients}>
              <FormattedMessage id="ResourceAccessPolicyTable.columns.clients" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colAllowed}>
              <FormattedMessage id="ResourceAccessPolicyTable.columns.allowed" />
            </CardTable.HeaderCell>
          </CardTable.Header>
          {VISIBLE_ACCESS_POLICY_KEYS.map((key) => (
            <CardTable.Row key={key}>
              <CardTable.Cell className={styles.colClients}>
                <div className={styles.clientText}>
                  <Text size="2" weight="medium">
                    <FormattedMessage id={accessPolicyLabelIDs[key]} />
                  </Text>
                  <Text size="1" color="gray">
                    <FormattedMessage id={accessPolicyDescriptionIDs[key]} />
                  </Text>
                </div>
              </CardTable.Cell>
              <CardTable.Cell className={styles.colAllowed}>
                <Toggle
                  checked={resource.accessPolicy[key]}
                  disabled={isUpdating}
                  onCheckedChange={(checked) => onToggle(key, checked)}
                />
              </CardTable.Cell>
            </CardTable.Row>
          ))}
        </CardTable>
        <ConfirmationDialog
          open={pendingKey != null}
          onOpenChange={onConfirmDialogOpenChange}
          title={
            <FormattedMessage id="ResourceAccessPolicyTable.confirm.title" />
          }
          description={
            <FormattedMessage
              id="ResourceAccessPolicyTable.confirm.description"
              values={{
                category:
                  pendingKey != null
                    ? renderToString(accessPolicyLabelIDs[pendingKey])
                    : "",
              }}
            />
          }
          confirmText={
            <FormattedMessage id="ResourceAccessPolicyTable.confirm.confirm" />
          }
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={onConfirm}
          onCancel={onCancel}
          loading={isUpdating}
          confirmColor="red"
        />
      </>
    );
  };
