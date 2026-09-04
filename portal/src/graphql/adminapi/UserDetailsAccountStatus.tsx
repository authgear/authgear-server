import React, { useContext, useState, useCallback, useMemo } from "react";
import cn from "classnames";
import { Badge, Button, Dialog, RadioGroup, Text } from "@radix-ui/themes";
import { FormattedMessage, Context } from "../../intl";
import { useNavigate } from "react-router-dom";
import { DateTime, SystemZone } from "luxon";
import {
  CalendarIcon,
  CardStackMinusIcon,
  CircleBackslashIcon,
  TrashIcon,
} from "@radix-ui/react-icons";

import ListCellLayout from "../../ListCellLayout";
import ExternalLink from "../../ExternalLink";
import ErrorDialog from "../../error/ErrorDialog";
import { useSetDisabledStatusMutation } from "./mutations/setDisabledStatusMutation";
import { useSetAccountValidPeriodMutation } from "./mutations/setAccountValidPeriodMutation";
import { useAnonymizeUserMutation } from "./mutations/anonymizeUserMutation";
import { useScheduleAccountAnonymizationMutation } from "./mutations/scheduleAccountAnonymization";
import { useUnscheduleAccountAnonymizationMutation } from "./mutations/unscheduleAccountAnonymization";
import { useDeleteUserMutation } from "./mutations/deleteUserMutation";
import { useScheduleAccountDeletionMutation } from "./mutations/scheduleAccountDeletion";
import { useUnscheduleAccountDeletionMutation } from "./mutations/unscheduleAccountDeletion";
import { useResetAccountLockoutMutation } from "./mutations/resetAccountLockoutMutation";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import styles from "./UserDetailsAccountStatus.module.css";
import DateTimePicker from "../../DateTimePicker";
import {
  ErrorParseRule,
  makeInvalidAccountStatusTransitionErrorParseRule,
} from "../../error/parse";
import { Callout } from "../../components/v2/Callout/Callout";
import { TextField } from "../../components/v2/TextField/TextField";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";

function StatusActionButton({
  className,
  disabled,
  destructive = false,
  solid = false,
  onClick,
  children,
}: {
  className?: string;
  disabled?: boolean;
  destructive?: boolean;
  solid?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <Button
      className={cn(styles.statusActionButton, className)}
      size="2"
      variant={solid ? "solid" : "outline"}
      color={destructive ? "red" : "gray"}
      disabled={disabled}
      onClick={onClick}
    >
      {children}
    </Button>
  );
}

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
  accountLockout?: {
    isLocked: boolean;
    lockoutType: string;
    lockedUntil?: string | null;
    lockedIPs: Array<{ ipAddress: string; lockedUntil: string }>;
  } | null;
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

interface AccountLockoutCellProps {
  data: AccountStatus;
  onClickResetAccountLockout: () => void;
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

function getMostAppropriateAccountState(accountStatus: AccountStatus):
  | {
      state: "scheduled-deletion";
      deleteAt: Date;
    }
  | {
      state: "anonymized";
    }
  | {
      state: "scheduled-anonymization";
      anonymizeAt: Date;
    }
  | {
      state: "less-than-account-valid-from";
      accountValidFrom: Date;
    }
  | {
      state: "greater-than-or-equal-to-account-valid-until";
      accountValidUntil: Date;
    }
  | {
      state: "disabled";
      temporarilyDisabledUntil: Date | null;
    }
  | {
      state: "normal";
    } {
  const now = new Date();
  if (accountStatus.deleteAt != null) {
    return {
      state: "scheduled-deletion",
      deleteAt: new Date(accountStatus.deleteAt),
    };
  }
  if (accountStatus.isAnonymized) {
    return {
      state: "anonymized",
    };
  }
  if (accountStatus.anonymizeAt != null) {
    return {
      state: "scheduled-anonymization",
      anonymizeAt: new Date(accountStatus.anonymizeAt),
    };
  }
  if (
    accountStatus.accountValidFrom != null &&
    now.getTime() < new Date(accountStatus.accountValidFrom).getTime()
  ) {
    return {
      state: "less-than-account-valid-from",
      accountValidFrom: new Date(accountStatus.accountValidFrom),
    };
  }
  if (
    accountStatus.accountValidUntil != null &&
    now.getTime() >= new Date(accountStatus.accountValidUntil).getTime()
  ) {
    return {
      state: "greater-than-or-equal-to-account-valid-until",
      accountValidUntil: new Date(accountStatus.accountValidUntil),
    };
  }
  if (accountStatus.isDisabled) {
    return {
      state: "disabled",
      temporarilyDisabledUntil:
        accountStatus.temporarilyDisabledUntil == null
          ? null
          : new Date(accountStatus.temporarilyDisabledUntil),
    };
  }
  return {
    state: "normal",
  };
}

export function getMostAppropriateAction(
  accountStatus: AccountStatus
):
  | "unschedule-deletion"
  | "unschedule-anonymization"
  | "edit-account-valid-period"
  | "re-enable"
  | "disable"
  | "no-action" {
  const accountState = getMostAppropriateAccountState(accountStatus);
  switch (accountState.state) {
    case "scheduled-deletion":
      return "unschedule-deletion";
    case "scheduled-anonymization":
      return "unschedule-anonymization";
    case "less-than-account-valid-from":
      return "edit-account-valid-period";
    case "greater-than-or-equal-to-account-valid-until":
      return "edit-account-valid-period";
    case "disabled":
      return "re-enable";
    case "anonymized":
      return "no-action";
    case "normal":
      return "disable";
  }
}

function getButtonStates(data: AccountStatus): ButtonStates {
  const now = new Date();

  const accountValidFrom =
    data.accountValidFrom != null ? new Date(data.accountValidFrom) : null;
  const accountValidUntil =
    data.accountValidUntil != null ? new Date(data.accountValidUntil) : null;
  const outsideValidPeriod =
    (accountValidFrom != null && now.getTime() < accountValidFrom.getTime()) ||
    (accountValidUntil != null && now.getTime() >= accountValidUntil.getTime());
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
  calloutColor?: "gray";
  calloutSize?: "1" | "2" | "3";
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
    calloutColor,
    calloutSize = "2",
    accountValidFrom,
    accountValidUntil,
    onPickAccountValidFrom,
    onPickAccountValidUntil,
  } = props;

  const { locale } = useContext(Context);
  const formattedZone = formatSystemZone(new Date(), locale);
  const showEndAtWarning = useMemo(
    () =>
      accountValidUntil != null &&
      accountValidUntil.getTime() < new Date().getTime(),
    [accountValidUntil]
  );
  return (
    <div className={cn(className, "flex flex-col gap-2")}>
      <Callout
        type="info"
        color={calloutColor}
        size={calloutSize}
        showCloseButton={false}
        text={
          <FormattedMessage
            id="AccountValidPeriodForm.timezone-description"
            values={{
              timezone: formattedZone,
            }}
          />
        }
      />
      <DateTimePicker
        pickedDateTime={accountValidFrom}
        minDateTime={null}
        onPickDateTime={onPickAccountValidFrom}
        showClearButton={true}
        label={
          <Text as="label" size="2" weight="medium">
            <FormattedMessage id="AccountValidPeriodForm.start-at" />
          </Text>
        }
        hint={
          <Text size="1" color="gray">
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
          <Text as="label" size="2" weight="medium">
            <FormattedMessage id="AccountValidPeriodForm.end-at" />
          </Text>
        }
        hint={
          <Text size="1" color="gray">
            <FormattedMessage id="AccountValidPeriodForm.hint" />
          </Text>
        }
      />
      {showEndAtWarning ? (
        <Callout
          type="warning"
          color="yellow"
          size={calloutSize}
          showCloseButton={false}
          text={<FormattedMessage id="AccountValidPeriodForm.end-at-warning" />}
        />
      ) : null}
    </div>
  );
}

const DisableUserCell: React.VFC<DisableUserCellProps> =
  function DisableUserCell(props) {
    const { locale } = useContext(Context);
    const { data, onClickDisable, onClickReenable } = props;
    const buttonStates = getButtonStates(data);
    return (
      <ListCellLayout className={styles.actionCell}>
        <div className={styles.actionCellLabel}>
          <Text as="p" size="2" weight="medium" className={styles.cellLabel}>
            <FormattedMessage id="UserDetailsAccountStatus.disable-user.title" />
          </Text>
        </div>
        <div className={styles.actionCellBody}>
          <Text as="p" size="2" className={styles.cellBody}>
            <FormattedMessage id="UserDetailsAccountStatus.disable-user.body" />
          </Text>
        </div>
        {buttonStates.toggleDisable.isDisabledIndefinitelyOrTemporarily &&
        (buttonStates.toggleDisable.disableReason != null ||
          buttonStates.toggleDisable.temporarilyDisabledUntil != null) ? (
          <div className={styles.actionCellDescription}>
            {buttonStates.toggleDisable.disableReason != null ? (
              <>
                <Text as="p" size="2" className={styles.cellMeta}>
                  <FormattedMessage
                    id="UserDetailsAccountStatus.disable-user.reason"
                    values={{
                      reason: buttonStates.toggleDisable.disableReason,
                    }}
                  />
                </Text>
              </>
            ) : null}

            {buttonStates.toggleDisable.temporarilyDisabledUntil != null ? (
              <>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.cellLabel}
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
            <StatusActionButton
              disabled={buttonStates.toggleDisable.buttonDisabled}
              onClick={onClickReenable}
            >
              <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.enable" />
            </StatusActionButton>
            <StatusActionButton
              disabled={buttonStates.toggleDisable.buttonDisabled}
              destructive={true}
              onClick={onClickDisable}
            >
              <CircleBackslashIcon />
              <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.edit-schedule" />
            </StatusActionButton>
          </div>
        ) : (
          <StatusActionButton
            disabled={buttonStates.toggleDisable.buttonDisabled}
            destructive={true}
            className={styles.actionCellActionButton}
            onClick={onClickDisable}
          >
            <CircleBackslashIcon />
            <FormattedMessage id="UserDetailsAccountStatus.disable-user.action.disable" />
          </StatusActionButton>
        )}
      </ListCellLayout>
    );
  };

const AccountValidPeriodCell: React.VFC<AccountValidPeriodCellProps> =
  function AccountValidPeriodCell(props) {
    const { locale } = useContext(Context);
    const {
      data,
      onClickSetAccountValidPeriod,
      onClickEditAccountValidPeriod,
    } = props;
    const buttonStates = getButtonStates(data);
    return (
      <ListCellLayout className={styles.actionCell}>
        <div className={styles.actionCellLabel}>
          <Text as="p" size="2" weight="medium" className={styles.cellLabel}>
            <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.title" />
          </Text>
        </div>
        <div className={styles.actionCellBody}>
          <Text as="p" size="2" className={styles.cellBody}>
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
                        // eslint-disable-next-line react/no-unstable-nested-components
                        strong: (chunks: React.ReactNode) => (
                          <strong>{chunks}</strong>
                        ),
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
                        // eslint-disable-next-line react/no-unstable-nested-components
                        strong: (chunks: React.ReactNode) => (
                          <strong>{chunks}</strong>
                        ),
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
        <StatusActionButton
          disabled={buttonStates.setAccountValidPeriod.buttonDisabled}
          destructive={true}
          className={styles.actionCellActionButton}
          onClick={
            buttonStates.setAccountValidPeriod.accountValidFrom == null &&
            buttonStates.setAccountValidPeriod.accountValidUntil == null
              ? onClickSetAccountValidPeriod
              : onClickEditAccountValidPeriod
          }
        >
          <CalendarIcon />
          {buttonStates.setAccountValidPeriod.accountValidFrom == null &&
          buttonStates.setAccountValidPeriod.accountValidUntil == null ? (
            <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.action.set" />
          ) : (
            <FormattedMessage id="UserDetailsAccountStatus.account-valid-period.action.edit" />
          )}
        </StatusActionButton>
      </ListCellLayout>
    );
  };

const AnonymizeUserCell: React.VFC<AnonymizeUserCellProps> =
  function AnonymizeUserCell(props) {
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
          as="p"
          size="2"
          weight="medium"
          className={cn(styles.actionCellLabel, styles.cellLabel)}
        >
          <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.title" />
        </Text>
        <Text
          as="p"
          size="2"
          className={cn(styles.actionCellBody, styles.cellBody)}
        >
          <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.body" />
        </Text>
        {buttonStates.anonymize.anonymizeAt != null ? (
          <div className={styles.actionCellActionButtonContainer}>
            <StatusActionButton
              disabled={buttonStates.anonymize.buttonDisabled}
              onClick={onClickCancelAnonymization}
            >
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.cancel" />
            </StatusActionButton>
            <StatusActionButton
              disabled={buttonStates.anonymize.buttonDisabled}
              destructive={true}
              solid={true}
              onClick={onClickAnonymizeImmediately}
            >
              <CardStackMinusIcon />
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize-now" />
            </StatusActionButton>
          </div>
        ) : (
          <StatusActionButton
            disabled={buttonStates.anonymize.buttonDisabled}
            destructive={true}
            solid={true}
            className={styles.actionCellActionButton}
            onClick={onClickAnonymizeOrSchedule}
          >
            <CardStackMinusIcon />
            <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize" />
          </StatusActionButton>
        )}
      </ListCellLayout>
    );
  };

const RemoveUserCell: React.VFC<RemoveUserCellProps> = function RemoveUserCell(
  props
) {
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
        as="p"
        size="2"
        weight="medium"
        className={cn(styles.actionCellLabel, styles.cellLabel)}
      >
        <FormattedMessage id="UserDetailsAccountStatus.remove-user.title" />
      </Text>
      <Text
        as="p"
        size="2"
        className={cn(styles.actionCellBody, styles.cellBody)}
      >
        <FormattedMessage id="UserDetailsAccountStatus.remove-user.body" />
      </Text>
      {buttonStates.delete.deleteAt != null ? (
        <div className={styles.actionCellActionButtonContainer}>
          <StatusActionButton
            disabled={buttonStates.delete.buttonDisabled}
            onClick={onClickCancelDeletion}
          >
            <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.cancel" />
          </StatusActionButton>
          <StatusActionButton
            disabled={buttonStates.delete.buttonDisabled}
            destructive={true}
            solid={true}
            onClick={onClickDeleteImmediately}
          >
            <TrashIcon />
            <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove-now" />
          </StatusActionButton>
        </div>
      ) : (
        <StatusActionButton
          disabled={buttonStates.delete.buttonDisabled}
          destructive={true}
          solid={true}
          className={styles.actionCellActionButton}
          onClick={onClickDeleteOrSchedule}
        >
          <TrashIcon />
          <FormattedMessage id="UserDetailsAccountStatus.remove-user.action.remove" />
        </StatusActionButton>
      )}
    </ListCellLayout>
  );
};

const AccountLockoutCell: React.VFC<AccountLockoutCellProps> =
  function AccountLockoutCell(props) {
    const { locale } = useContext(Context);
    const { data, onClickResetAccountLockout } = props;
    const lockout = data.accountLockout;

    if (!lockout) {
      return null;
    }

    return (
      <ListCellLayout className={styles.actionCell}>
        <div className={styles.actionCellLabel}>
          <Text as="p" size="2" weight="medium" className={styles.cellLabel}>
            <FormattedMessage id="UserDetailsAccountStatus.account-lockout.title" />
          </Text>
        </div>
        <div className={styles.actionCellBody}>
          <Text as="p" size="2" className={styles.cellBody}>
            {lockout.isLocked ? (
              <>
                {lockout.lockoutType === "per_user" ? (
                  <FormattedMessage id="UserDetailsAccountStatus.account-lockout.body--locked-per-user" />
                ) : (
                  <FormattedMessage id="UserDetailsAccountStatus.account-lockout.body--locked-per-ip" />
                )}
                {lockout.lockedUntil ? (
                  <div style={{ marginTop: "8px" }}>
                    <FormattedMessage
                      id="UserDetailsAccountStatus.account-lockout.locked-until"
                      values={{
                        until:
                          formatDatetime(
                            locale,
                            new Date(lockout.lockedUntil)
                          ) ?? "",
                      }}
                    />
                  </div>
                ) : null}
                {lockout.lockoutType === "per_user_per_ip" &&
                lockout.lockedIPs.length > 0 ? (
                  <div style={{ marginTop: "8px" }}>
                    <Text as="p" size="1" weight="medium">
                      <FormattedMessage id="UserDetailsAccountStatus.account-lockout.locked-ips" />
                    </Text>
                    <ul style={{ marginTop: "4px", paddingLeft: "20px" }}>
                      {lockout.lockedIPs.map((ip, index) => (
                        <li key={index}>
                          {ip.ipAddress} -{" "}
                          {formatDatetime(locale, new Date(ip.lockedUntil)) ??
                            ""}
                        </li>
                      ))}
                    </ul>
                  </div>
                ) : null}
              </>
            ) : (
              <FormattedMessage id="UserDetailsAccountStatus.account-lockout.body--unlocked" />
            )}
          </Text>
          <Text as="p" size="1" color="gray" style={{ marginTop: "8px" }}>
            <FormattedMessage
              id="UserDetailsAccountStatus.account-lockout.learn-more"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                docLink: (chunks: React.ReactNode) => (
                  <ExternalLink href="https://docs.authgear.com/reference/rate-limits/account-lockout">
                    {chunks}
                  </ExternalLink>
                ),
              }}
            />
          </Text>
        </div>
        {lockout.isLocked ? (
          <StatusActionButton
            disabled={false}
            className={styles.actionCellActionButton}
            onClick={onClickResetAccountLockout}
          >
            <FormattedMessage id="UserDetailsAccountStatus.account-lockout.action.reset" />
          </StatusActionButton>
        ) : null}
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
    const onClickResetAccountLockout = useCallback(() => {
      setMode("reset-account-lockout");
      setDialogKey((prev) => prev + 1);
      setDialogHidden(false);
    }, []);

    const onDismiss: AccountStatusDialogProps["onDismiss"] = useCallback(
      async (info) => {
        setDialogHidden(true);
        if (info.deletedUser) {
          await navigate("./../..");
        }
      },
      [navigate]
    );

    return (
      <div className={styles.root}>
        <section className={styles.statusCard}>
          <Text as="p" size="3" weight="medium" className={styles.title}>
            <FormattedMessage id="UserDetailsAccountStatus.title" />
          </Text>
          <div className={styles.statusContent}>
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
            <AccountLockoutCell
              data={data}
              onClickResetAccountLockout={onClickResetAccountLockout}
            />
          </div>
        </section>
        <section className={cn(styles.statusCard, styles.dangerZoneCard)}>
          <Text
            as="p"
            size="3"
            weight="medium"
            className={cn(styles.title, styles.dangerZoneTitle)}
          >
            <FormattedMessage id="UserDetailsAccountStatus.danger-zone.title" />
          </Text>
          <div className={styles.statusContent}>
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
        </section>
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
  onDismiss: (info: { deletedUser: boolean }) => void | Promise<void>;
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
    | "reset-account-lockout"
    | "auto";
  accountStatus: AccountStatus;
}

export function AccountStatusDialog(
  props: AccountStatusDialogProps
): React.ReactElement {
  const { isHidden, onDismiss, mode, accountStatus } = props;
  const { locale, renderToString } = useContext(Context);

  const [defaultTemporarilyDisabledUntil] = useState(() =>
    DateTime.fromJSDate(new Date())
      .set({
        second: 0,
        millisecond: 0,
      })
      .plus({
        days: 7,
      })
      .toJSDate()
  );

  const [disableChoiceGroupKey, setDisableChoiceGroupKey] = useState<
    "indefinitely" | "temporarily"
  >("indefinitely");

  const [temporarilyDisabledUntil, setTemporarilyDisabledUntil] = useState(() =>
    accountStatus.temporarilyDisabledUntil != null
      ? new Date(accountStatus.temporarilyDisabledUntil)
      : defaultTemporarilyDisabledUntil
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
      } else if (date.getTime() === accountValidUntil.getTime()) {
        setAccountValidFrom(new Date(date.getTime() - 60 * 60 * 1000));
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
      } else if (date.getTime() === accountValidFrom.getTime()) {
        setAccountValidUntil(new Date(date.getTime() + 60 * 60 * 1000));
      } else {
        setAccountValidUntil(date);
      }
    },
    [accountValidFrom]
  );

  const disableForm = useMemo(() => {
    const formattedZone = formatSystemZone(new Date(), locale);
    return (
      <div className={styles.disableForm}>
        <div className="flex flex-col gap-2">
          <Text as="label" size="2" weight="medium">
            <FormattedMessage id="AccountStatusDialog.disable-user.disable-period.label" />
          </Text>
          <RadioGroup.Root
            className={styles.disablePeriodOptions}
            value={disableChoiceGroupKey}
            onValueChange={(value) => {
              setDisableChoiceGroupKey(value as "indefinitely" | "temporarily");
            }}
          >
            <RadioGroup.Item value="indefinitely">
              <FormattedMessage id="AccountStatusDialog.disable-user.disable-period.options.indefinitely" />
            </RadioGroup.Item>
            <RadioGroup.Item value="temporarily">
              <FormattedMessage id="AccountStatusDialog.disable-user.disable-period.options.temporarily" />
            </RadioGroup.Item>
          </RadioGroup.Root>
          {disableChoiceGroupKey === "temporarily" ? (
            <div className="flex flex-col gap-2">
              <Callout
                type="info"
                color="gray"
                size="1"
                showCloseButton={false}
                text={
                  <FormattedMessage
                    id="AccountStatusDialog.disable-user.timezone-description"
                    values={{
                      timezone: formattedZone,
                    }}
                  />
                }
              />
              <DateTimePicker
                minDateTime={"now"}
                pickedDateTime={temporarilyDisabledUntil}
                // @ts-expect-error
                onPickDateTime={setTemporarilyDisabledUntil}
                showClearButton={false}
              />
            </div>
          ) : null}
        </div>
        <TextField
          size="2"
          label={
            <FormattedMessage id="AccountStatusDialog.disable-user.disable-reason.label" />
          }
          placeholder={renderToString(
            "AccountStatusDialog.disable-user.disable-reason.placeholder"
          )}
          value={disableReason}
          onChange={(event) => {
            setDisableReason(event.currentTarget.value);
          }}
        />
      </div>
    );
  }, [
    disableChoiceGroupKey,
    disableReason,
    locale,
    renderToString,
    temporarilyDisabledUntil,
  ]);

  const accountValidPeriodForm = useMemo(() => {
    return (
      <AccountValidPeriodForm
        className={styles.accountValidPeriodDialogForm}
        calloutColor="gray"
        calloutSize="1"
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
  const {
    resetAccountLockout,
    loading: resetAccountLockoutLoading,
    error: resetAccountLockoutError,
  } = useResetAccountLockoutMutation();

  const loading =
    setDisabledStatusLoading ||
    setAccountValidPeriodLoading ||
    anonymizeUserLoading ||
    scheduleAccountAnonymizationLoading ||
    unscheduleAccountAnonymizationLoading ||
    deleteUserLoading ||
    scheduleAccountDeletionLoading ||
    unscheduleAccountDeletionLoading ||
    resetAccountLockoutLoading;
  const error =
    setDisabledStatusError ||
    setAccountValidPeriodError ||
    anonymizeUserError ||
    scheduleAccountAnonymizationError ||
    unscheduleAccountAnonymizationError ||
    deleteUserError ||
    scheduleAccountDeletionError ||
    unscheduleAccountDeletionError ||
    resetAccountLockoutError;

  const onDialogDismiss = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    onDismiss({ deletedUser: false });
  }, [loading, isHidden, onDismiss]);

  const onClickDisable = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await setDisabledStatus({
      userID: accountStatus.id,
      isDisabled: true,
      reason: sanitizedDisableReason,
      temporarilyDisabledFrom:
        disableChoiceGroupKey === "indefinitely" ? null : new Date(),
      temporarilyDisabledUntil:
        disableChoiceGroupKey === "indefinitely"
          ? null
          : temporarilyDisabledUntil,
    });
    onDismiss({ deletedUser: false });
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

  const onClickReenable = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await setDisabledStatus({
      userID: accountStatus.id,
      isDisabled: false,
      reason: null,
      temporarilyDisabledFrom: null,
      temporarilyDisabledUntil: null,
    });
    onDismiss({ deletedUser: false });
  }, [accountStatus.id, isHidden, loading, onDismiss, setDisabledStatus]);

  const onClickSetAccountValidPeriod = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await setAccountValidPeriod({
      userID: accountStatus.id,
      accountValidFrom: accountValidFrom,
      accountValidUntil: accountValidUntil,
    });
    onDismiss({ deletedUser: false });
  }, [
    accountStatus.id,
    accountValidFrom,
    accountValidUntil,
    isHidden,
    loading,
    onDismiss,
    setAccountValidPeriod,
  ]);

  const onClickAnonymize = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await anonymizeUser(accountStatus.id);
    await onDismiss({ deletedUser: false });
  }, [accountStatus.id, anonymizeUser, isHidden, loading, onDismiss]);

  const onClickScheduleAnonymization = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await scheduleAccountAnonymization(accountStatus.id);
    await onDismiss({ deletedUser: false });
  }, [
    accountStatus.id,
    isHidden,
    loading,
    onDismiss,
    scheduleAccountAnonymization,
  ]);

  const onClickUnscheduleAnonymization = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await unscheduleAccountAnonymization(accountStatus.id);
    await onDismiss({ deletedUser: false });
  }, [
    accountStatus.id,
    isHidden,
    loading,
    onDismiss,
    unscheduleAccountAnonymization,
  ]);

  const onClickDelete = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    try {
      await deleteUser(accountStatus.id);
      await onDismiss({ deletedUser: true });
    } catch (_e) {
      await onDismiss({ deletedUser: false });
    }
  }, [accountStatus.id, deleteUser, isHidden, loading, onDismiss]);

  const onClickScheduleDeletion = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await scheduleAccountDeletion(accountStatus.id);
    await onDismiss({ deletedUser: false });
  }, [accountStatus.id, isHidden, loading, onDismiss, scheduleAccountDeletion]);

  const onClickUnscheduleDeletion = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await unscheduleAccountDeletion(accountStatus.id);
    await onDismiss({ deletedUser: false });
  }, [
    accountStatus.id,
    isHidden,
    loading,
    onDismiss,
    unscheduleAccountDeletion,
  ]);

  const onClickResetAccountLockout = useCallback(async () => {
    if (loading || isHidden) {
      return;
    }
    await resetAccountLockout(accountStatus.id);
    await onDismiss({ deletedUser: false });
  }, [accountStatus.id, isHidden, loading, onDismiss, resetAccountLockout]);

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
        <Button
          size="2"
          color="indigo"
          disabled={loading}
          loading={loading}
          onClick={() => {
            onClickUnscheduleDeletion().finally(() => {});
          }}
        >
          <FormattedMessage id="AccountStatusDialog.cancel-deletion.action.cancel-deletion" />
        </Button>
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
        <Button
          size="2"
          color="indigo"
          disabled={loading}
          loading={loading}
          onClick={() => {
            onClickUnscheduleAnonymization().finally(() => {});
          }}
        >
          <FormattedMessage id="AccountStatusDialog.cancel-anonymization.action.cancel-anonymization" />
        </Button>
      );
    };
    const prepareReenable = () => {
      title = <FormattedMessage id="AccountStatusDialog.reenable-user.title" />;
      subText = (
        <FormattedMessage
          id="AccountStatusDialog.reenable-user.description"
          values={{
            ...args,
            // eslint-disable-next-line react/no-unstable-nested-components
            strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
          }}
        />
      );
      button1 = (
        <Button
          size="2"
          color="indigo"
          disabled={loading}
          loading={loading}
          onClick={() => {
            onClickReenable().finally(() => {});
          }}
        >
          <FormattedMessage id="AccountStatusDialog.reenable-user.action.reenable" />
        </Button>
      );
    };
    const prepareDisable = () => {
      title = <FormattedMessage id="AccountStatusDialog.disable-user.title" />;
      subText = (
        <FormattedMessage
          id="AccountStatusDialog.disable-user.description"
          values={{
            ...args,
            // eslint-disable-next-line react/no-unstable-nested-components
            strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
          }}
        />
      );
      body = disableForm;
      button1 = (
        <Button
          size="2"
          color="red"
          disabled={loading}
          loading={loading}
          onClick={() => {
            onClickDisable().finally(() => {});
          }}
        >
          <FormattedMessage id="AccountStatusDialog.disable-user.action.disable" />
        </Button>
      );
    };
    const prepareEditAccountValidPeriod = () => {
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
        <Button
          size="2"
          color="indigo"
          disabled={loading}
          loading={loading}
          onClick={() => {
            onClickSetAccountValidPeriod().finally(() => {});
          }}
        >
          <FormattedMessage id="AccountStatusDialog.account-valid-period.action.edit" />
        </Button>
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
          <Button
            size="2"
            color="indigo"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickSetAccountValidPeriod().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.account-valid-period.action.save" />
          </Button>
        );
        break;
      case "edit-account-valid-period":
        prepareEditAccountValidPeriod();
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
          <Button
            size="2"
            color="indigo"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickScheduleAnonymization().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.anonymize-user.action.schedule-anonymization" />
          </Button>
        );
        button2 = (
          <Button
            size="2"
            color="red"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickAnonymize().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.anonymize-user.action.anonymize-immediately" />
          </Button>
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
          <Button
            size="2"
            color="red"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickAnonymize().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.anonymize-user.action.anonymize-immediately" />
          </Button>
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
          <Button
            size="2"
            color="indigo"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickScheduleDeletion().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.delete-user.action.schedule-deletion" />
          </Button>
        );
        button2 = (
          <Button
            size="2"
            color="red"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickDelete().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.delete-user.action.delete-immediately" />
          </Button>
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
          <Button
            size="2"
            color="red"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickDelete().finally(() => {});
            }}
          >
            <FormattedMessage id="AccountStatusDialog.delete-user.action.delete-immediately" />
          </Button>
        );
        break;
      case "reset-account-lockout":
        title = (
          <FormattedMessage id="UserDetailsAccountStatus.account-lockout.confirm-dialog.title" />
        );
        subText = (
          <FormattedMessage
            id="UserDetailsAccountStatus.account-lockout.confirm-dialog.description"
            values={args}
          />
        );
        button1 = (
          <Button
            size="2"
            color="indigo"
            disabled={loading}
            loading={loading}
            onClick={() => {
              onClickResetAccountLockout().finally(() => {});
            }}
          >
            <FormattedMessage id="UserDetailsAccountStatus.account-lockout.action.reset" />
          </Button>
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
          case "edit-account-valid-period":
            prepareEditAccountValidPeriod();
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
    onClickResetAccountLockout,
    onClickScheduleAnonymization,
    onClickScheduleDeletion,
    onClickSetAccountValidPeriod,
    onClickUnscheduleAnonymization,
    onClickUnscheduleDeletion,
  ]);

  const accountStatusErrorRules: ErrorParseRule[] = useMemo(() => {
    return [
      makeInvalidAccountStatusTransitionErrorParseRule(
        "AccountValidFromShouldBeBeforeTemporarilyDisabledFrom",
        "UserDetailsAccountStatus.error.temporary-disable-until-later-than-valid-period"
      ),
      makeInvalidAccountStatusTransitionErrorParseRule(
        "TemporarilyDisabledUntilShouldBeBeforeAccountValidUntil",
        "UserDetailsAccountStatus.error.temporary-disable-until-later-than-valid-period"
      ),
    ];
  }, []);

  return (
    <>
      <Dialog.Root
        open={!isHidden}
        onOpenChange={(open) => {
          if (!open) {
            onDialogDismiss();
          }
        }}
      >
        <Dialog.Content maxWidth="560px" size="3">
          <Dialog.Title>
            {dialogContentPropsAndDialogSlots.dialogContentProps.title}
          </Dialog.Title>
          {dialogContentPropsAndDialogSlots.dialogContentProps.subText !=
          null ? (
            <Dialog.Description size="2">
              {dialogContentPropsAndDialogSlots.dialogContentProps.subText}
            </Dialog.Description>
          ) : null}
          {dialogContentPropsAndDialogSlots.body}
          <div className={styles.dialogActions}>
            <SecondaryButton
              size="2"
              disabled={loading}
              text={<FormattedMessage id="cancel" />}
              onClick={onDialogDismiss}
            />
            {dialogContentPropsAndDialogSlots.button1}
            {dialogContentPropsAndDialogSlots.button2}
          </div>
        </Dialog.Content>
      </Dialog.Root>
      <ErrorDialog
        error={error}
        rules={accountStatusErrorRules}
        titleMessageID="UserDetailsAccountStatus.error.title"
      />
    </>
  );
}

export interface AccountStatusBadgeProps {
  className?: string;
  accountStatus: AccountStatus;
}

export function AccountStatusBadge(
  props: AccountStatusBadgeProps
): React.ReactElement | null {
  const { accountStatus, className } = props;
  const accountState = getMostAppropriateAccountState(accountStatus);
  let id: string | null = null;
  switch (accountState.state) {
    case "scheduled-deletion":
      id = "AccountStatusBadge.scheduled-deletion";
      break;
    case "anonymized":
      id = "AccountStatusBadge.anonymized";
      break;
    case "scheduled-anonymization":
      id = "AccountStatusBadge.scheduled-anonymization";
      break;
    case "less-than-account-valid-from":
      id = "AccountStatusBadge.account-outside-valid-period";
      break;
    case "greater-than-or-equal-to-account-valid-until":
      id = "AccountStatusBadge.account-outside-valid-period";
      break;
    case "disabled":
      id = "AccountStatusBadge.disabled";
      break;
    case "normal":
      break;
  }
  if (id == null) {
    return null;
  }

  return (
    <Badge className={className} color="red" radius="small">
      <FormattedMessage id={id} />
    </Badge>
  );
}

export interface AccountStatusMessageBarProps {
  accountStatus: AccountStatus;
}

export function AccountStatusMessageBar(
  props: AccountStatusMessageBarProps
): React.ReactElement | null {
  const { accountStatus } = props;
  const { locale } = useContext(Context);
  const accountState = getMostAppropriateAccountState(accountStatus);

  let message: React.ReactNode = null;
  switch (accountState.state) {
    case "scheduled-deletion":
      message = (
        <FormattedMessage
          id="AccountStatusMessageBar.scheduled-deletion"
          values={{
            date: formatDatetime(locale, accountState.deleteAt) ?? "",
          }}
        />
      );
      break;
    case "anonymized":
      message = <FormattedMessage id="AccountStatusMessageBar.anonymized" />;
      break;
    case "scheduled-anonymization":
      message = (
        <FormattedMessage
          id="AccountStatusMessageBar.scheduled-anonymization"
          values={{
            date: formatDatetime(locale, accountState.anonymizeAt) ?? "",
          }}
        />
      );
      break;
    case "less-than-account-valid-from":
      message = (
        <FormattedMessage
          id="AccountStatusMessageBar.account-valid-from"
          values={{
            date: formatDatetime(locale, accountState.accountValidFrom) ?? "",
          }}
        />
      );
      break;
    case "greater-than-or-equal-to-account-valid-until":
      message = (
        <FormattedMessage
          id="AccountStatusMessageBar.account-valid-until"
          values={{
            date: formatDatetime(locale, accountState.accountValidUntil) ?? "",
          }}
        />
      );
      break;
    case "disabled":
      if (accountState.temporarilyDisabledUntil != null) {
        message = (
          <FormattedMessage
            id="AccountStatusMessageBar.disabled-tempoararily"
            values={{
              date:
                formatDatetime(locale, accountState.temporarilyDisabledUntil) ??
                "",
            }}
          />
        );
      } else {
        message = (
          <FormattedMessage id="AccountStatusMessageBar.disabled-indefinitely" />
        );
      }
      break;
    case "normal":
      break;
  }
  if (message == null) {
    return null;
  }

  return <Callout type="warning" text={message} showCloseButton={false} />;
}

export default UserDetailsAccountStatus;
