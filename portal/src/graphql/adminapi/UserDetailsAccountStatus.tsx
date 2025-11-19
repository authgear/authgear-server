import { IStyle, Label, Text } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import React, { useContext } from "react";
import styles from "./UserDetailsAccountStatus.module.css";
import ListCellLayout from "../../ListCellLayout";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import OutlinedActionButton from "../../components/common/OutlinedActionButton";
import { formatDatetime } from "../../util/formatDatetime";

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

interface AccountValidPeriodCellProps {
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
    const { data } = props;
    const buttonStates = useButtonStates(data);
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
            />
            <OutlinedActionButton
              disabled={buttonStates.toggleDisable.buttonDisabled}
              theme={themes.destructive}
              iconProps={{ iconName: "Calendar" }}
              text={
                <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.edit-schedule" />
              }
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
    const buttonStates = useButtonStates(data);
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
    const { data } = props;
    const buttonStates = useButtonStates(data);
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
        />
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
          />
          <OutlinedActionButton
            disabled={buttonStates.delete.buttonDisabled}
            theme={themes.destructive}
            className={styles.actionCellActionButton}
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
          className={styles.actionCellActionButton}
          iconProps={{ iconName: "Delete" }}
          text={
            <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove" />
          }
        />
      )}
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
          <AccountValidPeriodCell data={data} />
          <AnonymizeUserCell data={data} />
          <RemoveUserCell data={data} />
        </div>
      </div>
    );
  };

export default UserDetailsAccountStatus;
