import React, { useMemo, useContext, useState, useCallback, useRef } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import PasswordPolicySettings from "./PasswordPolicySettings";
import ForgotPasswordSettings from "./ForgotPasswordSettings";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import SwitchTabBlockerDialog from "../../SwitchTabBlockerDialog";

import styles from "./PasswordsScreen.module.scss";

const PASSWORD_POLICY_PIVOT_KEY = "password_policy";
const FORGOT_PASSWORD_POLICY_KEY = "forgot_password";

type PivotKey = typeof PASSWORD_POLICY_PIVOT_KEY | typeof FORGOT_PASSWORD_POLICY_KEY;

const PasswordsScreen: React.FC = function PasswordsScreen() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);
  const { loading, error, data, refetch } = useAppConfigQuery(appID);
  const { effectiveAppConfig, rawAppConfig } = useMemo(() => {
    const node = data?.node;
    return node?.__typename === "App"
      ? {
          effectiveAppConfig: node.effectiveAppConfig,
          rawAppConfig: node.rawAppConfig,
        }
      : {
          effectiveAppConfig: null,
          rawAppConfig: null,
        };
  }, [data]);

  const [currentPivotKey, setCurrentPivotKey] = useState<PivotKey>(PASSWORD_POLICY_PIVOT_KEY);
  const selectedPivotKeyRef = useRef<PivotKey>(PASSWORD_POLICY_PIVOT_KEY);
  const [isPasswordPolicyFormModified, setIsPasswordPolicyFormModified] = useState(false);
  const [isForgotPasswordFormModified, setIsForgotPasswordFormModified] = useState(false);
  const [shouldDisplayDiscardChangesDialog, setShouldDisplayDiscardChangesDialog] = useState(false);

  const onPivotLinkClick = useCallback((item?: PivotItem) => {
    if (!item || !item.props.itemKey) {
      return;
    }

    const newPivotKey = item.props.itemKey as PivotKey;
    selectedPivotKeyRef.current = newPivotKey;

    if (newPivotKey !== currentPivotKey) {
      switch (currentPivotKey) {
        case PASSWORD_POLICY_PIVOT_KEY:
          if (isPasswordPolicyFormModified) {
            setShouldDisplayDiscardChangesDialog(true);
          } else {
            setCurrentPivotKey(newPivotKey);
          }
          break;
        case FORGOT_PASSWORD_POLICY_KEY:
          if (isForgotPasswordFormModified) {
            setShouldDisplayDiscardChangesDialog(true);
          } else {
            setCurrentPivotKey(newPivotKey);
          }
          break;
        default:
          break;
      }
    }
  }, [setCurrentPivotKey, setShouldDisplayDiscardChangesDialog, currentPivotKey, isPasswordPolicyFormModified, isForgotPasswordFormModified]);

  const onIsPasswordPolicyFormModifiedChange = useCallback((modified: boolean) => {
    setIsPasswordPolicyFormModified(modified);
  }, [setIsPasswordPolicyFormModified]);
  
  const onIsForgotPasswordFormModifiedChange = useCallback((modified: boolean) => {
    setIsForgotPasswordFormModified(modified);
  }, [setIsForgotPasswordFormModified]);

  const onDialogDismiss = useCallback(() => {
    setShouldDisplayDiscardChangesDialog(false);
  }, [setShouldDisplayDiscardChangesDialog]);

  const onDialogConfirm = useCallback(() => {
    setShouldDisplayDiscardChangesDialog(false);
    // The pivot item would be unmounted and reset to initial state when switching back.
    // So don't need to reset form data.
    setCurrentPivotKey(selectedPivotKeyRef.current);
  }, [setShouldDisplayDiscardChangesDialog, setCurrentPivotKey]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={cn(styles.root, { [styles.loading]: updatingAppConfig })}>
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="PasswordsScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot
            onLinkClick={onPivotLinkClick}
            selectedKey={currentPivotKey}
          >
            <PivotItem
              headerText={renderToString("PasswordsScreen.password-policy.title")}
              itemKey={PASSWORD_POLICY_PIVOT_KEY}
            >
              <PasswordPolicySettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
                onIsFormModifiedChange={onIsPasswordPolicyFormModifiedChange}
              />
            </PivotItem>
            <PivotItem
              headerText={renderToString("PasswordsScreen.forgot-password.title")}
              itemKey={FORGOT_PASSWORD_POLICY_KEY}
            >
              <ForgotPasswordSettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
                onIsFormModifiedChange={onIsForgotPasswordFormModifiedChange}
              />
            </PivotItem>
          </Pivot>
        </div>
      </div>
      <SwitchTabBlockerDialog
        hidden={!shouldDisplayDiscardChangesDialog}
        onDialogConfirm={onDialogConfirm}
        onDialogDismiss={onDialogDismiss}
      />
    </main>
  );
};

export default PasswordsScreen;
