import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import { FormattedMessage } from "@oursky/react-messageformat";
import React, { useMemo } from "react";
import ScreenContent from "../../ScreenContent";
import styles from "./IPBlocklistScreen.module.css";
import ScreenDescription from "../../ScreenDescription";
import FormContainer from "../../FormContainer";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { useParams } from "react-router-dom";
import { PortalAPIAppConfig } from "../../types";
import {
  IPBlocklistForm,
  IPBlocklistFormState,
  toCIDRs,
} from "../../components/ipblocklist/IPBlocklistForm";
import { produce } from "immer";

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
  const blockedIPCIDRs = portalRule.source.cidrs?.join(",") ?? "";
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

const IPBlocklistScreen: React.FC = function IPBlocklistScreen() {
  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: "~/attack-protection",
        label: <FormattedMessage id="AttackProtectionScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="IPBlocklistScreen.title" /> },
    ];
  }, []);

  const { appID } = useParams() as { appID: string };

  const appConfigForm = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });

  return (
    <FormContainer
      form={appConfigForm}
      canSave={true}
      stickyFooterComponent={true}
      showDiscardButton={true}
    >
      <ScreenContent>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="IPBlocklistScreen.description" />
        </ScreenDescription>
      </ScreenContent>
      <div className={styles.widget}>
        <IPBlocklistForm
          state={appConfigForm.state}
          setState={appConfigForm.setState}
        />
      </div>
    </FormContainer>
  );
};

export default IPBlocklistScreen;
