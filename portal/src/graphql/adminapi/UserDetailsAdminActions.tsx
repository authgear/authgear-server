import { Label, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React from "react";
import cn from "classnames";
import styles from "./UserDetailsAdminActions.module.css";
import ListCellLayout from "../../ListCellLayout";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";

interface DisableUserCellProps {
  isDisabled: boolean;
}

interface AnonymizeUserCellProps {
  isScheduledAnonymization: boolean;
}

interface RemoveUserCellProps {
  isScheduledRemoval: boolean;
}

const DisableUserCell: React.VFC<DisableUserCellProps> =
  function DisableUserCell(props) {
    const { themes } = useSystemConfig();
    const { isDisabled } = props;
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
          />
        )}
      </ListCellLayout>
    );
  };

const AnonymizeUserCell: React.VFC<AnonymizeUserCellProps> =
  function AnonymizeUserCell(props) {
    const { themes } = useSystemConfig();
    const { isScheduledAnonymization } = props;
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
          />
        )}
      </ListCellLayout>
    );
  };

const RemoveUserCell: React.VFC<RemoveUserCellProps> = function RemoveUserCell(
  props
) {
  const { themes } = useSystemConfig();
  const { isScheduledRemoval } = props;
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
        />
      )}
    </ListCellLayout>
  );
};

const UserDetailsAdminActions: React.VFC = function UserDetailsAdminActions() {
  return (
    <div>
      <Label className={styles.standardAttributesTitle}>
        <Text variant="xLarge">
          <FormattedMessage id="UserDetailsAdminActions.title" />
        </Text>
      </Label>
      <DisableUserCell isDisabled={false} />
      <AnonymizeUserCell isScheduledAnonymization={false} />
      <RemoveUserCell isScheduledRemoval={false} />
    </div>
  );
};

export default UserDetailsAdminActions;
