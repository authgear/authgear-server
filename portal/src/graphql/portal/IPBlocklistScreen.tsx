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
} from "../../components/ipblocklist/IPBlocklistForm";

interface FormState extends IPBlocklistFormState {}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {
    isEnabled: false,
    blockedIPCIDRs: "",
    blockedCountryAlpha2s: [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  _currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return {
    ...config,
  };
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
