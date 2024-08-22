import { IStyle, Label, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React from "react";
import cn from "classnames";
import styles from "./UserDetailsAdminActions.module.css";
import ListCellLayout from "../../ListCellLayout";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import AdminActionButton from "./AdminActionButton";

const labelTextStyle: IStyle = {
  lineHeight: "20px",
  fontWeight: 600,
};

const bodyTextStyle: IStyle = {
  lineHeight: "20px",
  maxWidth: "500px",
};

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
        className={cn(styles.actionCell, styles["cell--not-first"])}
      >
        <Text
          className={cn(styles.actionCellLabel)}
          styles={{
            root: labelTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAdminActions.disable-user.title" />
        </Text>
        <Text
          className={cn(styles.actionCellBody)}
          styles={{
            root: bodyTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAdminActions.disable-user.body" />
        </Text>
        {isDisabled ? (
          <AdminActionButton
            theme={themes.actionButton}
            className={cn(styles.actionCellActionButton)}
            iconProps={{ iconName: "Play" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.disable-user.action.enable" />
            }
            onClick={onDisableData}
          />
        ) : (
          <AdminActionButton
            theme={themes.destructive}
            className={cn(styles.actionCellActionButton)}
            iconProps={{ iconName: "Blocked" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.disable-user.action.disable" />
            }
            onClick={onDisableData}
          />
        )}
        <div className={cn(styles.actionCellSpacer)} />
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
        className={cn(styles.actionCell, styles["cell--not-first"])}
      >
        <Text
          className={cn(styles.actionCellLabel)}
          styles={{
            root: labelTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAdminActions.anonymize-user.title" />
        </Text>
        <Text
          className={cn(styles.actionCellBody)}
          styles={{
            root: bodyTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAdminActions.anonymize-user.body" />
        </Text>
        {isScheduledAnonymization ? (
          <AdminActionButton
            theme={themes.actionButton}
            className={cn(styles.actionCellActionButton)}
            iconProps={{ iconName: "Undo" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.anonymize-user.action.cancel" />
            }
            onClick={onCancelAnonymizeData}
          />
        ) : (
          <AdminActionButton
            theme={themes.destructive}
            className={cn(styles.actionCellActionButton)}
            iconProps={{ iconName: "Archive" }}
            text={
              <FormattedMessage id="UserDetailsAdminActions.anonymize-user.action.anonymize" />
            }
            onClick={onAnonymizeData}
          />
        )}
        <div className={cn(styles.actionCellSpacer)} />
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
      className={cn(
        styles.actionCell,
        styles["cell--not-first"],
        styles["cell--last"]
      )}
    >
      <Text
        className={cn(styles.actionCellLabel)}
        styles={{
          root: labelTextStyle,
        }}
      >
        <FormattedMessage id="UserDetailsAdminActions.remove-user.title" />
      </Text>
      <Text
        className={cn(styles.actionCellBody)}
        styles={{
          root: bodyTextStyle,
        }}
      >
        <FormattedMessage id="UserDetailsAdminActions.remove-user.body" />
      </Text>
      {isScheduledRemoval ? (
        <AdminActionButton
          theme={themes.actionButton}
          className={cn(styles.actionCellActionButton)}
          iconProps={{ iconName: "Undo" }}
          text={
            <FormattedMessage id="UserDetailsAdminActions.remove-user.action.cancel" />
          }
          onClick={onCancelRemoveData}
        />
      ) : (
        <AdminActionButton
          theme={themes.destructive}
          className={cn(styles.actionCellActionButton)}
          iconProps={{ iconName: "Delete" }}
          text={
            <FormattedMessage id="UserDetailsAdminActions.remove-user.action.remove" />
          }
          onClick={onRemoveData}
        />
      )}
      <div className={cn(styles.actionCellSpacer)} />
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
        <Label>
          <Text variant="xLarge">
            <FormattedMessage id="UserDetailsAdminActions.title" />
          </Text>
        </Label>
        <div className="-mt-3">
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
      </div>
    );
  };

export default UserDetailsAdminActions;
