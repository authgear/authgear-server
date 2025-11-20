import React, { useContext, useState, useCallback, useMemo } from "react";
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
  DatePicker,
  TimePicker,
  IComboBox,
  ITimeRange,
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
import { useAnonymizeUserMutation } from "./mutations/anonymizeUserMutation";
import { useScheduleAccountAnonymizationMutation } from "./mutations/scheduleAccountAnonymization";
import { useUnscheduleAccountAnonymizationMutation } from "./mutations/unscheduleAccountAnonymization";
import { useDeleteUserMutation } from "./mutations/deleteUserMutation";
import { useScheduleAccountDeletionMutation } from "./mutations/scheduleAccountDeletion";
import { useUnscheduleAccountDeletionMutation } from "./mutations/unscheduleAccountDeletion";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import styles from "./UserDetailsAccountStatus.module.css";

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

function usePickedDateAndTime(opts: {
  pickedDate: Date;
  pickedTime: Date;
}): Date {
  const { pickedDate, pickedTime } = opts;
  return useMemo(() => {
    return DateTime.fromJSDate(pickedDate)
      .set({
        hour: pickedTime.getHours(),
        minute: pickedTime.getMinutes(),
        second: 0,
        millisecond: 0,
      })
      .toJSDate();
  }, [pickedDate, pickedTime]);
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

function useDatePickerProps(opts: {
  minDate: Date;
  pickedTime: Date;
  setDate: (date: Date) => void;
  setTime: (date: Date) => void;
}): {
  onSelectDate: (date: Date | null | undefined) => void;
} {
  const { minDate, pickedTime, setDate, setTime } = opts;

  const onSelectDate = useCallback(
    (date: Date | null | undefined) => {
      if (date == null) {
        return;
      }

      const dateTime_minDate = DateTime.fromJSDate(minDate).startOf("day");
      const dateTime_pickedDate = DateTime.fromJSDate(date).startOf("day");

      // Do not allow to pick a date less than minDate.
      if (dateTime_pickedDate.valueOf() < dateTime_minDate.valueOf()) {
        return;
      }

      // Adjust the time.
      if (dateTime_pickedDate.valueOf() === dateTime_minDate.valueOf()) {
        setDate(date);
        if (
          pickedTime.getHours() < minDate.getHours() ||
          pickedTime.getMinutes() < minDate.getMinutes()
        ) {
          setTime(minDate);
        }
        return;
      }

      setDate(date);
    },
    [minDate, pickedTime, setDate, setTime]
  );

  return {
    onSelectDate,
  };
}

function useTimePickerTimeProps(opts: {
  minDate: Date;
  pickedDate: Date;
  setTime: (date: Date) => void;
}): {
  increments: number;
  timeRange: ITimeRange;
  onChange: (e: React.FormEvent<IComboBox>, time: Date) => void;
} {
  const increments = 60;
  const { minDate, pickedDate, setTime } = opts;
  const startOfDay_minDate = DateTime.fromJSDate(minDate).startOf("day");
  const startOfDay_pickedDate = DateTime.fromJSDate(pickedDate).startOf("day");

  const onChange = useCallback(
    (_e: React.FormEvent<IComboBox>, time: Date) => {
      setTime(time);
    },
    [setTime]
  );
  // This should not happen.
  if (startOfDay_pickedDate.valueOf() < startOfDay_minDate.valueOf()) {
    return {
      increments,
      onChange,
      timeRange: {
        start: 0,
        end: 0,
      },
    };
  }
  if (startOfDay_pickedDate.valueOf() > startOfDay_minDate.valueOf()) {
    return {
      increments,
      onChange,
      timeRange: {
        start: 0,
        end: 0,
      },
    };
  }
  return {
    increments,
    onChange,
    timeRange: {
      start: minDate.getHours(),
      end: 0,
    },
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
          <AccountValidPeriodCell data={data} />
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

  const [temporarilyDisabledUntil_date, setTemporarilyDisabledUntil_date] =
    useState(minDate);
  const [temporarilyDisabledUntil_time, setTemporarilyDisabledUntil_time] =
    useState(minDate);

  const datePickerProps = useDatePickerProps({
    minDate,
    pickedTime: temporarilyDisabledUntil_time,
    setDate: setTemporarilyDisabledUntil_date,
    setTime: setTemporarilyDisabledUntil_time,
  });
  const timePickerProps = useTimePickerTimeProps({
    minDate,
    pickedDate: temporarilyDisabledUntil_date,
    setTime: setTemporarilyDisabledUntil_time,
  });

  const temporarilyDisabledUntil = usePickedDateAndTime({
    pickedDate: temporarilyDisabledUntil_date,
    pickedTime: temporarilyDisabledUntil_time,
  });

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
            <div className="flex flex-row gap-2">
              <DatePicker
                {...datePickerProps}
                className="flex-1"
                disabled={disableChoiceGroupKey !== "temporarily"}
                value={temporarilyDisabledUntil_date}
              />
              <TimePicker
                {...timePickerProps}
                className="flex-1"
                disabled={disableChoiceGroupKey !== "temporarily"}
                allowFreeform={false}
                showSeconds={false}
                useHour12={false}
                value={temporarilyDisabledUntil_time}
              />
            </div>
          </div>
        </div>
      );
    },
    [
      datePickerProps,
      disableChoiceGroupKey,
      locale,
      temporarilyDisabledUntil_date,
      temporarilyDisabledUntil_time,
      timePickerProps,
    ]
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

  const {
    setDisabledStatus,
    loading: setDisabledStatusLoading,
    error: setDisabledStatusError,
  } = useSetDisabledStatusMutation();
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
    anonymizeUserLoading ||
    scheduleAccountAnonymizationLoading ||
    unscheduleAccountAnonymizationLoading ||
    deleteUserLoading ||
    scheduleAccountDeletionLoading ||
    unscheduleAccountDeletionLoading;
  const error =
    setDisabledStatusError ||
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
        break;
      case "edit-account-valid-period":
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
    disableForm,
    loading,
    mode,
    onClickAnonymize,
    onClickDelete,
    onClickDisable,
    onClickReenable,
    onClickScheduleAnonymization,
    onClickScheduleDeletion,
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

export default UserDetailsAccountStatus;
