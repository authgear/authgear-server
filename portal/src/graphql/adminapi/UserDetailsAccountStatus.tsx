import { IStyle, Label, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React from "react";
import cn from "classnames";
import styles from "./UserDetailsAccountStatus.module.css";
import ListCellLayout from "../../ListCellLayout";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import OutlinedActionButton from "../../components/common/OutlinedActionButton";

const labelTextStyle: IStyle = {
  lineHeight: "20px",
  fontWeight: 600,
};

const bodyTextStyle: IStyle = {
  lineHeight: "20px",
  maxWidth: "500px",
};

interface DisableUserCellProps {
  data: UserQueryNodeFragment;
}

interface AnonymizeUserCellProps {
  data: UserQueryNodeFragment;
}

interface RemoveUserCellProps {
  data: UserQueryNodeFragment;
}

interface ButtonStates {
  toggleDisable: {
    buttonDisabled: boolean;
    isDisabledIndefinitelyOrTemporarily: boolean;
  };
  setAccountValidPeriod: {
    buttonDisabled: boolean;
    accountValidFrom: Date | null;
    accountValidUntil: Date | null;
  };
  anonymize: {
    buttonDisabled: boolean;
    anonymizeAt: Date | null;
  };
  delete: {
    buttonDisabled: boolean;
    deleteAt: Date | null;
  };
}

function useButtonStates(data: UserQueryNodeFragment): ButtonStates {
  const now = new Date();

  const accountValidFrom =
    data.accountValidFrom != null ? new Date(data.accountValidFrom) : null;
  const accountValidUntil =
    data.accountValidUntil != null ? new Date(data.accountValidUntil) : null;
  const outsideValidPeriod =
    accountValidFrom != null
      ? now.getTime() < accountValidFrom.getTime()
      : accountValidUntil != null
      ? now.getTime() >= accountValidUntil.getTime()
      : false;
  const insideValidPeriod = !outsideValidPeriod;

  const temporarilyDisabledFrom =
    data.temporarilyDisabledFrom != null
      ? new Date(data.temporarilyDisabledFrom)
      : null;
  const temporarilyDisabledUntil =
    data.temporarilyDisabledUntil != null
      ? new Date(data.temporarilyDisabledUntil)
      : null;
  const temporarilyDisabled =
    temporarilyDisabledFrom != null &&
    temporarilyDisabledUntil != null &&
    now.getTime() >= temporarilyDisabledFrom.getTime() &&
    now.getTime() < temporarilyDisabledUntil.getTime();

  const indefinitelyDisabled =
    data.isDisabled &&
    data.deleteAt == null &&
    data.anonymizeAt == null &&
    insideValidPeriod &&
    !temporarilyDisabled;

  return {
    toggleDisable: {
      buttonDisabled: data.isAnonymized || outsideValidPeriod,
      isDisabledIndefinitelyOrTemporarily:
        temporarilyDisabled || indefinitelyDisabled,
    },
    setAccountValidPeriod: {
      buttonDisabled: data.isAnonymized,
      accountValidFrom,
      accountValidUntil,
    },
    anonymize: {
      buttonDisabled: data.isAnonymized,
      anonymizeAt: data.anonymizeAt != null ? new Date(data.anonymizeAt) : null,
    },
    delete: {
      buttonDisabled: false,
      deleteAt: data.deleteAt != null ? new Date(data.deleteAt) : null,
    },
  };
}

const DisableUserCell: React.VFC<DisableUserCellProps> =
  function DisableUserCell(props) {
    const { themes } = useSystemConfig();
    const { data } = props;
    const buttonStates = useButtonStates(data);
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
          <FormattedMessage id="UserDetailsAccountStatus.disable-user.title" />
        </Text>
        <Text
          className={cn(styles.actionCellBody)}
          styles={{
            root: bodyTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAccountStatus.disable-user.body" />
        </Text>
        <OutlinedActionButton
          disabled={buttonStates.toggleDisable.buttonDisabled}
          theme={
            buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily
              ? themes.actionButton
              : themes.destructive
          }
          className={cn(styles.actionCellActionButton)}
          iconProps={
            buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily
              ? { iconName: "Play" }
              : { iconName: "Blocked" }
          }
          text={
            buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily ? (
              <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.enable" />
            ) : (
              <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.disable" />
            )
          }
        />
        <div className={cn(styles.actionCellSpacer)} />
      </ListCellLayout>
    );
  };

const AnonymizeUserCell: React.VFC<AnonymizeUserCellProps> =
  function AnonymizeUserCell(props) {
    const { themes } = useSystemConfig();
    const { data } = props;
    const buttonStates = useButtonStates(data);
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
          <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.title" />
        </Text>
        <Text
          className={cn(styles.actionCellBody)}
          styles={{
            root: bodyTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.body" />
        </Text>
        <OutlinedActionButton
          disabled={buttonStates.anonymize.buttonDisabled}
          theme={
            buttonStates.anonymize.anonymizeAt != null
              ? themes.actionButton
              : themes.destructive
          }
          className={cn(styles.actionCellActionButton)}
          iconProps={
            buttonStates.anonymize.anonymizeAt != null
              ? { iconName: "Undo" }
              : { iconName: "Archive" }
          }
          text={
            buttonStates.anonymize.anonymizeAt != null ? (
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.cancel" />
            ) : (
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize" />
            )
          }
        />
        <div className={cn(styles.actionCellSpacer)} />
      </ListCellLayout>
    );
  };

const RemoveUserCell: React.VFC<RemoveUserCellProps> = function RemoveUserCell(
  props
) {
  const { themes } = useSystemConfig();
  const { data } = props;
  const buttonStates = useButtonStates(data);
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
        <FormattedMessage id="UserDetailsAccountStatus.remove-user.title" />
      </Text>
      <Text
        className={cn(styles.actionCellBody)}
        styles={{
          root: bodyTextStyle,
        }}
      >
        <FormattedMessage id="UserDetailsAccountStatus.remove-user.body" />
      </Text>
      {buttonStates.delete.deleteAt != null ? (
        <div className={cn(styles.actionCellActionButtonContainer)}>
          <OutlinedActionButton
            disabled={buttonStates.delete.buttonDisabled}
            theme={themes.actionButton}
            className={cn(styles.actionCellActionButton)}
            iconProps={{ iconName: "Undo" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.cancel" />
            }
          />
          <OutlinedActionButton
            disabled={buttonStates.delete.buttonDisabled}
            theme={themes.destructive}
            className={cn(styles.actionCellActionButton)}
            iconProps={{ iconName: "Delete" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove-now" />
            }
          />
        </div>
      ) : (
        <OutlinedActionButton
          disabled={buttonStates.delete.buttonDisabled}
          theme={themes.destructive}
          className={cn(styles.actionCellActionButton)}
          iconProps={{ iconName: "Delete" }}
          text={
            <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove" />
          }
        />
      )}
      <div className={cn(styles.actionCellSpacer)} />
    </ListCellLayout>
  );
};

interface UserDetailsAccountStatusProps {
  data: UserQueryNodeFragment;
}

const UserDetailsAccountStatus: React.VFC<UserDetailsAccountStatusProps> =
  function UserDetailsAccountStatus(props) {
    const { data } = props;
    return (
      <div>
        <Label>
          <Text variant="xLarge">
            <FormattedMessage id="UserDetailsAccountStatus.title" />
          </Text>
        </Label>
        <div className="-mt-3">
          <DisableUserCell data={data} />
          <AnonymizeUserCell data={data} />
          <RemoveUserCell data={data} />
        </div>
      </div>
    );
  };

export default UserDetailsAccountStatus;
