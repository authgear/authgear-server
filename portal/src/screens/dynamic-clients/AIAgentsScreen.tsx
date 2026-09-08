import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Heading, Text } from "@radix-ui/themes";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import PortalLink from "../../Link";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { useAppFeatureConfigQuery } from "../../graphql/portal/query/appFeatureConfigQuery";
import { usePivotNavigation } from "../../hook/usePivot";
import { smallestBlockQuota } from "../../util/usageLimit";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import {
  ClientCountCard,
  ClientCountSource,
} from "../../components/dynamic-clients/ClientCountCard";
import { DynamicClientAllowedResources } from "../../components/dynamic-clients/DynamicClientAllowedResources";
import {
  MetadataDocumentsContent,
  constructConfig as constructCIMDConfig,
  constructFormState as constructCIMDFormState,
} from "../../components/dynamic-clients/MetadataDocumentsContent";
import {
  SelfRegistrationContent,
  constructConfig as constructDCRConfig,
  constructFormState as constructDCRFormState,
} from "../../components/dynamic-clients/SelfRegistrationContent";
import styles from "./Applications.module.css";

type TabKey = "overview" | "cimd" | "dcr";

// Module-level so its identity is stable: usePivotNavigation keeps it in an
// effect's dependency list.
const TAB_KEYS: TabKey[] = ["overview", "cimd", "dcr"];

interface TabOption {
  value: string;
  label: string;
}

interface TabBarProps {
  tabs: TabOption[];
  value: TabKey;
  onValueChange: (key: TabKey) => void;
}

const TabBar: React.VFC<TabBarProps> = function TabBar({
  tabs,
  value,
  onValueChange,
}) {
  return (
    <OverflowTabs
      className={styles.widget}
      value={value}
      onValueChange={onValueChange as (value: string) => void}
      tabs={tabs}
    />
  );
};

/**
 * The tab bar as rendered inside a tab that owns a form. Switching tab
 * unmounts that form, so pending edits would go without a word; this asks
 * first. Dirtiness comes from the FormContainerBase context rather than a
 * prop, so it can only ever report the form it is inside.
 */
const GuardedTabBar: React.VFC<TabBarProps> = function GuardedTabBar({
  tabs,
  value,
  onValueChange,
}) {
  const { getIsDirty } = useFormContainerBaseContext();
  const [pendingTab, setPendingTab] = useState<TabKey | null>(null);

  const onRequestChange = useCallback(
    (key: TabKey) => {
      if (getIsDirty()) {
        setPendingTab(key);
        return;
      }
      onValueChange(key);
    },
    [getIsDirty, onValueChange]
  );

  const onConfirmDiscard = useCallback(() => {
    const key = pendingTab;
    setPendingTab(null);
    if (key != null) {
      onValueChange(key);
    }
  }, [pendingTab, onValueChange]);

  const onCancelDiscard = useCallback(() => {
    setPendingTab(null);
  }, []);

  // onOpenChange fires in both directions; only a dismissal should drop the
  // pending tab.
  const onOpenChange = useCallback((open: boolean) => {
    if (!open) {
      setPendingTab(null);
    }
  }, []);

  return (
    <>
      <TabBar tabs={tabs} value={value} onValueChange={onRequestChange} />
      <ConfirmationDialog
        open={pendingTab != null}
        onOpenChange={onOpenChange}
        title={<FormattedMessage id="FormContainer.reset-dialog.title" />}
        description={
          <FormattedMessage id="FormContainer.reset-dialog.message" />
        }
        confirmText={
          <FormattedMessage id="FormContainer.reset-dialog.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        confirmColor="red"
        onConfirm={onConfirmDiscard}
        onCancel={onCancelDiscard}
      />
    </>
  );
};

const PageHeader: React.VFC = function PageHeader() {
  return (
    <div className={cn(styles.widget, styles.pageHeader)}>
      <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
        <FormattedMessage id="AIAgentsScreen.title" />
      </Heading>
      <Text as="p" size="2" color="gray" className={styles.pageDescription}>
        <FormattedMessage id="AIAgentsScreen.description" />
      </Text>
    </div>
  );
};

const OverviewTab: React.VFC = function OverviewTab() {
  const { appID } = useParams() as { appID: string };
  const featureConfig = useAppFeatureConfigQuery(appID);

  const sources = useMemo<ClientCountSource[]>(
    () => [
      {
        source: OAuthClientSource.Cimd,
        quota: smallestBlockQuota(
          featureConfig.effectiveFeatureConfig?.usage?.limits?.oauth_client_cimd
        ),
      },
      {
        source: OAuthClientSource.Dcr,
        quota: smallestBlockQuota(
          featureConfig.effectiveFeatureConfig?.usage?.limits?.oauth_client_dcr
        ),
      },
    ],
    [featureConfig.effectiveFeatureConfig]
  );

  return (
    <div className={cn(styles.widget, styles.stack)}>
      {/* Absolute: the listing is a sibling of this page under
          configuration/apps, not a child of it. */}
      <ClientCountCard
        sources={sources}
        listPath={`/project/${appID}/configuration/apps/dynamic-clients`}
      />
      <SettingsSectionCard
        contentClassName="gap-4"
        title={<FormattedMessage id="AIAgentsScreen.allowed-resources.title" />}
        description={
          <FormattedMessage id="AIAgentsScreen.allowed-resources.description" />
        }
      >
        <DynamicClientAllowedResources />
        <Text as="p" size="2" color="gray">
          <FormattedMessage
            id="AIAgentsScreen.allowed-resources.manage"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              apiResourcesLink: (chunks: React.ReactNode) => (
                <PortalLink to={`/project/${appID}/api-resources`}>
                  {chunks}
                </PortalLink>
              ),
            }}
          />
        </Text>
      </SettingsSectionCard>
    </div>
  );
};

interface MechanismTabProps {
  tabBarProps: TabBarProps;
}

const DCRTab: React.VFC<MechanismTabProps> = function DCRTab({ tabBarProps }) {
  const { appID } = useParams() as { appID: string };
  // Its own form: a save from this tab must not carry the sibling
  // mechanism's edits.
  const form = useAppConfigForm({
    appID,
    constructFormState: constructDCRFormState,
    constructConfig: constructDCRConfig,
  });

  if (form.isLoading) {
    return <ShowLoading />;
  }
  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <ScreenContent layout="list">
        <PageHeader />
        <GuardedTabBar {...tabBarProps} />
        <SelfRegistrationContent form={form} />
      </ScreenContent>
    </FormContainer>
  );
};

const CIMDTab: React.VFC<MechanismTabProps> = function CIMDTab({
  tabBarProps,
}) {
  const { appID } = useParams() as { appID: string };
  const form = useAppConfigForm({
    appID,
    constructFormState: constructCIMDFormState,
    constructConfig: constructCIMDConfig,
  });

  if (form.isLoading) {
    return <ShowLoading />;
  }
  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <ScreenContent layout="list">
        <PageHeader />
        <GuardedTabBar {...tabBarProps} />
        <MetadataDocumentsContent form={form} />
      </ScreenContent>
    </FormContainer>
  );
};

const AIAgentsScreen: React.VFC = function AIAgentsScreen() {
  const { renderToString } = useContext(Context);
  const { selectedKey, onChangeKey } = usePivotNavigation<TabKey>(TAB_KEYS);

  const tabs = useMemo<TabOption[]>(
    () => [
      {
        value: "overview",
        label: renderToString("AIAgentsScreen.tab.overview"),
      },
      { value: "cimd", label: renderToString("AIAgentsScreen.tab.cimd") },
      { value: "dcr", label: renderToString("AIAgentsScreen.tab.dcr") },
    ],
    [renderToString]
  );

  const tabBarProps: TabBarProps = {
    tabs,
    value: selectedKey,
    onValueChange: onChangeKey,
  };

  switch (selectedKey) {
    case "cimd":
      return <CIMDTab tabBarProps={tabBarProps} />;
    case "dcr":
      return <DCRTab tabBarProps={tabBarProps} />;
    default:
      // Overview owns no form, so it needs no FormContainer and no guard.
      return (
        <ScreenContent layout="list">
          <PageHeader />
          <TabBar {...tabBarProps} />
          <OverviewTab />
        </ScreenContent>
      );
  }
};

export default AIAgentsScreen;
