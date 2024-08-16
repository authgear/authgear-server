import { Label, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React from "react";
import cn from "classnames";
import styles from "./UserDetailsAdminActions.module.css";
import ListCellLayout from "../../ListCellLayout";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { UserQueryNodeFragment } from "./query/userQuery.generated";

interface DisableUserCellProps {
  isDisabled: boolean;
  onDisableData: () => void;
}

interface AnonymizeUserCellProps {
  isScheduledAnonymization: boolean;
  onAnonymizeData: () => void;
  onCancelAnonymizeData: () => void;
}

interface RemoveUserCellProps {
  isScheduledRemoval: boolean;
  onRemoveData: () => void;
  onCancelRemoveData: () => void;
}

const DisableUserCell: React.VFC<DisableUserCellProps> =
  function DisableUserCell(props) {
    const { themes } = useSystemConfig();
    const { isDisabled, onDisableData } = props;
    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.actionCell,
          styles["cell--not-first"]
        )}
      >
        <Text className={cn(styles.cellLabel, styles.actionCellLabel)}>
          <FormattedMessage id="UserDetailsAdminActions.disable-user.title" />
        </Text>
        <Text className={cn(styles.actionCellBody, styles.actionCellLast)}>
          <FormattedMessage id="UserDetailsAdminActions.disable-user.body" />
        </Text>
        {isDisabled ? (
          <DefaultButton
            theme={themes.actionButton}
            className={cn(styles.actionCellActionButton, styles.actionBorder)}
            iconProps={{ iconName: "Play" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.disable-user.action.enable" />
            }
            onClick={onDisableData}
          />
        ) : (
          <DefaultButton
            theme={themes.destructive}
            className={cn(
              styles.actionCellActionButton,
              styles.destructiveBorder
            )}
            iconProps={{ iconName: "Blocked" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.disable-user.action.disable" />
            }
            onClick={onDisableData}
          />
        )}
      </ListCellLayout>
    );
  };

const AnonymizeUserCell: React.VFC<AnonymizeUserCellProps> =
  function AnonymizeUserCell(props) {
    const { themes } = useSystemConfig();
    const { isScheduledAnonymization, onAnonymizeData, onCancelAnonymizeData } =
      props;
    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.actionCell,
          styles["cell--not-first"]
        )}
      >
        <Text className={cn(styles.cellLabel, styles.actionCellLabel)}>
          <FormattedMessage id="UserDetailsAdminActions.anonymize-user.title" />
        </Text>
        <Text className={cn(styles.actionCellBody, styles.actionCellLast)}>
          <FormattedMessage id="UserDetailsAdminActions.anonymize-user.body" />
        </Text>
        {isScheduledAnonymization ? (
          <DefaultButton
            theme={themes.actionButton}
            className={cn(styles.actionCellActionButton, styles.actionBorder)}
            iconProps={{ iconName: "Undo" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.anonymize-user.action.cancel" />
            }
            onClick={onCancelAnonymizeData}
          />
        ) : (
          <DefaultButton
            theme={themes.destructive}
            className={cn(
              styles.actionCellActionButton,
              styles.destructiveBorder
            )}
            iconProps={{ iconName: "Archive" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.anonymize-user.action.anonymize" />
            }
            onClick={onAnonymizeData}
          />
        )}
      </ListCellLayout>
    );
  };

const RemoveUserCell: React.VFC<RemoveUserCellProps> = function RemoveUserCell(
  props
) {
  const { themes } = useSystemConfig();
  const { isScheduledRemoval, onRemoveData, onCancelRemoveData } = props;
  return (
    <ListCellLayout
      className={cn(styles.cell, styles.actionCell, styles["cell--not-first"])}
    >
      <Text className={cn(styles.cellLabel, styles.actionCellLabel)}>
        <FormattedMessage id="UserDetailsAdminActions.remove-user.title" />
      </Text>
      <Text className={cn(styles.actionCellBody, styles.actionCellLast)}>
        <FormattedMessage id="UserDetailsAdminActions.remove-user.body" />
      </Text>
      {isScheduledRemoval ? (
        <DefaultButton
          theme={themes.actionButton}
          className={cn(styles.actionCellActionButton, styles.actionBorder)}
          iconProps={{ iconName: "Undo" }}
          text={
            <FormattedMessage id="UserDetailsAdminActions.remove-user.action.cancel" />
          }
          onClick={onCancelRemoveData}
        />
      ) : (
        <DefaultButton
          theme={themes.destructive}
          className={cn(
            styles.actionCellActionButton,
            styles.destructiveBorder
          )}
          iconProps={{ iconName: "Delete" }}
          text={
            <FormattedMessage id="UserDetailsAdminActions.remove-user.action.remove" />
          }
          onClick={onRemoveData}
        />
      )}
    </ListCellLayout>
  );
};

interface UserDetailsAdminActionsProps {
  data: UserQueryNodeFragment;
  onRemoveData: () => void;
  onAnonymizeData: () => void;
  handleDataStatusChange: () => void;
}

const UserDetailsAdminActions: React.VFC<UserDetailsAdminActionsProps> =
  function UserDetailsAdminActions(props) {
    const { data, onRemoveData, onAnonymizeData, handleDataStatusChange } =
      props;

    return (
      <div>
        <Label className={styles.standardAttributesTitle}>
          <Text variant="xLarge">
            <FormattedMessage id="UserDetailsAdminActions.title" />
          </Text>
        </Label>
        {data.isAnonymized || data.deleteAt != null ? null : (
          <>
            <DisableUserCell
              isDisabled={data.isDisabled}
              onDisableData={handleDataStatusChange}
            />
            <AnonymizeUserCell
              isScheduledAnonymization={data.anonymizeAt != null}
              onAnonymizeData={onAnonymizeData}
              onCancelAnonymizeData={handleDataStatusChange}
            />
          </>
        )}
        <RemoveUserCell
          isScheduledRemoval={data.deleteAt != null}
          onRemoveData={onRemoveData}
          onCancelRemoveData={handleDataStatusChange}
        />
      </div>
    );
  };

export default UserDetailsAdminActions;
