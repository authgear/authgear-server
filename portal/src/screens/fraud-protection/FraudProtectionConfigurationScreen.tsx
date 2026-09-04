import React, {
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Text } from "@radix-ui/themes";
import { Address4, Address6 } from "ip-address";
import { produce } from "immer";
import { default as parseLibPhoneNumber } from "libphonenumber-js";
import { Context, FormattedMessage } from "../../intl";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { FormContainerBase } from "../../FormContainerBase";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { FeatureDisabledCallout } from "../../components/v2/FeatureDisabledCallout/FeatureDisabledCallout";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { APIError } from "../../error/error";
import {
  LocalValidationError,
  makeLocalValidationError,
} from "../../error/validation";
import {
  FraudProtectionDecisionAction,
  FraudProtectionFeatureConfig,
  PortalAPIAppConfig,
} from "../../types";
import { usePivotNavigation } from "../../hook/usePivot";
import { clearEmptyObject } from "../../util/misc";
import { useAppFeatureConfigQuery } from "../../graphql/portal/query/appFeatureConfigQuery";
import { Tag } from "../../CustomTagPicker";
import FraudProtectionOverviewTab from "../../components/fraud-protection/FraudProtectionOverviewTab";
import FraudProtectionLogsTab from "../../components/fraud-protection/FraudProtectionLogsTab";
import FraudProtectionSettingsTab from "../../components/fraud-protection/FraudProtectionSettingsTab";
import styles from "./FraudProtectionConfigurationScreen.module.css";

interface FormState {
  enabled: boolean;
  enforcementMode: FraudProtectionDecisionAction;
  ipAllowlist: string;
  phoneAllowlist: string;
  ipCountryAllowlist: string[];
  phoneCountryAllowlist: string[];
}

function splitAllowlist(raw: string): string[] {
  return raw
    .split(/,|\n/)
    .map((item) => item.trim())
    .filter((item) => item !== "");
}

function toCIDRs(raw: string): string[] {
  return splitAllowlist(raw).map((item) => {
    if (item.includes("/")) {
      return item;
    }
    if (Address4.isValid(item)) {
      return `${item}/32`;
    }
    if (Address6.isValid(item)) {
      return `${item}/128`;
    }
    return item;
  });
}

function escapeRegExp(input: string): string {
  return input.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function isValidRegex(input: string): boolean {
  try {
    // eslint-disable-next-line no-new
    new RegExp(input);
    return true;
  } catch {
    return false;
  }
}

function normalizePhoneAllowlistItemForSave(item: string): string {
  const parsed = parseLibPhoneNumber(item);
  if (parsed?.isPossible() === true) {
    return `^${escapeRegExp(parsed.number)}$`;
  }
  return item;
}

function toPhoneRegex(raw: string): string[] {
  return splitAllowlist(raw).map((item) => {
    return normalizePhoneAllowlistItemForSave(item);
  });
}

function toDisplayPhoneAllowlistItem(item: string): string {
  const match = /^\^\\\+([1-9]\d+)\$$/.exec(item);
  if (match?.[1] != null) {
    return `+${match[1]}`;
  }
  return item;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    enabled: config.fraud_protection?.enabled ?? true,
    enforcementMode: config.fraud_protection?.decision?.action ?? "record_only",
    ipAllowlist:
      config.fraud_protection?.decision?.always_allow?.ip_address?.cidrs?.join(
        "\n"
      ) ?? "",
    phoneAllowlist:
      config.fraud_protection?.decision?.always_allow?.phone_number?.regex
        ?.map(toDisplayPhoneAllowlistItem)
        .join("\n") ?? "",
    ipCountryAllowlist:
      config.fraud_protection?.decision?.always_allow?.ip_address
        ?.geo_location_codes ?? [],
    phoneCountryAllowlist:
      config.fraud_protection?.decision?.always_allow?.phone_number
        ?.geo_location_codes ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    draft.fraud_protection ??= {};
    draft.fraud_protection.enabled = currentState.enabled;

    draft.fraud_protection.decision ??= {};
    draft.fraud_protection.decision.action = currentState.enforcementMode;

    draft.fraud_protection.decision.always_allow ??= {};
    draft.fraud_protection.decision.always_allow.ip_address ??= {};
    draft.fraud_protection.decision.always_allow.phone_number ??= {};

    const cidrs = toCIDRs(currentState.ipAllowlist);
    if (cidrs.length > 0) {
      draft.fraud_protection.decision.always_allow.ip_address.cidrs = cidrs;
    } else {
      delete draft.fraud_protection.decision.always_allow.ip_address.cidrs;
    }

    const regex = toPhoneRegex(currentState.phoneAllowlist);
    if (regex.length > 0) {
      draft.fraud_protection.decision.always_allow.phone_number.regex = regex;
    } else {
      delete draft.fraud_protection.decision.always_allow.phone_number.regex;
    }

    const ipGeos = currentState.ipCountryAllowlist;
    if (ipGeos.length > 0) {
      draft.fraud_protection.decision.always_allow.ip_address.geo_location_codes =
        ipGeos;
    } else {
      delete draft.fraud_protection.decision.always_allow.ip_address
        .geo_location_codes;
    }

    const phoneGeos = currentState.phoneCountryAllowlist;
    if (phoneGeos.length > 0) {
      draft.fraud_protection.decision.always_allow.phone_number.geo_location_codes =
        phoneGeos;
    } else {
      delete draft.fraud_protection.decision.always_allow.phone_number
        .geo_location_codes;
    }

    clearEmptyObject(draft);
  });
}

function validateFormState(state: FormState): APIError | null {
  const invalidItems: string[] = [];
  for (const item of splitAllowlist(state.phoneAllowlist)) {
    const normalized = normalizePhoneAllowlistItemForSave(item);
    if (normalized === item && !isValidRegex(item)) {
      invalidItems.push(item);
    }
  }

  if (invalidItems.length === 0) {
    return null;
  }

  const errors: LocalValidationError[] = invalidItems.map((item) => ({
    location: "/fraud_protection/decision/always_allow/phone_number/regex",
    messageID: "FraudProtectionConfigurationScreen.allowlist.phone.invalidItem",
    arguments: { item },
  }));
  return makeLocalValidationError(errors);
}

type FraudProtectionTab = "overview" | "logs" | "settings";

interface FraudProtectionConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  fraudProtectionFeatureConfig?: FraudProtectionFeatureConfig;
  selectedKey: FraudProtectionTab;
  onChangeKey: (key: FraudProtectionTab) => void;
  onToggleEnabledAndSave: (enabled: boolean) => void;
}

const FraudProtectionConfigurationContent: React.VFC<FraudProtectionConfigurationContentProps> =
  function FraudProtectionConfigurationContent(props) {
    const {
      form,
      fraudProtectionFeatureConfig,
      selectedKey,
      onChangeKey,
      onToggleEnabledAndSave,
    } = props;
    const { renderToString } = useContext(Context);
    const { state, setState, getIsDirty } = form;
    const isModifiable = fraudProtectionFeatureConfig?.is_modifiable ?? false;

    const contentRef = useRef<HTMLDivElement>(null);
    // getIsDirty's identity changes exactly when the dirtiness does, so this
    // memo recomputes only when the save bar's visibility would change.
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);

    const tabs = useMemo(
      () => [
        {
          value: "overview",
          label: renderToString(
            "FraudProtectionConfigurationScreen.tab.overview.title"
          ),
        },
        {
          value: "logs",
          label: renderToString(
            "FraudProtectionConfigurationScreen.tab.logs.title"
          ),
        },
        {
          value: "settings",
          label: renderToString(
            "FraudProtectionConfigurationScreen.tab.settings.title"
          ),
        },
      ],
      [renderToString]
    );

    const onEnforcementModeChange = useCallback(
      (value: FraudProtectionDecisionAction) => {
        setState((current) => ({
          ...current,
          enforcementMode: value,
        }));
      },
      [setState]
    );

    const onIPAllowlistChange = useCallback(
      (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const value = e.currentTarget.value;
        setState((current) => ({
          ...current,
          ipAllowlist: value,
        }));
      },
      [setState]
    );

    const onPhoneAllowlistChange = useCallback(
      (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const value = e.currentTarget.value;
        setState((current) => ({
          ...current,
          phoneAllowlist: value,
        }));
      },
      [setState]
    );

    const onIPCountryAllowlistChange = useCallback(
      (items?: Tag[]) => {
        setState((current) => ({
          ...current,
          ipCountryAllowlist: items?.map((it) => it.key as string) ?? [],
        }));
      },
      [setState]
    );

    const onPhoneCountryAllowlistChange = useCallback(
      (items?: Tag[]) => {
        setState((current) => ({
          ...current,
          phoneCountryAllowlist: items?.map((it) => it.key as string) ?? [],
        }));
      },
      [setState]
    );

    return (
      <APIResourceScreenLayout
        layout="list"
        breadcrumbItems={[
          {
            to: "",
            label: (
              <FormattedMessage id="FraudProtectionConfigurationScreen.title" />
            ),
          },
        ]}
        headerDescription={
          <Text as="p" size="2" className={styles.headerDescription}>
            <FormattedMessage id="FraudProtectionConfigurationScreen.description" />
          </Text>
        }
      >
        <div
          ref={contentRef}
          className={cn(
            styles.content,
            selectedKey === "settings" && isDirty && styles.contentInset
          )}
        >
          {isModifiable ? null : (
            <FeatureDisabledCallout messageID="FraudProtectionConfigurationScreen.disabled" />
          )}
          <Toggle
            checked={state.enabled}
            disabled={!isModifiable}
            textWeight="medium"
            text={renderToString(
              "FraudProtectionConfigurationScreen.enable.label"
            )}
            onCheckedChange={onToggleEnabledAndSave}
          />
          {state.enabled ? (
            <div className={styles.settings}>
              <OverflowTabs
                value={selectedKey}
                onValueChange={(value) =>
                  onChangeKey(value as FraudProtectionTab)
                }
                tabs={tabs}
              />
              {selectedKey === "overview" ? (
                <FraudProtectionOverviewTab
                  enabled={state.enabled}
                  enforcementMode={state.enforcementMode}
                  onChangeToSettings={() => onChangeKey("settings")}
                />
              ) : null}
              {selectedKey === "logs" ? <FraudProtectionLogsTab /> : null}
              {selectedKey === "settings" ? (
                <FraudProtectionSettingsTab
                  isModifiable={isModifiable}
                  enforcementMode={state.enforcementMode}
                  ipAllowlist={state.ipAllowlist}
                  phoneAllowlist={state.phoneAllowlist}
                  ipCountryAllowlist={state.ipCountryAllowlist}
                  phoneCountryAllowlist={state.phoneCountryAllowlist}
                  onEnforcementModeChange={onEnforcementModeChange}
                  onIPAllowlistChange={onIPAllowlistChange}
                  onPhoneAllowlistChange={onPhoneAllowlistChange}
                  onIPCountryAllowlistChange={onIPCountryAllowlistChange}
                  onPhoneCountryAllowlistChange={onPhoneCountryAllowlistChange}
                />
              ) : null}
            </div>
          ) : null}
          {state.enabled && selectedKey === "settings" ? (
            <SaveFunctionBar anchorRef={contentRef} />
          ) : null}
        </div>
      </APIResourceScreenLayout>
    );
  };

// Access to this screen is gated at the route level by RequireAppFeature (see
// AppRoot), which redirects to getting started when the project lacks fraud
// protection. The screen therefore assumes the feature is available and is free
// to use usePivotNavigation for its tabs without racing the redirect.
const FraudProtectionConfigurationScreen: React.VFC =
  function FraudProtectionConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const [showDisableConfirmation, setShowDisableConfirmation] =
      useState(false);
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
      validate: validateFormState,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);
    const { selectedKey, onChangeKey } = usePivotNavigation<FraudProtectionTab>(
      ["overview", "logs", "settings"],
      undefined,
      undefined,
      true // push history so the back button returns to the previous tab
    );

    const handleToggleEnabledAndSave = useCallback(
      (enabled: boolean) => {
        if (!enabled && form.state.enabled) {
          setShowDisableConfirmation(true);
        } else if (enabled && !form.state.enabled) {
          form.saveWith((_current) => ({
            ...form.initialState,
            enabled: true,
          }));
        }
      },
      [form]
    );

    const handleConfirmDisable = useCallback(() => {
      setShowDisableConfirmation(false);
      form.saveWith((_current) => ({ ...form.initialState, enabled: false }));
    }, [form]);

    const handleCancelDisable = useCallback(() => {
      setShowDisableConfirmation(false);
    }, []);

    if (form.isLoading || featureConfig.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (featureConfig.loadError) {
      return (
        <ShowError
          error={featureConfig.loadError}
          onRetry={featureConfig.reload}
        />
      );
    }

    const isModifiable =
      featureConfig.effectiveFeatureConfig?.fraud_protection?.is_modifiable ??
      false;

    return (
      <FormContainerBase form={form} canSave={isModifiable}>
        <FraudProtectionConfigurationContent
          form={form}
          fraudProtectionFeatureConfig={
            featureConfig.effectiveFeatureConfig?.fraud_protection
          }
          selectedKey={selectedKey}
          onChangeKey={onChangeKey}
          onToggleEnabledAndSave={handleToggleEnabledAndSave}
        />
        <ConfirmationDialog
          open={showDisableConfirmation}
          onOpenChange={(open) => {
            if (!open) {
              handleCancelDisable();
            }
          }}
          title={
            <FormattedMessage id="FraudProtectionConfigurationScreen.disable.confirmation.title" />
          }
          description={
            <FormattedMessage id="FraudProtectionConfigurationScreen.disable.confirmation.description" />
          }
          confirmText={
            <FormattedMessage id="FraudProtectionConfigurationScreen.disable.confirmation.confirm" />
          }
          cancelText={
            <FormattedMessage id="FraudProtectionConfigurationScreen.disable.confirmation.cancel" />
          }
          confirmColor="red"
          onConfirm={handleConfirmDisable}
          onCancel={handleCancelDisable}
        />
      </FormContainerBase>
    );
  };

export default FraudProtectionConfigurationScreen;
