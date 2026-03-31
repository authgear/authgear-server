import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  PivotItem,
  Text,
} from "@fluentui/react";
import { Address4, Address6 } from "ip-address";
import { produce } from "immer";
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
import Toggle from "../../Toggle";
import FormContainer from "../../FormContainer";
import FormTextField from "../../FormTextField";
import {
  FraudProtectionDecisionAction,
  FraudProtectionFeatureConfig,
  PortalAPIAppConfig,
} from "../../types";
import { AGPivot } from "../../components/common/AGPivot";
import { usePivotNavigation } from "../../hook/usePivot";
import { clearEmptyObject } from "../../util/misc";
import { parsePhoneNumber } from "../../util/phone";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import styles from "./FraudProtectionConfigurationScreen.module.css";

interface FormState {
  enabled: boolean;
  enforcementMode: FraudProtectionDecisionAction;
  ipAllowlist: string;
  phoneAllowlist: string;
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

function toPhoneRegex(raw: string): string[] {
  return splitAllowlist(raw).map((item) => {
    const normalized = parsePhoneNumber(item);
    if (normalized == null) {
      return item;
    }
    return `^${escapeRegExp(normalized)}$`;
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
    enforcementMode:
      config.fraud_protection?.decision?.action ?? "record_only",
    ipAllowlist:
      config.fraud_protection?.decision?.always_allow?.ip_address?.cidrs?.join(
        "\n"
      ) ?? "",
    phoneAllowlist:
      config.fraud_protection?.decision?.always_allow?.phone_number?.regex
        ?.map(toDisplayPhoneAllowlistItem)
        .join("\n") ?? "",
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

    clearEmptyObject(draft);
  });
}

interface FraudProtectionConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  fraudProtectionFeatureConfig?: FraudProtectionFeatureConfig;
}

type FraudProtectionTab = "overview" | "logs" | "settings";

const FraudProtectionConfigurationContent: React.VFC<FraudProtectionConfigurationContentProps> =
  function FraudProtectionConfigurationContent(props) {
    const { form, fraudProtectionFeatureConfig } = props;
    const { renderToString } = useContext(Context);
    const { state, setState } = form;
    const isModifiable =
      fraudProtectionFeatureConfig?.is_modifiable ?? false;
    const { selectedKey, onLinkClick } =
      usePivotNavigation<FraudProtectionTab>([
        "overview",
        "logs",
        "settings",
      ]);

    const enforcementModeOptions = useMemo<IChoiceGroupOption[]>(() => {
      return [
        {
          key: "record_only",
          text: renderToString(
            "FraudProtectionConfigurationScreen.enforcement.observe.label"
          ),
        },
        {
          key: "deny_if_any_warning",
          text: renderToString(
            "FraudProtectionConfigurationScreen.enforcement.block.label"
          ),
        },
      ];
    }, [renderToString]);

    const onEnableChange = useCallback(
      (_event, checked?: boolean) => {
        setState((current) => ({
          ...current,
          enabled: checked ?? false,
        }));
      },
      [setState]
    );

    const onEnforcementModeChange = useCallback(
      (_event, option?: IChoiceGroupOption) => {
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
      (_event, value?: string) => {
        setState((current) => ({
          ...current,
          ipAllowlist: value ?? "",
        }));
      },
      [setState]
    );

    const onPhoneAllowlistChange = useCallback(
      (_event, value?: string) => {
        setState((current) => ({
          ...current,
          phoneAllowlist: value ?? "",
        }));
      },
      [setState]
    );

    return (
      <ScreenContent>
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
          <div className={styles.page}>
            <div className={styles.enableSection}>
              <Toggle
                checked={state.enabled}
                disabled={!isModifiable}
                onChange={onEnableChange}
                label={renderToString(
                  "FraudProtectionConfigurationScreen.enable.label"
                )}
              />
            </div>
            {state.enabled ? (
              <div className={styles.settings}>
                <AGPivot
                  selectedKey={selectedKey}
                  onLinkClick={onLinkClick}
                >
                  <PivotItem headerText="Overview" itemKey="overview" />
                  <PivotItem headerText="Logs" itemKey="logs" />
                  <PivotItem headerText="Settings" itemKey="settings" />
                </AGPivot>
                {selectedKey === "overview" ? (
                  <section className={styles.firstSection}>
                    <Text as="h2" variant="xLarge" block={true}>
                      <FormattedMessage id="FraudProtectionConfigurationScreen.tab.overview.title" />
                    </Text>
                    <Text block={true}>
                      <FormattedMessage id="FraudProtectionConfigurationScreen.tab.overview.description" />
                    </Text>
                  </section>
                ) : null}
                {selectedKey === "logs" ? (
                  <section className={styles.firstSection}>
                    <Text as="h2" variant="xLarge" block={true}>
                      <FormattedMessage id="FraudProtectionConfigurationScreen.tab.logs.title" />
                    </Text>
                    <Text block={true}>
                      <FormattedMessage id="FraudProtectionConfigurationScreen.tab.logs.description" />
                    </Text>
                  </section>
                ) : null}
                {selectedKey === "settings" ? (
                  <>
                    <section className={styles.firstSection}>
                      <Text as="h2" variant="xLarge" block={true}>
                        <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.title" />
                      </Text>
                      <ChoiceGroup
                        disabled={!isModifiable}
                        selectedKey={state.enforcementMode}
                        options={enforcementModeOptions}
                        onChange={onEnforcementModeChange}
                      />
                    </section>
                    <section className={styles.section}>
                      <Text as="h2" variant="xLarge" block={true}>
                        <FormattedMessage id="FraudProtectionConfigurationScreen.allowlist.title" />
                      </Text>
                      <FormTextField
                        className={styles.field}
                        parentJSONPointer="/fraud_protection/decision/always_allow/ip_address"
                        fieldName="cidrs"
                        label={renderToString(
                          "FraudProtectionConfigurationScreen.allowlist.ip.label"
                        )}
                        description={renderToString(
                          "FraudProtectionConfigurationScreen.allowlist.ip.description"
                        )}
                        placeholder="127.0.0.1/32"
                        multiline={true}
                        resizable={false}
                        disabled={!isModifiable}
                        value={state.ipAllowlist}
                        onChange={onIPAllowlistChange}
                      />
                      <FormTextField
                        className={styles.field}
                        parentJSONPointer="/fraud_protection/decision/always_allow/phone_number"
                        fieldName="regex"
                        label={renderToString(
                          "FraudProtectionConfigurationScreen.allowlist.phone.label"
                        )}
                        description={renderToString(
                          "FraudProtectionConfigurationScreen.allowlist.phone.description"
                        )}
                        placeholder="+1 555 123 4567"
                        multiline={true}
                        resizable={false}
                        disabled={!isModifiable}
                        value={state.phoneAllowlist}
                        onChange={onPhoneAllowlistChange}
                      />
                    </section>
                  </>
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
    });
    const featureConfig = useAppFeatureConfigQuery(appID);

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
      >
        <FraudProtectionConfigurationContent
          form={form}
          fraudProtectionFeatureConfig={
            featureConfig.effectiveFeatureConfig?.fraud_protection
          }
        />
      </FormContainer>
    );
  };

export default FraudProtectionConfigurationScreen;
