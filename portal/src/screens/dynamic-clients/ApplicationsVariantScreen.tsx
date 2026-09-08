/**
 * SIDE-BY-SIDE EXPERIMENT — delete this file, its CSS, its route and its nav
 * entry once one layout wins.
 *
 * An alternative to the shipped arrangement, reachable at the same time so
 * the two can be compared in a running portal. The shipped one spreads this
 * across four places: the Applications screen's two tabs
 * (ApplicationsConfigurationScreen), and a nav page per mechanism
 * (MetadataDocumentsScreen, SelfRegistrationScreen). This one is a single
 * first-level page: two tabs, and inside the second, the two mechanisms as
 * nested tabs under the sections they share.
 *
 * Every part is the real thing rather than a mock: the client list, the count
 * card, the allowed-resources card and both mechanism sections are the same
 * components the shipped screens render, so what differs between the two
 * designs is only the arrangement.
 *
 * Each outer tab is its own component owning FormContainer > ScreenContent,
 * in that order. FormContainer renders a whole page shell (DefaultLayout), so
 * it has to be outermost -- putting it inside ScreenContent drops a page
 * layout into a grid cell and the tab renders empty. That is also why the
 * page chrome is passed down into each tab rather than rendered above them:
 * only one form may be mounted at a time, and the mechanism tabs each need
 * their own.
 */
import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import {
  FileTextIcon,
  LayersIcon,
  PlusCircledIcon,
} from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import PortalLink from "../../Link";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { useAppFeatureConfigQuery } from "../../graphql/portal/query/appFeatureConfigQuery";
import { usePivotNavigation } from "../../hook/usePivot";
import { useScreenBreakpoint } from "../../hook/useScreenBreakpoint";
import { smallestBlockQuota } from "../../util/usageLimit";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import {
  ApplicationsListSection,
  constructConfig as constructClientsConfig,
  constructFormState as constructClientsFormState,
} from "../../graphql/portal/ApplicationsListSection";
import {
  ClientCountCard,
  ClientCountSource,
} from "../../components/dynamic-clients/ClientCountCard";
import { DynamicClientAllowedResources } from "../../components/dynamic-clients/DynamicClientAllowedResources";
import {
  MetadataDocumentsContent,
  constructConfig as constructCIMDConfig,
  constructFormState as constructCIMDFormState,
} from "./MetadataDocumentsScreen";
import {
  SelfRegistrationContent,
  constructConfig as constructDCRConfig,
  constructFormState as constructDCRFormState,
} from "./SelfRegistrationScreen";
import styles from "./ApplicationsVariantScreen.module.css";

type TabKey = "applications" | "dynamic-clients";
type ModeKey = "overview" | "cimd" | "dcr";

// Module-level so their identity is stable: usePivotNavigation keeps them in
// an effect's dependency list.
const TAB_KEYS: TabKey[] = ["applications", "dynamic-clients"];
const MODE_KEYS: ModeKey[] = ["overview", "cimd", "dcr"];

// The outer tabs live in the hash, so the rail needs a slot of its own or
// the two levels overwrite each other.
const MODE_SEARCH_PARAM = "mode";

// Matches the icon size the Email/SMS Templates screen's channel cards use.
const MODE_ICON_SIZE = "1.375rem";

interface ModeRailProps {
  options: IconRadioCardOption<ModeKey>[];
  value: ModeKey;
  onValueChange: (mode: ModeKey) => void;
  columns: number;
}

const ModeRail: React.VFC<ModeRailProps> = function ModeRail({
  options,
  value,
  onValueChange,
  columns,
}) {
  return (
    <IconRadioCards
      size="3"
      numberOfColumns={columns}
      itemFillSpaces={true}
      value={value}
      onValueChange={onValueChange}
      options={options}
    />
  );
};

/**
 * The rail as rendered inside a mode that owns a form. Switching mode
 * unmounts that form, so pending edits would go without a word; this asks
 * first. It reads dirtiness from the FormContainerBase context rather than
 * taking it as a prop, so it can only ever report the form it is inside.
 */
const GuardedModeRail: React.VFC<ModeRailProps> = function GuardedModeRail({
  options,
  value,
  onValueChange,
  columns,
}) {
  const { getIsDirty } = useFormContainerBaseContext();
  const [pendingMode, setPendingMode] = useState<ModeKey | null>(null);

  const onRequestChange = useCallback(
    (mode: ModeKey) => {
      if (getIsDirty()) {
        setPendingMode(mode);
        return;
      }
      onValueChange(mode);
    },
    [getIsDirty, onValueChange]
  );

  const onConfirmDiscard = useCallback(() => {
    const mode = pendingMode;
    setPendingMode(null);
    if (mode != null) {
      onValueChange(mode);
    }
  }, [pendingMode, onValueChange]);

  const onCancelDiscard = useCallback(() => {
    setPendingMode(null);
  }, []);

  // onOpenChange fires in both directions; only a dismissal should drop the
  // pending mode.
  const onOpenChange = useCallback((open: boolean) => {
    if (!open) {
      setPendingMode(null);
    }
  }, []);

  return (
    <>
      <ModeRail
        options={options}
        value={value}
        onValueChange={onRequestChange}
        columns={columns}
      />
      <ConfirmationDialog
        open={pendingMode != null}
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

interface TabBodyProps {
  // The page title and the outer tab bar, rendered by each tab inside its own
  // ScreenContent.
  chrome: React.ReactNode;
}

const ApplicationsTab: React.VFC<TabBodyProps> = function ApplicationsTab({
  chrome,
}) {
  const { appID } = useParams() as { appID: string };
  const form = useAppConfigForm({
    appID,
    constructFormState: constructClientsFormState,
    constructConfig: constructClientsConfig,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  if (form.isLoading || featureConfig.isLoading) {
    return <ShowLoading />;
  }
  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <ScreenContent layout="list">
        {chrome}
        <ApplicationsListSection
          form={form}
          planName={featureConfig.planName}
          oauthClientsSoftMaximum={
            featureConfig.effectiveFeatureConfig?.oauth?.client?.soft_maximum
          }
          oauthClientsHardMaximum={
            featureConfig.effectiveFeatureConfig?.oauth?.client?.maximum
          }
        />
      </ScreenContent>
    </FormContainer>
  );
};

const SharedSections: React.VFC = function SharedSections() {
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
    <div className={cn(styles.widget, styles.shared)}>
      {/* Absolute: the list lives beside the shipped Applications screen, not
          under this variant's route. */}
      <ClientCountCard
        sources={sources}
        listPath={`/project/${appID}/configuration/apps/dynamic-clients`}
      />
      <SettingsSectionCard
        contentClassName="gap-4"
        title={
          <FormattedMessage id="DynamicClientsTab.allowed-resources.title" />
        }
        description={
          <FormattedMessage id="DynamicClientsTab.allowed-resources.description" />
        }
      >
        <DynamicClientAllowedResources />
        <Text as="p" size="2" color="gray">
          <FormattedMessage
            id="DynamicClientsTab.allowed-resources.manage"
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

interface ModeProps extends TabBodyProps {
  railProps: ModeRailProps;
}

// Overview needs no form of its own -- both cards read their own data.
const OverviewMode: React.VFC<ModeProps> = function OverviewMode({
  chrome,
  railProps,
}) {
  return (
    <ScreenContent layout="list">
      {chrome}
      <div className={cn(styles.widget, styles.modeLayout)}>
        <div className={styles.modeRail}>
          <ModeRail {...railProps} />
        </div>
        <div className={styles.modeContent}>
          <SharedSections />
        </div>
      </div>
    </ScreenContent>
  );
};

const CIMDMode: React.VFC<ModeProps> = function CIMDMode({
  chrome,
  railProps,
}) {
  const { appID } = useParams() as { appID: string };
  // Its own form, like the nav-level screen has: a save from this tab must
  // not carry the sibling mechanism's edits.
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
        {chrome}
        <div className={cn(styles.widget, styles.modeLayout)}>
          <div className={styles.modeRail}>
            <GuardedModeRail {...railProps} />
          </div>
          <div className={styles.modeContent}>
            <MetadataDocumentsContent form={form} />
          </div>
        </div>
      </ScreenContent>
    </FormContainer>
  );
};

const DCRMode: React.VFC<ModeProps> = function DCRMode({ chrome, railProps }) {
  const { appID } = useParams() as { appID: string };
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
        {chrome}
        <div className={cn(styles.widget, styles.modeLayout)}>
          <div className={styles.modeRail}>
            <GuardedModeRail {...railProps} />
          </div>
          <div className={styles.modeContent}>
            <SelfRegistrationContent form={form} />
          </div>
        </div>
      </ScreenContent>
    </FormContainer>
  );
};

const ApplicationsVariantScreen: React.VFC =
  function ApplicationsVariantScreen() {
    const { renderToString } = useContext(Context);
    const screenBreakpoint = useScreenBreakpoint();
    const { selectedKey, onChangeKey } = usePivotNavigation<TabKey>(TAB_KEYS);
    const { selectedKey: selectedMode, onChangeKey: onChangeMode } =
      usePivotNavigation<ModeKey>(MODE_KEYS, undefined, MODE_SEARCH_PARAM);

    const tabs = useMemo(
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

    const modeOptions = useMemo<IconRadioCardOption<ModeKey>[]>(
      () => [
        {
          value: "overview",
          icon: <LayersIcon width={MODE_ICON_SIZE} height={MODE_ICON_SIZE} />,
          title: (
            <FormattedMessage id="ApplicationsVariantScreen.mode.overview" />
          ),
        },
        {
          value: "cimd",
          icon: <FileTextIcon width={MODE_ICON_SIZE} height={MODE_ICON_SIZE} />,
          title: <FormattedMessage id="ApplicationsVariantScreen.mode.cimd" />,
        },
        {
          value: "dcr",
          icon: (
            <PlusCircledIcon width={MODE_ICON_SIZE} height={MODE_ICON_SIZE} />
          ),
          title: <FormattedMessage id="ApplicationsVariantScreen.mode.dcr" />,
        },
      ],
      []
    );

    const chrome = (
      <>
        <div className={cn(styles.widget, styles.pageHeader)}>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="ApplicationsVariantScreen.title" />
          </Text>
        </div>
        <OverflowTabs
          className={styles.widget}
          value={selectedKey}
          onValueChange={onChangeKey as (value: string) => void}
          tabs={tabs}
        />
      </>
    );

    const railProps: ModeRailProps = {
      options: modeOptions,
      value: selectedMode,
      onValueChange: onChangeMode,
      columns: screenBreakpoint === "mobile" ? 3 : 1,
    };

    if (selectedKey === "applications") {
      return <ApplicationsTab chrome={chrome} />;
    }

    switch (selectedMode) {
      case "cimd":
        return <CIMDMode chrome={chrome} railProps={railProps} />;
      case "dcr":
        return <DCRMode chrome={chrome} railProps={railProps} />;
      default:
        return <OverviewMode chrome={chrome} railProps={railProps} />;
    }
  };

export default ApplicationsVariantScreen;
