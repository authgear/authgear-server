import React, { useContext, useState, useCallback, useMemo } from "react";
import cn from "classnames";
import {
  IStyle,
  Label,
  Text,
  Dialog,
  DialogFooter,
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  MessageBar,
  MessageBarType,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { useNavigate } from "react-router-dom";
import { DateTime, SystemZone } from "luxon";

import { useSystemConfig } from "../../context/SystemConfigContext";
import ListCellLayout from "../../ListCellLayout";
import OutlinedActionButton from "../../components/common/OutlinedActionButton";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import TextField from "../../TextField";
import ErrorDialog from "../../error/ErrorDialog";
import { useSetDisabledStatusMutation } from "./mutations/setDisabledStatusMutation";
import { useSetAccountValidPeriodMutation } from "./mutations/setAccountValidPeriodMutation";
import { useAnonymizeUserMutation } from "./mutations/anonymizeUserMutation";
import { useScheduleAccountAnonymizationMutation } from "./mutations/scheduleAccountAnonymization";
import { useUnscheduleAccountAnonymizationMutation } from "./mutations/unscheduleAccountAnonymization";
import { useDeleteUserMutation } from "./mutations/deleteUserMutation";
import { useScheduleAccountDeletionMutation } from "./mutations/scheduleAccountDeletion";
import { useUnscheduleAccountDeletionMutation } from "./mutations/unscheduleAccountDeletion";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import styles from "./UserDetailsAccountStatus.module.css";
import DateTimePicker from "../../DateTimePicker";

const labelTextStyle: IStyle = {
  lineHeight: "20px",
  fontWeight: 600,
};

const bodyTextStyle: IStyle = {
  lineHeight: "20px",
  maxWidth: "500px",
};

const choiceGroupStyle = {
  flexContainer: {
    selectors: {
      ".ms-ChoiceField": {
        display: "block",
      },
    },
  },
};

const dialogStyles = { main: { minHeight: 0 } };

export interface AccountStatus {
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
  onClickReenable: () => void;
}

interface AccountValidPeriodCellProps {
  data: AccountStatus;
  onClickSetAccountValidPeriod: () => void;
  onClickEditAccountValidPeriod: () => void;
}

interface AnonymizeUserCellProps {
  data: AccountStatus;
  onClickAnonymizeOrSchedule: () => void;
  onClickCancelAnonymization: () => void;
  onClickAnonymizeImmediately: () => void;
}

interface RemoveUserCellProps {
  data: AccountStatus;
  onClickDeleteOrSchedule: () => void;
  onClickCancelDeletion: () => void;
  onClickDeleteImmediately: () => void;
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

function formatSystemZone(now: Date, locale: string): string {
  const zone = new SystemZone();
  return `${zone.offsetName(now.getTime(), {
    format: "short",
    locale,
  })} (${zone.name})`;
}

function useMinDate(ref: Date): Date {
  // Add 1 hour so that the minDate is never less than now.
  return DateTime.fromJSDate(ref)
    .plus({
      hour: 1,
    })
    .set({
      minute: 0,
      second: 0,
      millisecond: 0,
    })
    .toJSDate();
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

function getButtonStates(data: AccountStatus): ButtonStates {
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

export interface AccountValidPeriodFormProps {
  className?: string;
  accountValidFrom: Date | null;
  onPickAccountValidFrom: (date: Date | null) => void;
  accountValidUntil: Date | null;
  onPickAccountValidUntil: (date: Date | null) => void;
}

export function AccountValidPeriodForm(
  props: AccountValidPeriodFormProps
): React.ReactElement {
  const {
    className,
    accountValidFrom,
    accountValidUntil,
    onPickAccountValidFrom,
    onPickAccountValidUntil,
  } = props;

  const { themes } = useSystemConfig();
  const { locale } = useContext(Context);
  const formattedZone = formatSystemZone(new Date(), locale);
  return (
    <div className={cn(className, "flex flex-col gap-2")}>
      <MessageBar
        messageBarType={MessageBarType.info}
        styles={{
          iconContainer: {
            display: "none",
          },
        }}
      >
        <FormattedMessage
          id="AccountValidPeriodForm.timezone-description"
          values={{
            timezone: formattedZone,
          }}
        />
      </MessageBar>
      <DateTimePicker
        pickedDateTime={accountValidFrom}
        minDateTime={null}
        onPickDateTime={onPickAccountValidFrom}
        showClearButton={true}
        label={
          <Label>
            <FormattedMessage id="AccountValidPeriodForm.start-at" />
          </Label>
        }
        hint={
          <Text
            variant="small"
            styles={{
              root: {
                color: themes.main.semanticColors.bodySubtext,
              },
            }}
          >
            <FormattedMessage id="AccountValidPeriodForm.hint" />
          </Text>
        }
      />
      <DateTimePicker
        pickedDateTime={accountValidUntil}
        minDateTime={null}
        onPickDateTime={onPickAccountValidUntil}
        showClearButton={true}
        label={
          <Label>
            <FormattedMessage id="AccountValidPeriodForm.end-at" />
          </Label>
        }
        hint={
          <Text
            variant="small"
            styles={{
              root: {
                color: themes.main.semanticColors.bodySubtext,
              },
            }}
          >
            <FormattedMessage id="AccountValidPeriodForm.hint" />
          </Text>
        }
      />
    </div>
  );
}

const DisableUserCell: React.VFC<DisableUserCellProps> =
  function DisableUserCell(props) {
    const { locale } = useContext(Context);
    const { themes } = useSystemConfig();
    const { data, onClickDisable, onClickReenable } = props;
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
              onClick={onClickReenable}
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
    const {
      data,
      onClickSetAccountValidPeriod,
      onClickEditAccountValidPeriod,
    } = props;
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
          onClick={
            buttonStates.setAccountValidPeriod.accountValidFrom == null &&
            buttonStates.setAccountValidPeriod.accountValidUntil == null
              ? onClickSetAccountValidPeriod
              : onClickEditAccountValidPeriod
          }
        />
      </ListCellLayout>
    );
  };

const AnonymizeUserCell: React.VFC<AnonymizeUserCellProps> =
  function AnonymizeUserCell(props) {
    const { themes } = useSystemConfig();
    const {
      data,
      onClickAnonymizeOrSchedule,
      onClickCancelAnonymization,
      onClickAnonymizeImmediately,
    } = props;
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
        {buttonStates.anonymize.anonymizeAt != null ? (
          <div className={styles.actionCellActionButtonContainer}>
            <OutlinedActionButton
              disabled={buttonStates.anonymize.buttonDisabled}
              theme={themes.actionButton}
              iconProps={{ iconName: "Undo" }}
              text={
                <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.cancel" />
              }
              onClick={onClickCancelAnonymization}
            />
            <OutlinedActionButton
              disabled={buttonStates.anonymize.buttonDisabled}
              theme={themes.destructive}
              iconProps={{ iconName: "Archive" }}
              text={
                <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize-now" />
              }
              onClick={onClickAnonymizeImmediately}
            />
          </div>
        ) : (
          <OutlinedActionButton
            disabled={buttonStates.anonymize.buttonDisabled}
            theme={themes.destructive}
            className={styles.actionCellActionButton}
            iconProps={{ iconName: "Archive" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize" />
            }
            onClick={onClickAnonymizeOrSchedule}
          />
        )}
      </ListCellLayout>
    );
  };

const RemoveUserCell: React.VFC<RemoveUserCellProps> = function RemoveUserCell(
  props
) {
  const { themes } = useSystemConfig();
  const {
    data,
    onClickCancelDeletion,
    onClickDeleteOrSchedule,
    onClickDeleteImmediately,
  } = props;
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
            iconProps={{ iconName: "Undo" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.cancel" />
            }
            onClick={onClickCancelDeletion}
          />
          <OutlinedActionButton
            disabled={buttonStates.delete.buttonDisabled}
            theme={themes.destructive}
            iconProps={{ iconName: "Delete" }}
            text={
              <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove-now" />
            }
            onClick={onClickDeleteImmediately}
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
          onClick={onClickDeleteOrSchedule}
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
    const navigate = useNavigate();

    const [dialogHidden, setDialogHidden] = useState(true);
    // Mount a new dialog on every open of the dialog.
    const [dialogKey, setDialogKey] = useState(0);
    const [mode, setMode] = useState<AccountStatusDialogProps["mode"]>("auto");

    const onClickDisable = useCallback(() => {
      setMode("disable");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickReenable = useCallback(() => {
      setMode("re-enable");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickSetAccountValidPeriod = useCallback(() => {
      setMode("set-account-valid-period");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickEditAccountValidPeriod = useCallback(() => {
      setMode("edit-account-valid-period");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickAnonymizeOrSchedule = useCallback(() => {
      setMode("anonymize-or-schedule");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickCancelAnonymization = useCallback(() => {
      setMode("cancel-anonymization");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickAnonymizeImmediately = useCallback(() => {
      setMode("anonymize-immediately");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickDeleteOrSchedule = useCallback(() => {
      setMode("delete-or-schedule");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickDeleteImmediately = useCallback(() => {
      setMode("delete-immediately");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);
    const onClickCancelDeletion = useCallback(() => {
      setMode("cancel-deletion");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);

    const onDismiss: AccountStatusDialogProps["onDismiss"] = useCallback(
      (info) => {
        setDialogHidden(true);
        if (info.deletedUser) {
          setTimeout(() => navigate("./../.."), 0);
        }
      },
      [navigate]
    );

    return (
      <div>
        <Label>
          <Text variant="xLarge">
            <FormattedMessage id="UserDetailsAccountStatus.title" />
          </Text>
        </Label>
        <div className="-mt-3">
          <DisableUserCell
            data={data}
            onClickDisable={onClickDisable}
            onClickReenable={onClickReenable}
          />
          <AccountValidPeriodCell
            data={data}
            onClickSetAccountValidPeriod={onClickSetAccountValidPeriod}
            onClickEditAccountValidPeriod={onClickEditAccountValidPeriod}
          />
          <AnonymizeUserCell
            data={data}
            onClickAnonymizeImmediately={onClickAnonymizeImmediately}
            onClickAnonymizeOrSchedule={onClickAnonymizeOrSchedule}
            onClickCancelAnonymization={onClickCancelAnonymization}
          />
          <RemoveUserCell
            data={data}
            onClickCancelDeletion={onClickCancelDeletion}
            onClickDeleteImmediately={onClickDeleteImmediately}
            onClickDeleteOrSchedule={onClickDeleteOrSchedule}
          />
        </div>
        <AccountStatusDialog
          key={String(dialogKey)}
          accountStatus={data}
          isHidden={dialogHidden}
          mode={mode}
          onDismiss={onDismiss}
        />
      </div>
    );
  };

export interface AccountStatusDialogProps {
  isHidden: boolean;
  onDismiss: (info: { deletedUser: boolean }) => void;
  mode:
    | "disable"
    | "re-enable"
    | "set-account-valid-period"
    | "edit-account-valid-period"
    | "anonymize-or-schedule"
    | "cancel-anonymization"
    | "anonymize-immediately"
    | "delete-or-schedule"
    | "cancel-deletion"
    | "delete-immediately"
    | "auto";
  accountStatus: AccountStatus;
}

export function AccountStatusDialog(
  props: AccountStatusDialogProps
): React.ReactElement {
  const { isHidden, onDismiss, mode, accountStatus } = props;
  const { themes } = useSystemConfig();
  const { locale, renderToString } = useContext(Context);

  const [mountedAt] = useState(() => new Date());
  const minDate = useMinDate(mountedAt);

  const [disableChoiceGroupKey, setDisableChoiceGroupKey] = useState<
    "indefinitely" | "temporarily"
  >("indefinitely");

  const [temporarilyDisabledUntil, setTemporarilyDisabledUntil] = useState(() =>
    DateTime.fromJSDate(minDate).plus({ days: 7 }).toJSDate()
  );

  const [disableReason, setDisableReason] = useState(
    () => accountStatus.disableReason ?? ""
  );

  const sanitizedDisableReason = useMemo(() => {
    const trimmed = disableReason.trim();
    if (trimmed === "") {
      return null;
    }
    return trimmed;
  }, [disableReason]);

  const [accountValidFrom, setAccountValidFrom] = useState<Date | null>(() => {
    if (accountStatus.accountValidFrom != null) {
      return new Date(accountStatus.accountValidFrom);
    }
    return null;
  });
  const [accountValidUntil, setAccountValidUntil] = useState<Date | null>(
    () => {
      if (accountStatus.accountValidUntil != null) {
        return new Date(accountStatus.accountValidUntil);
      }
      return null;
    }
  );

  const onPickAccountValidFrom = useCallback(
    (date: Date | null) => {
      if (date == null) {
        setAccountValidFrom(null);
      } else if (accountValidUntil == null) {
        setAccountValidFrom(date);
      } else if (date.getTime() > accountValidUntil.getTime()) {
        setAccountValidFrom(accountValidUntil);
        setAccountValidUntil(date);
      } else {
        setAccountValidFrom(date);
      }
    },
    [accountValidUntil]
  );
  const onPickAccountValidUntil = useCallback(
    (date: Date | null) => {
      if (date == null) {
        setAccountValidUntil(null);
      } else if (accountValidFrom == null) {
        setAccountValidUntil(date);
      } else if (date.getTime() < accountValidFrom.getTime()) {
        setAccountValidFrom(date);
        setAccountValidUntil(accountValidFrom);
      } else {
        setAccountValidUntil(date);
      }
    },
    [accountValidFrom]
  );

  const onRenderTemporarilyDisableFormField = useCallback(
    (
      props?: IChoiceGroupOption & IChoiceGroupOptionProps,
      render?: (
        props?: IChoiceGroupOption & IChoiceGroupOptionProps
      ) => JSX.Element | null
    ) => {
      const formattedZone = formatSystemZone(new Date(), locale);
      return (
        <div className="flex flex-col gap-2">
          {render?.(props)}
          {disableChoiceGroupKey === "temporarily" ? (
            <div className="flex flex-col ml-6 gap-2">
              <MessageBar
                messageBarType={MessageBarType.info}
                styles={{
                  iconContainer: {
                    display: "none",
                  },
                }}
              >
                <FormattedMessage
                  id="AccountStatusDialog.disable-user.timezone-description"
                  values={{
                    timezone: formattedZone,
                  }}
                />
              </MessageBar>
              <DateTimePicker
                minDateTime={minDate}
                pickedDateTime={temporarilyDisabledUntil}
                // @ts-expect-error
                onPickDateTime={setTemporarilyDisabledUntil}
                showClearButton={false}
              />
            </div>
          ) : null}
        </div>
      );
    },
    [disableChoiceGroupKey, locale, minDate, temporarilyDisabledUntil]
  );

  const disableChoiceGroupOptions: IChoiceGroupOption[] = useMemo(() => {
    return [
      {
        key: "indefinitely",
        text: renderToString(
          "AccountStatusDialog.disable-user.disable-period.options.indefinitely"
        ),
      },
      {
        key: "temporarily",
        text: renderToString(
          "AccountStatusDialog.disable-user.disable-period.options.temporarily"
        ),
        onRenderField: onRenderTemporarilyDisableFormField,
      },
    ];
  }, [onRenderTemporarilyDisableFormField, renderToString]);

  const onChangeDisableChoiceGroup = useCallback(
    (
      _?: React.FormEvent<HTMLElement | HTMLInputElement>,
      option?: IChoiceGroupOption
    ) => {
      if (!option?.key) return;
      setDisableChoiceGroupKey(option.key as any);
    },
    []
  );
  const onChangeDisableReason = useCallback(
    (
      _e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
      value?: string
    ) => {
      setDisableReason(value ?? "");
    },
    []
  );
  const disableForm = useMemo(() => {
    return (
      <div className="flex flex-col gap-4">
        <ChoiceGroup
          styles={choiceGroupStyle}
          // @ts-expect-error
          label={
            <FormattedMessage id="AccountStatusDialog.disable-user.disable-period.label" />
          }
          options={disableChoiceGroupOptions}
          selectedKey={disableChoiceGroupKey}
          onChange={onChangeDisableChoiceGroup}
        />
        <TextField
          // @ts-expect-error
          label={
            <FormattedMessage id="AccountStatusDialog.disable-user.disable-reason.label" />
          }
          placeholder={renderToString(
            "AccountStatusDialog.disable-user.disable-reason.placeholder"
          )}
          value={disableReason}
          onChange={onChangeDisableReason}
        />
      </div>
    );
  }, [
    disableChoiceGroupKey,
    disableChoiceGroupOptions,
    disableReason,
    onChangeDisableChoiceGroup,
    onChangeDisableReason,
    renderToString,
  ]);

  const accountValidPeriodForm = useMemo(() => {
    return (
      <AccountValidPeriodForm
        accountValidFrom={accountValidFrom}
        accountValidUntil={accountValidUntil}
        onPickAccountValidFrom={onPickAccountValidFrom}
        onPickAccountValidUntil={onPickAccountValidUntil}
      />
    );
  }, [
    accountValidFrom,
    accountValidUntil,
    onPickAccountValidFrom,
    onPickAccountValidUntil,
  ]);

  const {
    setDisabledStatus,
    loading: setDisabledStatusLoading,
    error: setDisabledStatusError,
  } = useSetDisabledStatusMutation();
  const {
    setAccountValidPeriod,
    loading: setAccountValidPeriodLoading,
    error: setAccountValidPeriodError,
  } = useSetAccountValidPeriodMutation();
  const {
    anonymizeUser,
    loading: anonymizeUserLoading,
    error: anonymizeUserError,
  } = useAnonymizeUserMutation();
  const {
    scheduleAccountAnonymization,
    loading: scheduleAccountAnonymizationLoading,
    error: scheduleAccountAnonymizationError,
  } = useScheduleAccountAnonymizationMutation();
  const {
    unscheduleAccountAnonymization,
    loading: unscheduleAccountAnonymizationLoading,
    error: unscheduleAccountAnonymizationError,
  } = useUnscheduleAccountAnonymizationMutation();
  const {
    deleteUser,
    loading: deleteUserLoading,
    error: deleteUserError,
  } = useDeleteUserMutation();
  const {
    scheduleAccountDeletion,
    loading: scheduleAccountDeletionLoading,
    error: scheduleAccountDeletionError,
  } = useScheduleAccountDeletionMutation();
  const {
    unscheduleAccountDeletion,
    loading: unscheduleAccountDeletionLoading,
    error: unscheduleAccountDeletionError,
  } = useUnscheduleAccountDeletionMutation();

  const loading =
    setDisabledStatusLoading ||
    setAccountValidPeriodLoading ||
    anonymizeUserLoading ||
    scheduleAccountAnonymizationLoading ||
    unscheduleAccountAnonymizationLoading ||
    deleteUserLoading ||
    scheduleAccountDeletionLoading ||
    unscheduleAccountDeletionLoading;
  const error =
    setDisabledStatusError ||
    setAccountValidPeriodError ||
    anonymizeUserError ||
    scheduleAccountAnonymizationError ||
    unscheduleAccountAnonymizationError ||
    deleteUserError ||
    scheduleAccountDeletionError ||
    unscheduleAccountDeletionError;

  const onDialogDismiss = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    onDismiss({ deletedUser: false });
  }, [loading, isHidden, onDismiss]);

  const onClickDisable = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    setDisabledStatus({
      userID: accountStatus.id,
      isDisabled: true,
      reason: sanitizedDisableReason,
      temporarilyDisabledFrom:
        disableChoiceGroupKey === "indefinitely" ? null : new Date(),
      temporarilyDisabledUntil:
        disableChoiceGroupKey === "indefinitely"
          ? null
          : temporarilyDisabledUntil,
    }).finally(() => onDismiss({ deletedUser: false }));
  }, [
    accountStatus.id,
    disableChoiceGroupKey,
    isHidden,
    loading,
    onDismiss,
    sanitizedDisableReason,
    setDisabledStatus,
    temporarilyDisabledUntil,
  ]);

  const onClickReenable = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    setDisabledStatus({
      userID: accountStatus.id,
      isDisabled: false,
      reason: null,
      temporarilyDisabledFrom: null,
      temporarilyDisabledUntil: null,
    }).finally(() => onDismiss({ deletedUser: false }));
  }, [accountStatus.id, isHidden, loading, onDismiss, setDisabledStatus]);

  const onClickSetAccountValidPeriod = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    setAccountValidPeriod({
      userID: accountStatus.id,
      accountValidFrom: accountValidFrom,
      accountValidUntil: accountValidUntil,
    }).finally(() => onDismiss({ deletedUser: false }));
  }, [
    accountStatus.id,
    accountValidFrom,
    accountValidUntil,
    isHidden,
    loading,
    onDismiss,
    setAccountValidPeriod,
  ]);

  const onClickAnonymize = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    anonymizeUser(accountStatus.id).finally(() =>
      onDismiss({ deletedUser: false })
    );
  }, [accountStatus.id, anonymizeUser, isHidden, loading, onDismiss]);

  const onClickScheduleAnonymization = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    scheduleAccountAnonymization(accountStatus.id).finally(() =>
      onDismiss({ deletedUser: false })
    );
  }, [
    accountStatus.id,
    isHidden,
    loading,
    onDismiss,
    scheduleAccountAnonymization,
  ]);

  const onClickUnscheduleAnonymization = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    unscheduleAccountAnonymization(accountStatus.id).finally(() =>
      onDismiss({ deletedUser: false })
    );
  }, [
    accountStatus.id,
    isHidden,
    loading,
    onDismiss,
    unscheduleAccountAnonymization,
  ]);

  const onClickDelete = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    deleteUser(accountStatus.id)
      .then(() => onDismiss({ deletedUser: true }))
      .catch(() => onDismiss({ deletedUser: false }));
  }, [accountStatus.id, deleteUser, isHidden, loading, onDismiss]);

  const onClickScheduleDeletion = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    scheduleAccountDeletion(accountStatus.id).finally(() =>
      onDismiss({ deletedUser: false })
    );
  }, [accountStatus.id, isHidden, loading, onDismiss, scheduleAccountDeletion]);

  const onClickUnscheduleDeletion = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    unscheduleAccountDeletion(accountStatus.id).finally(() =>
      onDismiss({ deletedUser: false })
    );
  }, [
    accountStatus.id,
    isHidden,
    loading,
    onDismiss,
    unscheduleAccountDeletion,
  ]);

  const dialogContentPropsAndDialogSlots: {
    dialogContentProps: {
      title: React.ReactElement | null;
      subText: React.ReactElement | null;
    };
    body: React.ReactElement | null;
    button1: React.ReactElement | null;
    button2: React.ReactElement | null;
  } = useMemo(() => {
    const args = {
      username:
        accountStatus.endUserAccountID ?? extractRawID(accountStatus.id),
    };

    let title: React.ReactElement | null = null;
    let subText: React.ReactElement | null = null;
    let body: React.ReactElement | null = null;
    let button1: React.ReactElement | null = null;
    let button2: React.ReactElement | null = null;

    const prepareUnscheduleDeletion = () => {
      title = (
        <FormattedMessage id="AccountStatusDialog.cancel-deletion.title" />
      );
      subText = (
        <FormattedMessage
          id="AccountStatusDialog.cancel-deletion.description"
          values={args}
        />
      );
      button1 = (
        <PrimaryButton
          theme={themes.main}
          disabled={loading}
          onClick={onClickUnscheduleDeletion}
          text={
            <FormattedMessage id="AccountStatusDialog.cancel-deletion.action.cancel-deletion" />
          }
        />
      );
    };
    const prepareUnscheduleAnonymization = () => {
      title = (
        <FormattedMessage id="AccountStatusDialog.cancel-anonymization.title" />
      );
      subText = (
        <FormattedMessage
          id="AccountStatusDialog.cancel-anonymization.description"
          values={args}
        />
      );
      button1 = (
        <PrimaryButton
          theme={themes.main}
          disabled={loading}
          onClick={onClickUnscheduleAnonymization}
          text={
            <FormattedMessage id="AccountStatusDialog.cancel-anonymization.action.cancel-anonymization" />
          }
        />
      );
    };
    const prepareReenable = () => {
      title = <FormattedMessage id="AccountStatusDialog.reenable-user.title" />;
      subText = (
        <FormattedMessage
          id="AccountStatusDialog.reenable-user.description"
          values={args}
        />
      );
      button1 = (
        <PrimaryButton
          theme={themes.main}
          disabled={loading}
          onClick={onClickReenable}
          text={
            <FormattedMessage id="AccountStatusDialog.reenable-user.action.reenable" />
          }
        />
      );
    };
    const prepareDisable = () => {
      title = <FormattedMessage id="AccountStatusDialog.disable-user.title" />;
      subText = (
        <FormattedMessage
          id="AccountStatusDialog.disable-user.description"
          values={args}
        />
      );
      body = disableForm;
      button1 = (
        <PrimaryButton
          theme={themes.destructive}
          disabled={loading}
          onClick={onClickDisable}
          text={
            <FormattedMessage id="AccountStatusDialog.disable-user.action.disable" />
          }
        />
      );
    };

    switch (mode) {
      case "disable":
        prepareDisable();
        break;
      case "re-enable":
        prepareReenable();
        break;
      case "set-account-valid-period":
        title = (
          <FormattedMessage id="AccountStatusDialog.account-valid-period.title--set" />
        );
        subText = (
          <FormattedMessage
            id="AccountStatusDialog.account-valid-period.description--set"
            values={args}
          />
        );
        body = accountValidPeriodForm;
        button1 = (
          <PrimaryButton
            theme={themes.main}
            disabled={loading}
            text={
              <FormattedMessage id="AccountStatusDialog.account-valid-period.action.save" />
            }
            onClick={onClickSetAccountValidPeriod}
          />
        );
        break;
      case "edit-account-valid-period":
        title = (
          <FormattedMessage id="AccountStatusDialog.account-valid-period.title--edit" />
        );
        subText = (
          <FormattedMessage
            id="AccountStatusDialog.account-valid-period.description--edit"
            values={args}
          />
        );
        body = accountValidPeriodForm;
        button1 = (
          <PrimaryButton
            theme={themes.main}
            disabled={loading}
            text={
              <FormattedMessage id="AccountStatusDialog.account-valid-period.action.edit" />
            }
            onClick={onClickSetAccountValidPeriod}
          />
        );
        break;
      case "anonymize-or-schedule":
        title = (
          <FormattedMessage id="AccountStatusDialog.anonymize-user.title" />
        );
        subText = (
          <FormattedMessage
            id="AccountStatusDialog.anonymize-user.description"
            values={args}
          />
        );
        button1 = (
          <PrimaryButton
            theme={themes.main}
            disabled={loading}
            onClick={onClickScheduleAnonymization}
            text={
              <FormattedMessage id="AccountStatusDialog.anonymize-user.action.schedule-anonymization" />
            }
          />
        );
        button2 = (
          <PrimaryButton
            theme={themes.destructive}
            disabled={loading}
            onClick={onClickAnonymize}
            text={
              <FormattedMessage id="AccountStatusDialog.anonymize-user.action.anonymize-immediately" />
            }
          />
        );
        break;
      case "cancel-anonymization":
        prepareUnscheduleAnonymization();
        break;
      case "anonymize-immediately":
        title = (
          <FormattedMessage id="AccountStatusDialog.anonymize-user.title" />
        );
        subText = (
          <FormattedMessage
            id="AccountStatusDialog.anonymize-user.description"
            values={args}
          />
        );
        button1 = (
          <PrimaryButton
            theme={themes.destructive}
            disabled={loading}
            onClick={onClickAnonymize}
            text={
              <FormattedMessage id="AccountStatusDialog.anonymize-user.action.anonymize-immediately" />
            }
          />
        );
        break;

      case "delete-or-schedule":
        title = <FormattedMessage id="AccountStatusDialog.delete-user.title" />;
        subText = (
          <FormattedMessage
            id="AccountStatusDialog.delete-user.description"
            values={args}
          />
        );
        button1 = (
          <PrimaryButton
            theme={themes.main}
            disabled={loading}
            onClick={onClickScheduleDeletion}
            text={
              <FormattedMessage id="AccountStatusDialog.delete-user.action.schedule-deletion" />
            }
          />
        );
        button2 = (
          <PrimaryButton
            theme={themes.destructive}
            disabled={loading}
            onClick={onClickDelete}
            text={
              <FormattedMessage id="AccountStatusDialog.delete-user.action.delete-immediately" />
            }
          />
        );
        break;
      case "cancel-deletion":
        prepareUnscheduleDeletion();
        break;

      case "delete-immediately":
        title = <FormattedMessage id="AccountStatusDialog.delete-user.title" />;
        subText = (
          <FormattedMessage
            id="AccountStatusDialog.delete-user.description"
            values={args}
          />
        );
        button1 = (
          <PrimaryButton
            theme={themes.destructive}
            disabled={loading}
            onClick={onClickDelete}
            text={
              <FormattedMessage id="AccountStatusDialog.delete-user.action.delete-immediately" />
            }
          />
        );
        break;
      case "auto": {
        const action = getMostAppropriateAction(accountStatus);
        switch (action) {
          case "unschedule-deletion":
            prepareUnscheduleDeletion();
            break;
          case "unschedule-anonymization":
            prepareUnscheduleAnonymization();
            break;
          case "re-enable":
            prepareReenable();
            break;
          case "disable":
            prepareDisable();
            break;
          case "no-action":
            break;
        }
        break;
      }
    }
    return { dialogContentProps: { title, subText }, body, button1, button2 };
  }, [
    accountStatus,
    accountValidPeriodForm,
    disableForm,
    loading,
    mode,
    onClickAnonymize,
    onClickDelete,
    onClickDisable,
    onClickReenable,
    onClickScheduleAnonymization,
    onClickScheduleDeletion,
    onClickSetAccountValidPeriod,
    onClickUnscheduleAnonymization,
    onClickUnscheduleDeletion,
    themes.destructive,
    themes.main,
  ]);

  return (
    <>
      <Dialog
        hidden={isHidden}
        onDismiss={onDialogDismiss}
        // @ts-expect-error
        dialogContentProps={dialogContentPropsAndDialogSlots.dialogContentProps}
        styles={dialogStyles}
        minWidth={560}
      >
        {dialogContentPropsAndDialogSlots.body}
        <DialogFooter>
          {dialogContentPropsAndDialogSlots.button1}
          {dialogContentPropsAndDialogSlots.button2}
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
}

export interface AccountStatusBadgeProps {
  className?: string;
  accountStatus: AccountStatus;
}

const warnBadgeStyle: IStyle = {
  padding: 4,
  borderRadius: 4,
  color: "#ffffff",
  backgroundColor: "#e23d3d",
};

export function AccountStatusBadge(
  props: AccountStatusBadgeProps
): React.ReactElement | null {
  const now = new Date();
  const { accountStatus, className } = props;
  const id =
    accountStatus.deleteAt != null
      ? "AccountStatusBadge.scheduled-deletion"
      : accountStatus.isAnonymized
      ? "AccountStatusBadge.anonymized"
      : accountStatus.anonymizeAt != null
      ? "AccountStatusBadge.scheduled-anonymization"
      : accountStatus.accountValidFrom != null &&
        now.getTime() < new Date(accountStatus.accountValidFrom).getTime()
      ? "AccountStatusBadge.account-outside-valid-period"
      : accountStatus.accountValidUntil != null &&
        now.getTime() >= new Date(accountStatus.accountValidUntil).getTime()
      ? "AccountStatusBadge.account-outside-valid-period"
      : accountStatus.isDisabled
      ? "AccountStatusBadge.disabled"
      : null;
  if (id == null) {
    return null;
  }

  return (
    <Text
      className={className}
      styles={{
        root: warnBadgeStyle,
      }}
    >
      <FormattedMessage id={id} />
    </Text>
  );
}

export interface AccountStatusMessageBarProps {
  accountStatus: AccountStatus;
}

export function AccountStatusMessageBar(
  props: AccountStatusMessageBarProps
): React.ReactElement | null {
  const now = new Date();
  const { accountStatus } = props;
  const { locale } = useContext(Context);

  let message = null;
  if (accountStatus.deleteAt != null) {
    message = (
      <FormattedMessage
        id="AccountStatusMessageBar.scheduled-deletion"
        values={{
          date: formatDatetime(locale, accountStatus.deleteAt) ?? "",
        }}
      />
    );
  } else if (accountStatus.isAnonymized) {
    message = <FormattedMessage id="AccountStatusMessageBar.anonymized" />;
  } else if (accountStatus.anonymizeAt != null) {
    message = (
      <FormattedMessage
        id="AccountStatusMessageBar.scheduled-anonymization"
        values={{
          date: formatDatetime(locale, accountStatus.anonymizeAt) ?? "",
        }}
      />
    );
  } else if (
    accountStatus.accountValidUntil != null &&
    now.getTime() >= new Date(accountStatus.accountValidUntil).getTime()
  ) {
    message = (
      <FormattedMessage
        id="AccountStatusMessageBar.account-valid-until"
        values={{
          date: formatDatetime(locale, accountStatus.accountValidUntil) ?? "",
        }}
      />
    );
  } else if (
    accountStatus.accountValidFrom != null &&
    now.getTime() < new Date(accountStatus.accountValidFrom).getTime()
  ) {
    message = (
      <FormattedMessage
        id="AccountStatusMessageBar.account-valid-from"
        values={{
          date: formatDatetime(locale, accountStatus.accountValidFrom) ?? "",
        }}
      />
    );
  } else if (
    accountStatus.isDisabled &&
    accountStatus.temporarilyDisabledUntil != null
  ) {
    message = (
      <FormattedMessage
        id="AccountStatusMessageBar.disabled-tempoararily"
        values={{
          date:
            formatDatetime(locale, accountStatus.temporarilyDisabledUntil) ??
            "",
        }}
      />
    );
  } else if (accountStatus.isDisabled) {
    message = (
      <FormattedMessage id="AccountStatusMessageBar.disabled-indefinitely" />
    );
  }

  if (message == null) {
    return null;
  }

  return (
    <MessageBar messageBarType={MessageBarType.warning}>{message}</MessageBar>
  );
}

export default UserDetailsAccountStatus;
