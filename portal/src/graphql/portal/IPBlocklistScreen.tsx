import React, {
  useCallback,
  useState,
  useEffect,
  useRef,
  useMemo,
} from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import ScreenContent from "../../ScreenContent";
import styles from "./IPBlocklistScreen.module.css";
import FormContainer from "../../FormContainer";
import { useCheckIPMutation } from "./mutations/checkIPMutation";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useParams } from "react-router-dom";
import { PortalAPIAppConfig } from "../../types";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  IPBlocklistForm,
  IPBlocklistFormState,
  toCIDRs,
  IPCheckResult,
} from "../../components/ipblocklist/IPBlocklistForm";
import { IPBlocklistCheckIPPanel } from "../../components/ipblocklist/IPBlocklistCheckIPPanel";
import { produce } from "immer";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { useFormContainerBaseContext } from "../../FormContainerBase";

const IP_FILTER_PORTAL_RULE_NAME = "__portal";

interface FormState extends IPBlocklistFormState {}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const ipFilter = config.network_protection?.ip_filter;

  if (ipFilter?.rules == null || ipFilter.rules.length === 0) {
    return {
      isEditAllowed: true,
      isEnabled: false,
      blockedIPCIDRs: "",
      blockedCountryAlpha2s: [],
    };
  }

  const portalRule = ipFilter.rules.find(
    (rule) => rule.name === IP_FILTER_PORTAL_RULE_NAME
  );

  if (ipFilter.rules.length > 1 || portalRule?.action !== "deny") {
    return {
      isEditAllowed: false,
      isEnabled: false,
      blockedIPCIDRs: "",
      blockedCountryAlpha2s: [],
    };
  }

  const isEnabled = true;
  const blockedIPCIDRs = portalRule.source.cidrs?.join("\n") ?? "";
  const blockedCountryAlpha2s = portalRule.source.geo_location_codes ?? [];

  return {
    isEditAllowed: true,
    isEnabled,
    blockedIPCIDRs,
    blockedCountryAlpha2s,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  if (!currentState.isEditAllowed) {
    return config;
  }
  if (!currentState.isEnabled) {
    return produce(config, (draft) => {
      draft.network_protection ??= {};
      draft.network_protection.ip_filter = {};
    });
  }

  return produce(config, (draft) => {
    draft.network_protection ??= {};
    draft.network_protection.ip_filter ??= {};
    draft.network_protection.ip_filter.default_action = "allow";
    draft.network_protection.ip_filter.rules = [
      {
        name: IP_FILTER_PORTAL_RULE_NAME,
        action: "deny",
        source: {
          cidrs: toCIDRs(currentState.blockedIPCIDRs),
          geo_location_codes: currentState.blockedCountryAlpha2s,
        },
      },
    ];
  });
}

interface IPBlocklistScreenContentProps {
  form: AppConfigFormModel<FormState>;
  appID: string;
  setCheckIPError: (error: unknown) => void;
}

const IPBlocklistScreenContent: React.VFC<IPBlocklistScreenContentProps> =
  function IPBlocklistScreenContent({ form, appID, setCheckIPError }) {
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const {
      checkIP,
      loading: checkingIP,
      error: checkIPMutationError,
    } = useCheckIPMutation(appID);
    const [ipToCheck, setIPToCheck] = useState("");
    const [checkIPResult, setCheckIPResult] = useState<IPCheckResult | null>(
      null
    );

    useEffect(() => {
      setCheckIPError(checkIPMutationError);
    }, [checkIPMutationError, setCheckIPError]);

    const onIPToCheckChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setIPToCheck(e.target.value);
      },
      []
    );

    const onCheckIP = useCallback(() => {
      setCheckIPError(null);
      checkIP(
        ipToCheck,
        toCIDRs(form.state.blockedIPCIDRs),
        form.state.blockedCountryAlpha2s
      )
        .then((result) => {
          setCheckIPResult({
            ipAddress: ipToCheck,
            result: Boolean(result),
          });
        })
        .catch(() => {});
    }, [
      checkIP,
      ipToCheck,
      form.state.blockedIPCIDRs,
      form.state.blockedCountryAlpha2s,
      setCheckIPError,
    ]);

    const [prevBlocklist, setPrevBlocklist] = useState({
      blockedIPCIDRs: form.state.blockedIPCIDRs,
      blockedCountryAlpha2s: form.state.blockedCountryAlpha2s,
    });
    if (
      prevBlocklist.blockedIPCIDRs !== form.state.blockedIPCIDRs ||
      prevBlocklist.blockedCountryAlpha2s !== form.state.blockedCountryAlpha2s
    ) {
      setPrevBlocklist({
        blockedIPCIDRs: form.state.blockedIPCIDRs,
        blockedCountryAlpha2s: form.state.blockedCountryAlpha2s,
      });
      setCheckIPResult(null);
    }

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="IPBlocklistScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="IPBlocklistScreen.description" />
          </Text>
        </div>

        <SettingsSectionCard
          className={cn(
            styles.widget,
            form.state.isEditAllowed &&
              !form.state.isEnabled &&
              styles.settingsCardAlignCenter
          )}
          contentClassName="gap-4"
          title={<FormattedMessage id="IPBlocklistScreen.settings.label" />}
        >
          <IPBlocklistForm state={form.state} setState={form.setState} />
        </SettingsSectionCard>

        {form.state.isEditAllowed && form.state.isEnabled ? (
          <SettingsSectionCard
            className={cn(
              styles.widget,
              isDirty && styles.settingsCardSaveBarClearance
            )}
            contentClassName="gap-4"
            title={<FormattedMessage id="IPBlocklistScreen.check-ip.label" />}
          >
            <IPBlocklistCheckIPPanel
              ipToCheck={ipToCheck}
              onIPToCheckChange={onIPToCheckChange}
              onCheckIP={onCheckIP}
              checkingIP={checkingIP}
              checkIPResult={checkIPResult}
            />
          </SettingsSectionCard>
        ) : null}

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const IPBlocklistScreen: React.FC = function IPBlocklistScreen() {
  const { appID } = useParams() as { appID: string };
  const [checkIPError, setCheckIPError] = useState<unknown>(null);

  const form = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });

  const clearCheckIPError = useCallback(async () => {
    setCheckIPError(null);
  }, []);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer
      form={form}
      canSave={true}
      localError={checkIPError}
      beforeSave={clearCheckIPError}
    >
      <IPBlocklistScreenContent
        form={form}
        appID={appID}
        setCheckIPError={setCheckIPError}
      />
    </FormContainer>
  );
};

export default IPBlocklistScreen;
