import React, { useCallback, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import PortalLink from "../../Link";
import ExternalLink from "../../ExternalLink";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { Badge } from "../v2/Badge/Badge";
import { ClientCountCard, ClientCountSource } from "./ClientCountCard";
import { DynamicClientAllowedResources } from "./DynamicClientAllowedResources";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./DynamicClientsTab.module.css";

export interface DynamicClientsTabProps {
  // Whether each mechanism is currently accepting clients, so the summary
  // answers "what is switched on here" without a trip to its page.
  cimdEnabled: boolean;
  dcrEnabled: boolean;
  // The smallest block-action quota configured for oauth_client_cimd, or null
  // when the plan does not limit CIMD-resolved clients.
  cimdClientQuota: number | null;
  // Likewise for oauth_client_dcr.
  dcrClientQuota: number | null;
}

/**
 * What CIMD and DCR share: how many clients each has produced, and the API
 * resources every dynamic third-party client can reach.
 *
 * Each mechanism's own settings live on its own nav-level screen, so this tab
 * links to them rather than embedding them -- otherwise it would be a summary
 * with no way onwards.
 */
export const DynamicClientsTab: React.VFC<DynamicClientsTabProps> =
  function DynamicClientsTab({
    cimdClientQuota,
    dcrClientQuota,
    cimdEnabled,
    dcrEnabled,
  }) {
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();

    const onConfigureMetadataDocuments = useCallback(() => {
      navigate("./metadata-documents");
    }, [navigate]);

    const onConfigureSelfRegistration = useCallback(() => {
      navigate("./self-registration");
    }, [navigate]);

    const sources = useMemo<ClientCountSource[]>(
      () => [
        { source: OAuthClientSource.Cimd, quota: cimdClientQuota },
        { source: OAuthClientSource.Dcr, quota: dcrClientQuota },
      ],
      [cimdClientQuota, dcrClientQuota]
    );

    return (
      <div className={styles.root}>
        <ClientCountCard sources={sources} />

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

        {/* Modelled on the Danger zone card (UserDetailsAccountStatus): one
            tinted card whose rows are divider-separated actions. The tint is
            what marks these two out as the mechanisms that let clients in,
            rather than two more settings among the summaries above. */}
        <section className={styles.mechanismCard}>
          <Text
            as="p"
            size="3"
            weight="medium"
            className={styles.mechanismCardTitle}
          >
            <FormattedMessage id="DynamicClientsTab.mechanisms.title" />
          </Text>
          <div className={styles.mechanismCardContent}>
            <div className={styles.mechanismRow}>
              <Text as="p" size="2" className={styles.mechanismRowLabel}>
                <FormattedMessage id="MetadataDocumentsScreen.title" />
              </Text>
              <Text as="p" size="2" className={styles.mechanismRowBody}>
                <FormattedMessage
                  id="CIMDSection.enable.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    mcpLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://docs.authgear.com/get-started/auth-for-mcp">
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
              </Text>
              <div className={styles.mechanismRowStatus}>
                <Badge
                  size="1"
                  variant={cimdEnabled ? "success" : "neutral"}
                  text={
                    <FormattedMessage
                      id={cimdEnabled ? "enabled" : "disabled"}
                    />
                  }
                />
              </div>
              <div className={styles.mechanismRowAction}>
                <PrimaryButton
                  size="2"
                  text={<FormattedMessage id="DynamicClientsTab.configure" />}
                  onClick={onConfigureMetadataDocuments}
                />
              </div>
            </div>
            <div className={styles.mechanismRow}>
              <Text as="p" size="2" className={styles.mechanismRowLabel}>
                <FormattedMessage id="SelfRegistrationScreen.title" />
              </Text>
              <Text as="p" size="2" className={styles.mechanismRowBody}>
                <FormattedMessage
                  id="DynamicClientsTab.enable.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    mcpLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://docs.authgear.com/get-started/auth-for-mcp">
                        {chunks}
                      </ExternalLink>
                    ),
                    // eslint-disable-next-line react/no-unstable-nested-components
                    dcrLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://docs.authgear.com/integration/dynamic-client-registration">
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
              </Text>
              <div className={styles.mechanismRowStatus}>
                <Badge
                  size="1"
                  variant={dcrEnabled ? "success" : "neutral"}
                  text={
                    <FormattedMessage
                      id={dcrEnabled ? "enabled" : "disabled"}
                    />
                  }
                />
              </div>
              <div className={styles.mechanismRowAction}>
                <PrimaryButton
                  size="2"
                  text={<FormattedMessage id="DynamicClientsTab.configure" />}
                  onClick={onConfigureSelfRegistration}
                />
              </div>
            </div>
          </div>
        </section>
      </div>
    );
  };
