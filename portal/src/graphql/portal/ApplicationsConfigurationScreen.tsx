import React, { useCallback, useContext, useMemo } from "react";
import { Heading } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import styles from "./ApplicationsConfigurationScreen.module.css";
import ScreenContent from "../../ScreenContent";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FormContainer from "../../FormContainer";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { usePivotNavigation } from "../../hook/usePivot";
import { smallestBlockQuota } from "../../util/usageLimit";
import { DynamicClientsTab } from "../../components/dynamic-clients/DynamicClientsTab";
import {
  ApplicationsListSection,
  FormState,
  constructConfig,
  constructFormState,
} from "./ApplicationsListSection";

// The dynamic-clients tab holds only what CIMD and DCR share; configuring
// either mechanism happens on its own nav-level screen.
type ApplicationsTabKey = "applications" | "dynamic-clients";

// Module-level so its identity is stable across renders: usePivotNavigation
// keeps it in an effect's dependency list.
const APPLICATIONS_TAB_KEYS: ApplicationsTabKey[] = [
  "applications",
  "dynamic-clients",
];

interface OAuthClientConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  planName: string | null;
  oauthClientsSoftMaximum: number | undefined;
  oauthClientsHardMaximum: number | undefined;
  selectedKey: ApplicationsTabKey;
  onChangeKey: (key: ApplicationsTabKey) => void;
  dcrClientQuota: number | null;
  cimdClientQuota: number | null;
}

const OAuthClientConfigurationContent: React.VFC<OAuthClientConfigurationContentProps> =
  function OAuthClientConfigurationContent(props) {
    const {
      form,
      planName,
      oauthClientsHardMaximum,
      oauthClientsSoftMaximum,
      selectedKey,
      onChangeKey,
      dcrClientQuota,
      cimdClientQuota,
    } = props;
    const { renderToString } = useContext(Context);

    const onTabChange = useCallback(
      (value: string) => {
        onChangeKey(value as ApplicationsTabKey);
      },
      [onChangeKey]
    );

    const tabOptions = useMemo(
      () => [
        {
          value: "applications",
          label: renderToString(
            "ApplicationsConfigurationScreen.tab.applications"
          ),
        },
        {
          value: "dynamic-clients",
          label: renderToString(
            "ApplicationsConfigurationScreen.tab.dynamic-clients"
          ),
        },
      ],
      [renderToString]
    );

    const cimdEnabled =
      form.effectiveConfig.oauth?.client_id_metadata_document?.enabled ?? false;
    const dcrEnabled =
      form.effectiveConfig.oauth?.dynamic_client_registration?.enabled ?? false;

    return (
      <ScreenContent layout="list">
        <div className={styles.pageHeader}>
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          </Heading>
        </div>
        <OverflowTabs
          className={styles.widget}
          value={selectedKey}
          onValueChange={onTabChange}
          tabs={tabOptions}
        />
        {selectedKey === "dynamic-clients" ? (
          <div className={styles.widget}>
            <DynamicClientsTab
              cimdClientQuota={cimdClientQuota}
              dcrClientQuota={dcrClientQuota}
              cimdEnabled={cimdEnabled}
              dcrEnabled={dcrEnabled}
            />
          </div>
        ) : (
          <ApplicationsListSection
            form={form}
            planName={planName}
            oauthClientsHardMaximum={oauthClientsHardMaximum}
            oauthClientsSoftMaximum={oauthClientsSoftMaximum}
          />
        )}
      </ScreenContent>
    );
  };

const ApplicationsConfigurationScreen: React.VFC =
  function ApplicationsConfigurationScreen() {
    const { appID } = useParams() as { appID: string };

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);
    const { selectedKey, onChangeKey } = usePivotNavigation<ApplicationsTabKey>(
      APPLICATIONS_TAB_KEYS
    );

    const dcrClientQuota = useMemo<number | null>(() => {
      return smallestBlockQuota(
        featureConfig.effectiveFeatureConfig?.usage?.limits?.oauth_client_dcr
      );
    }, [featureConfig.effectiveFeatureConfig]);

    const cimdClientQuota = useMemo<number | null>(() => {
      return smallestBlockQuota(
        featureConfig.effectiveFeatureConfig?.usage?.limits?.oauth_client_cimd
      );
    }, [featureConfig.effectiveFeatureConfig]);

    const oauthClientsHardMaximum = useMemo<number | undefined>(() => {
      return featureConfig.effectiveFeatureConfig?.oauth?.client?.maximum;
    }, [featureConfig]);

    const oauthClientsSoftMaximum = useMemo(() => {
      return featureConfig.effectiveFeatureConfig?.oauth?.client?.soft_maximum;
    }, [featureConfig]);

    const isLoading = useMemo(
      () => form.isLoading || featureConfig.isLoading,
      [form.isLoading, featureConfig.isLoading]
    );

    const error = useMemo(
      () => form.loadError ?? featureConfig.loadError,
      [form.loadError, featureConfig.loadError]
    );

    const onRetry = useCallback(() => {
      if (form.loadError) {
        form.reload();
      }

      if (featureConfig.loadError) {
        featureConfig.refetch().finally(() => {});
      }
    }, [form, featureConfig]);

    if (isLoading) {
      return <ShowLoading />;
    }

    if (error) {
      return <ShowError error={error} onRetry={onRetry} />;
    }

    return (
      <FormContainer form={form}>
        <OAuthClientConfigurationContent
          form={form}
          planName={featureConfig.planName}
          oauthClientsHardMaximum={oauthClientsHardMaximum}
          oauthClientsSoftMaximum={oauthClientsSoftMaximum}
          selectedKey={selectedKey}
          onChangeKey={onChangeKey}
          dcrClientQuota={dcrClientQuota}
          cimdClientQuota={cimdClientQuota}
        />
      </FormContainer>
    );
  };

export default ApplicationsConfigurationScreen;
