import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import { IChoiceGroupOption, ITag, PivotItem } from "@fluentui/react";
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
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import FormContainer from "../../FormContainer";
import Toggle from "../../Toggle";
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
import FeatureDisabledMessageBar from "../../graphql/portal/FeatureDisabledMessageBar";
import { AGPivot } from "../../components/common/AGPivot";
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
  onLinkClick: (item?: PivotItem) => void;
  onChangeKey: (key: FraudProtectionTab) => void;
}

const FraudProtectionConfigurationContent: React.VFC<FraudProtectionConfigurationContentProps> =
  function FraudProtectionConfigurationContent(props) {
    const {
      form,
      fraudProtectionFeatureConfig,
      selectedKey,
      onLinkClick,
      onChangeKey,
    } = props;
    const { renderToString } = useContext(Context);
    const { state, setState } = form;
    const isModifiable = fraudProtectionFeatureConfig?.is_modifiable ?? false;

    const onEnableChange = useCallback(
      (
        _event: React.FormEvent<HTMLElement | HTMLInputElement>,
        checked?: boolean
      ) => {
        setState((current) => ({
          ...current,
          enabled: checked ?? false,
        }));
      },
      [setState]
    );

    const onEnforcementModeChange = useCallback(
      (
        _event: React.FormEvent<HTMLElement | HTMLInputElement> | undefined,
        option?: IChoiceGroupOption
      ) => {
        const key = option?.key;
        if (key !== "record_only" && key !== "deny_if_any_warning") {
          return;
        }
        setState((current) => ({
          ...current,
          enforcementMode: key,
        }));
      },
      [setState]
    );

    const onIPAllowlistChange = useCallback(
      (
        _event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        setState((current) => ({
          ...current,
          ipAllowlist: value ?? "",
        }));
      },
      [setState]
    );

    const onPhoneAllowlistChange = useCallback(
      (
        _event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        setState((current) => ({
          ...current,
          phoneAllowlist: value ?? "",
        }));
      },
      [setState]
    );

    const onIPCountryAllowlistChange = useCallback(
      (items?: ITag[]) => {
        setState((current) => ({
          ...current,
          ipCountryAllowlist: items?.map((it) => it.key as string) ?? [],
        }));
      },
      [setState]
    );

    const onPhoneCountryAllowlistChange = useCallback(
      (items?: ITag[]) => {
        setState((current) => ({
          ...current,
          phoneCountryAllowlist: items?.map((it) => it.key as string) ?? [],
        }));
      },
      [setState]
    );

    return (
      <ScreenContent layout="list">
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="FraudProtectionConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="FraudProtectionConfigurationScreen.description" />
        </ScreenDescription>
        <div className={styles.content}>
          {isModifiable ? null : (
            <FeatureDisabledMessageBar
              className={styles.widget}
              messageID="FraudProtectionConfigurationScreen.disabled"
            />
          )}
          <div
            className={`${styles.page} ${
              selectedKey === "overview" ? styles.pageOverview : ""
            } ${selectedKey === "logs" ? styles.pageLogs : ""}`}
          >
            <Toggle
              checked={state.enabled}
              disabled={!isModifiable}
              label={renderToString(
                "FraudProtectionConfigurationScreen.enable.label"
              )}
              inlineLabel={false}
              onChange={onEnableChange}
            />
            {state.enabled ? (
              <div className={styles.settings}>
                <AGPivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
                  <PivotItem
                    headerText={renderToString(
                      "FraudProtectionConfigurationScreen.tab.overview.title"
                    )}
                    itemKey="overview"
                  />
                  <PivotItem
                    headerText={renderToString(
                      "FraudProtectionConfigurationScreen.tab.logs.title"
                    )}
                    itemKey="logs"
                  />
                  <PivotItem
                    headerText={renderToString(
                      "FraudProtectionConfigurationScreen.tab.settings.title"
                    )}
                    itemKey="settings"
                  />
                </AGPivot>
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
                    onPhoneCountryAllowlistChange={
                      onPhoneCountryAllowlistChange
                    }
                  />
                ) : null}
              </div>
            ) : null}
          </div>
        </div>
      </ScreenContent>
    );
  };

const FraudProtectionConfigurationScreen: React.VFC =
  function FraudProtectionConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
      validate: validateFormState,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);
    const { selectedKey, onLinkClick, onChangeKey } =
      usePivotNavigation<FraudProtectionTab>(["overview", "logs", "settings"]);

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
      <FormContainer
        form={form}
        canSave={isModifiable}
        showDiscardButton={true}
        stickyFooterComponent={true}
        hideFooterComponent={selectedKey !== "settings"}
      >
        <FraudProtectionConfigurationContent
          form={form}
          fraudProtectionFeatureConfig={
            featureConfig.effectiveFeatureConfig?.fraud_protection
          }
          selectedKey={selectedKey}
          onLinkClick={onLinkClick}
          onChangeKey={onChangeKey}
        />
      </FormContainer>
    );
  };

export default FraudProtectionConfigurationScreen;
