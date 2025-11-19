import React, { useContext, useState, useCallback } from "react";
import { IStyle, Label, Text } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { useNavigate } from "react-router-dom";

import { useSystemConfig } from "../../context/SystemConfigContext";
import ListCellLayout from "../../ListCellLayout";
import OutlinedActionButton from "../../components/common/OutlinedActionButton";
import SetUserDisabledDialog from "./SetUserDisabledDialog";
import AnonymizeUserDialog from "./AnonymizeUserDialog";
import DeleteUserDialog from "./DeleteUserDialog";
import { formatDatetime } from "../../util/formatDatetime";
import styles from "./UserDetailsAccountStatus.module.css";

const labelTextStyle: IStyle = {
  lineHeight: "20px",
  fontWeight: 600,
};

const bodyTextStyle: IStyle = {
  lineHeight: "20px",
  maxWidth: "500px",
};

interface AccountStatus {
  id: string;
  isDisabled: boolean;
  isAnonymized: boolean;
  disableReason?: string | null;
  accountValidFrom?: string | null;
  accountValidUntil?: string | null;
  temporarilyDisabledUntil?: string | null;
  temporarilyDisabledFrom?: string | null;
  deleteAt?: string | null;
  anonymizeAt?: string | null;
  endUserAccountID?: string | null;
}

interface DisableUserCellProps {
  data: AccountStatus;
  onClickDisable: () => void;
}

interface AccountValidPeriodCellProps {
  data: AccountStatus;
}

interface AnonymizeUserCellProps {
  data: AccountStatus;
  onClickAnonymize: () => void;
}

interface RemoveUserCellProps {
  data: AccountStatus;
  onClickDelete: () => void;
}

interface ButtonStates {
  toggleDisable: {
    buttonDisabled: boolean;
    isDisabledIndefinitelyOrTemporarily: boolean;
    disableReason: string | null;
    temporarilyDisabledUntil: Date | null;
  };
  setAccountValidPeriod: {
    buttonDisabled: boolean;
    accountValidFrom: Date | null;
    accountValidUntil: Date | null;
  };
  anonymize: {
    buttonDisabled: boolean;
    isAnonymized: boolean;
    anonymizeAt: Date | null;
  };
  delete: {
    buttonDisabled: boolean;
    deleteAt: Date | null;
  };
}

export function getMostAppropriateAction(
  data: AccountStatus
):
  | "unschedule-deletion"
  | "unschedule-anonymization"
  | "re-enable"
  | "disable"
  | "no-action" {
  const buttonStates = getButtonStates(data);
  if (buttonStates.delete.deleteAt != null) {
    return "unschedule-deletion";
  }
  if (buttonStates.anonymize.isAnonymized) {
    return "no-action";
  }
  if (buttonStates.anonymize.anonymizeAt != null) {
    return "unschedule-anonymization";
  }
  if (buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily) {
    return "re-enable";
  }
  return "disable";
}

export function getButtonStates(data: AccountStatus): ButtonStates {
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
      buttonDisabled:
        data.isAnonymized ||
        outsideValidPeriod ||
        data.anonymizeAt != null ||
        data.deleteAt != null,
      isDisabledIndefinitelyOrTemporarily:
        temporarilyDisabled || indefinitelyDisabled,
      disableReason: data.disableReason ?? null,
      temporarilyDisabledUntil,
    },
    setAccountValidPeriod: {
      buttonDisabled: data.isAnonymized,
      accountValidFrom,
      accountValidUntil,
    },
    anonymize: {
      buttonDisabled: data.isAnonymized,
      isAnonymized: data.isAnonymized,
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
    const { locale } = useContext(Context);
    const { themes } = useSystemConfig();
    const { data, onClickDisable } = props;
    const buttonStates = getButtonStates(data);
    return (
      <ListCellLayout className={styles.actionCell}>
        <div className={styles.actionCellLabel}>
          <Text
            styles={{
              root: labelTextStyle,
            }}
          >
            <FormattedMessage id="UserDetailsAccountStatus.disable-user.title" />
          </Text>
        </div>
        <div className={styles.actionCellBody}>
          <Text
            styles={{
              root: bodyTextStyle,
            }}
          >
            <FormattedMessage id="UserDetailsAccountStatus.disable-user.body" />
          </Text>
        </div>
        {buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily &&
        (buttonStates.toggleDisable.disableReason != null ||
          buttonStates.toggleDisable.temporarilyDisabledUntil != null) ? (
          <div className={styles.actionCellDescription}>
            {buttonStates.toggleDisable.disableReason != null ? (
              <>
                <Text
                  styles={{
                    root: labelTextStyle,
                  }}
                >
                  {buttonStates.toggleDisable.disableReason}
                </Text>
              </>
            ) : null}

            {buttonStates.toggleDisable.temporarilyDisabledUntil != null ? (
              <>
                <Text
                  styles={{
                    root: labelTextStyle,
                  }}
                >
                  <FormattedMessage
                    id="UserDetailsAccountStatus.disable-user.until"
                    values={{
                      until:
                        formatDatetime(
                          locale,
                          buttonStates.toggleDisable.temporarilyDisabledUntil
                        ) ?? "",
                    }}
                  />
                </Text>
              </>
            ) : null}
          </div>
        ) : null}
        {buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily ? (
          <div className={styles.actionCellActionButtonContainer}>
            <OutlinedActionButton
              disabled={buttonStates.toggleDisable.buttonDisabled}
              theme={themes.actionButton}
              iconProps={{ iconName: "Play" }}
              text={
                <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.enable" />
              }
              onClick={onClickDisable}
            />
            <OutlinedActionButton
              disabled={buttonStates.toggleDisable.buttonDisabled}
              theme={themes.destructive}
              iconProps={{ iconName: "Calendar" }}
              text={
                <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.edit-schedule" />
              }
              onClick={onClickDisable}
            />
          </div>
        ) : (
          <OutlinedActionButton
            disabled={buttonStates.toggleDisable.buttonDisabled}
            theme={themes.destructive}
            className={styles.actionCellActionButton}
            iconProps={{ iconName: "Blocked" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.disable" />
            }
            onClick={onClickDisable}
          />
        )}
      </ListCellLayout>
    );
  };

const AccountValidPeriodCell: React.VFC<AccountValidPeriodCellProps> =
  function AccountValidPeriodCell(props) {
    const { locale } = useContext(Context);
    const { themes } = useSystemConfig();
    const { data } = props;
    const buttonStates = getButtonStates(data);
    return (
      <ListCellLayout className={styles.actionCell}>
        <div className={styles.actionCellLabel}>
          <Text
            styles={{
              root: labelTextStyle,
            }}
          >
            <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.title" />
          </Text>
        </div>
        <div className={styles.actionCellBody}>
          <Text
            styles={{
              root: bodyTextStyle,
            }}
          >
            {buttonStates.setAccountValidPeriod.accountValidFrom == null &&
            buttonStates.setAccountValidPeriod.accountValidUntil == null ? (
              <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.body--unset" />
            ) : (
              <>
                {buttonStates.setAccountValidPeriod.accountValidFrom != null ? (
                  <>
                    <FormattedMessage
                      id="UserDetailsAccountStatus.account-valid-period.start"
                      values={{
                        start:
                          formatDatetime(
                            locale,
                            buttonStates.setAccountValidPeriod.accountValidFrom
                          ) ?? "",
                      }}
                    />
                    <br />
                  </>
                ) : null}
                {buttonStates.setAccountValidPeriod.accountValidUntil !=
                null ? (
                  <>
                    <FormattedMessage
                      id="UserDetailsAccountStatus.account-valid-period.end"
                      values={{
                        end:
                          formatDatetime(
                            locale,
                            buttonStates.setAccountValidPeriod.accountValidUntil
                          ) ?? "",
                      }}
                    />
                    <br />
                  </>
                ) : null}
                <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.body--set" />
              </>
            )}
          </Text>
        </div>
        <OutlinedActionButton
          disabled={buttonStates.setAccountValidPeriod.buttonDisabled}
          theme={themes.destructive}
          className={styles.actionCellActionButton}
          iconProps={{ iconName: "Calendar" }}
          text={
            buttonStates.setAccountValidPeriod.accountValidFrom == null &&
            buttonStates.setAccountValidPeriod.accountValidUntil == null ? (
              <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.action.set" />
            ) : (
              <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.action.edit" />
            )
          }
        />
      </ListCellLayout>
    );
  };

const AnonymizeUserCell: React.VFC<AnonymizeUserCellProps> =
  function AnonymizeUserCell(props) {
    const { themes } = useSystemConfig();
    const { data, onClickAnonymize } = props;
    const buttonStates = getButtonStates(data);
    return (
      <ListCellLayout className={styles.actionCell}>
        <Text
          className={styles.actionCellLabel}
          styles={{
            root: labelTextStyle,
          }}
        >
          <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.title" />
        </Text>
        <Text
          className={styles.actionCellBody}
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
          className={styles.actionCellActionButton}
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
          onClick={onClickAnonymize}
        />
      </ListCellLayout>
    );
  };

const RemoveUserCell: React.VFC<RemoveUserCellProps> = function RemoveUserCell(
  props
) {
  const { themes } = useSystemConfig();
  const { data, onClickDelete } = props;
  const buttonStates = getButtonStates(data);
  return (
    <ListCellLayout className={styles.actionCell}>
      <Text
        className={styles.actionCellLabel}
        styles={{
          root: labelTextStyle,
        }}
      >
        <FormattedMessage id="UserDetailsAccountStatus.remove-user.title" />
      </Text>
      <Text
        className={styles.actionCellBody}
        styles={{
          root: bodyTextStyle,
        }}
      >
        <FormattedMessage id="UserDetailsAccountStatus.remove-user.body" />
      </Text>
      {buttonStates.delete.deleteAt != null ? (
        <div className={styles.actionCellActionButtonContainer}>
          <OutlinedActionButton
            disabled={buttonStates.delete.buttonDisabled}
            theme={themes.actionButton}
            className={styles.actionCellActionButton}
            iconProps={{ iconName: "Undo" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.cancel" />
            }
            onClick={onClickDelete}
          />
          <OutlinedActionButton
            disabled={buttonStates.delete.buttonDisabled}
            theme={themes.destructive}
            className={styles.actionCellActionButton}
            iconProps={{ iconName: "Delete" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove-now" />
            }
            onClick={onClickDelete}
          />
        </div>
      ) : (
        <OutlinedActionButton
          disabled={buttonStates.delete.buttonDisabled}
          theme={themes.destructive}
          className={styles.actionCellActionButton}
          iconProps={{ iconName: "Delete" }}
          text={
            <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove" />
          }
          onClick={onClickDelete}
        />
      )}
    </ListCellLayout>
  );
};

interface UserDetailsAccountStatusProps {
  data: AccountStatus;
}

const UserDetailsAccountStatus: React.VFC<UserDetailsAccountStatusProps> =
  function UserDetailsAccountStatus(props) {
    const { data } = props;
    const buttonStates = getButtonStates(data);
    const navigate = useNavigate();

    const [userDisabledDialogIsHidden, setUserDisabledDialogIsHidden] =
      useState(true);
    const [anonymizeUserDialogIsHidden, setAnonymizeUserDialogIsHidden] =
      useState(true);
    const [deleteUserDialogIsHidden, setDeleteUserDialogIsHidden] =
      useState(true);

    const onDismissSetUserDisabledDialog = useCallback(() => {
      setUserDisabledDialogIsHidden(true);
    }, [setUserDisabledDialogIsHidden]);
    const onDismissAnonymizeUserDialog = useCallback(() => {
      setAnonymizeUserDialogIsHidden(true);
    }, []);
    const onDismissDeleteUserDialog = useCallback(
      (deletedUser: boolean) => {
        setDeleteUserDialogIsHidden(true);
        if (deletedUser) {
          setTimeout(() => navigate("./../.."), 0);
        }
      },
      [navigate]
    );

    const onClickDisable = useCallback(() => {
      setUserDisabledDialogIsHidden(false);
    }, []);
    const onClickAnonymize = useCallback(() => {
      setAnonymizeUserDialogIsHidden(false);
    }, []);
    const onClickDelete = useCallback(() => {
      setDeleteUserDialogIsHidden(false);
    }, []);

    return (
      <div>
        <Label>
          <Text variant="xLarge">
            <FormattedMessage id="UserDetailsAccountStatus.title" />
          </Text>
        </Label>
        <div className="-mt-3">
          <DisableUserCell data={data} onClickDisable={onClickDisable} />
          <AccountValidPeriodCell data={data} />
          <AnonymizeUserCell data={data} onClickAnonymize={onClickAnonymize} />
          <RemoveUserCell data={data} onClickDelete={onClickDelete} />
        </div>
        <DeleteUserDialog
          isHidden={deleteUserDialogIsHidden}
          onDismiss={onDismissDeleteUserDialog}
          userID={data.id}
          userDeleteAt={data.deleteAt ?? null}
          endUserAccountIdentifier={data.endUserAccountID ?? undefined}
        />
        <AnonymizeUserDialog
          isHidden={anonymizeUserDialogIsHidden}
          onDismiss={onDismissAnonymizeUserDialog}
          userID={data.id}
          userAnonymizeAt={data.anonymizeAt ?? null}
          endUserAccountIdentifier={data.endUserAccountID ?? undefined}
        />
        <SetUserDisabledDialog
          isHidden={userDisabledDialogIsHidden}
          onDismiss={onDismissSetUserDisabledDialog}
          userID={data.id}
          userIsDisabled={
            buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily
          }
          endUserAccountIdentifier={data.endUserAccountID ?? undefined}
        />
      </div>
    );
  };

export default UserDetailsAccountStatus;
