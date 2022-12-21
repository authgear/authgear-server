import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter, IDialogContentProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ErrorDialog from "../../error/ErrorDialog";
import { useAnonymizeUserMutation } from "./mutations/anonymizeUserMutation";
import { useScheduleAccountAnonymizationMutation } from "./mutations/scheduleAccountAnonymization";
import { extractRawID } from "../../util/graphql";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

interface AnonymizeUserDialogProps {
  isHidden: boolean;
  onDismiss: (anonymizedUser: boolean) => void;
  userID: string;
  userAnonymizeAt: string | null;
  endUserAccountIdentifier: string | undefined;
}

const AnonymizeUserDialog: React.VFC<AnonymizeUserDialogProps> = React.memo(
  function AnonymizeUserDialog(props: AnonymizeUserDialogProps) {
    const {
      isHidden,
      onDismiss,
      userID,
      userAnonymizeAt,
      endUserAccountIdentifier,
    } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
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

    const loading = anonymizeUserLoading || scheduleAccountAnonymizationLoading;
    const error = anonymizeUserError || scheduleAccountAnonymizationError;

    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    const onClickRemove = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      anonymizeUser(userID)
        .then(() => onDismiss(true))
        .catch(() => onDismiss(false));
    }, [loading, isHidden, anonymizeUser, userID, onDismiss]);

    const onClickScheduleAnonymization = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      scheduleAccountAnonymization(userID)
        .then(() => onDismiss(true))
        .catch(() => onDismiss(false));
      onDismiss(false);
    }, [loading, isHidden, scheduleAccountAnonymization, userID, onDismiss]);

    const dialogContentProps: IDialogContentProps = useMemo(
      () => ({
        title: renderToString("AnonymizeUserDialog.title"),
        subText: renderToString("AnonymizeUserDialog.text", {
          username: endUserAccountIdentifier ?? extractRawID(userID),
        }),
      }),
      [renderToString, endUserAccountIdentifier, userID]
    );

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
          minWidth={560}
        >
          <DialogFooter>
            {userAnonymizeAt == null ? (
              <PrimaryButton
                onClick={onClickScheduleAnonymization}
                disabled={loading}
                text={
                  <FormattedMessage id="AnonymizeUserDialog.label.schedule-removal" />
                }
              />
            ) : null}
            <PrimaryButton
              theme={themes.destructive}
              onClick={onClickRemove}
              disabled={loading}
              text={
                <FormattedMessage id="AnonymizeUserDialog.label.remove-immediately" />
              }
            />
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
);

export default AnonymizeUserDialog;
