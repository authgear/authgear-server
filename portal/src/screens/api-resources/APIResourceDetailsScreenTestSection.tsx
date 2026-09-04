import React, { useContext, useMemo, useState, useCallback } from "react";
import { Select, Tabs, Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useNavigate } from "react-router-dom";
import { useEndpoints } from "../../hook/useEndpoints";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import { CodeField } from "../../components/common/CodeField";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import { LocationState } from "./APIResourceDetailsScreen";
import { useSearchParamsState } from "../../hook/useSearchParamsState";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { parseRawError } from "../../error/parse";
import { copyToClipboard } from "../../util/clipboard";
import {
  ExampleCodeVariant,
  useExampleCode,
} from "../../components/api-resources/useExampleCode";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { FormField } from "../../components/v2/FormField/FormField";
import ExternalLink from "../../ExternalLink";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import styles from "./APIResourceDetailsTestSection.module.css";

export function APIResourceDetailsScreenTestSection({
  resource,
  effectiveAppConfig,
  secretConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
}): React.ReactElement | null {
  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();
  const { setErrors } = useErrorMessageBarContext();
  const [selectedClientId, setSelectedClientId] = useSearchParamsState<string>(
    "client",
    ""
  );
  const [isGenerating, setIsGenerating] = useState(false);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [selectedCodeVariant, setSelectedCodeVariant] =
    useState<ExampleCodeVariant>(ExampleCodeVariant.curl);
  const [tokenCopied, setTokenCopied] = useState(false);
  const [codeCopied, setCodeCopied] = useState(false);
  const { startReauthentication, isRevealing } =
    useStartReauthentication<LocationState>();

  const selectedClient = useMemo(() => {
    // The client must still be authorized for this resource: unauthorizing
    // the selected application (in the section above) must clear the whole
    // test section, not just its entry in the dropdown, even though the
    // client itself still exists in the app config.
    if (!resource.clientIDs.includes(selectedClientId)) {
      return undefined;
    }
    return effectiveAppConfig.oauth?.clients?.find(
      (client) => client.client_id === selectedClientId
    );
  }, [effectiveAppConfig, selectedClientId, resource.clientIDs]);

  const [prevSelectedClient, setPrevSelectedClient] = useState(selectedClient);
  if (prevSelectedClient !== selectedClient) {
    setPrevSelectedClient(selectedClient);
    setAccessToken(null);
  }

  const selectedClientSecret = useMemo((): string | null => {
    if (!secretConfig || !selectedClient?.client_id) {
      return null;
    }
    const secret = secretConfig.oauthClientSecrets?.find(
      (secret) => secret.clientID === selectedClient.client_id
    );
    if (secret?.keys != null && secret.keys.length > 0 && secret.keys[0].key) {
      return secret.keys[0].key;
    }
    return null;
  }, [secretConfig, selectedClient]);

  const { token: tokenEndpoint } = useEndpoints(
    effectiveAppConfig.http?.public_origin ?? "",
    selectedClient?.x_application_type
  );

  const authorizedApplicationsOptions = useMemo(() => {
    const authorizedClientIDs = new Set(resource.clientIDs);
    return (
      effectiveAppConfig.oauth?.clients
        ?.filter((clientConfig) => {
          return authorizedClientIDs.has(clientConfig.client_id);
        })
        .map((clientConfig) => ({
          value: clientConfig.client_id,
          label: clientConfig.name ?? clientConfig.client_name ?? "",
        })) ?? []
    );
  }, [effectiveAppConfig.oauth?.clients, resource.clientIDs]);

  const revealSecrets = useCallback(() => {
    startReauthentication(navigate, {
      isClientSecretRevealed: true,
    }).catch((e) => {
      console.error(e);
    });
  }, [navigate, startReauthentication]);

  const onGenerate = useCallback(() => {
    if (selectedClientSecret == null) {
      revealSecrets();
    } else {
      const body = new URLSearchParams();
      body.append("client_id", selectedClientId);
      body.append("grant_type", "client_credentials");
      body.append("resource", resource.resourceURI);
      body.append("client_secret", selectedClientSecret);
      setIsGenerating(true);
      const generate = async () => {
        try {
          const response = await fetch(tokenEndpoint, {
            method: "POST",
            headers: {
              "Content-Type": "application/x-www-form-urlencoded",
            },
            body: body.toString(),
          });

          if (!response.ok) {
            throw new Error(`invalid response status: ${response.status}`);
          }

          const data = await response.json();
          setAccessToken(data.access_token);
        } catch (error) {
          console.error("Error generating access token:", error);
          setErrors(parseRawError(error));
          setAccessToken(null);
        } finally {
          setIsGenerating(false);
        }
      };
      void generate();
    }
  }, [
    selectedClientSecret,
    revealSecrets,
    selectedClientId,
    resource.resourceURI,
    tokenEndpoint,
    setErrors,
  ]);

  const exampleCode = useExampleCode({
    variant: selectedCodeVariant,
    tokenEndpoint,
    resourceURI: resource.resourceURI,
    clientSecret: selectedClientSecret,
    clientID: selectedClientId,
  });

  const onCopyToken = useCallback(() => {
    if (accessToken == null) {
      return;
    }
    copyToClipboard(accessToken);
    setTokenCopied(true);
    window.setTimeout(() => setTokenCopied(false), 2000);
  }, [accessToken]);

  const onCopyCode = useCallback(() => {
    copyToClipboard(exampleCode);
    setCodeCopied(true);
    window.setTimeout(() => setCodeCopied(false), 2000);
  }, [exampleCode]);

  const selectPlaceholder =
    authorizedApplicationsOptions.length === 0
      ? renderToString("APIResourceDetailsScreen.test.selectApplication.empty")
      : renderToString("APIResourceDetailsScreen.test.selectApplication");

  return (
    <div className={styles.root}>
      <SettingsSectionCard
        title={<FormattedMessage id="APIResourceDetailsScreen.section.test" />}
        description={
          <FormattedMessage id="APIResourceDetailsScreen.test.description" />
        }
        contentClassName={styles.cardContent}
      >
        <FormField
          size="2"
          label={
            <FormattedMessage id="APIResourceDetailsScreen.test.authorizedApplications" />
          }
        >
          <Select.Root
            value={selectedClient?.client_id || undefined}
            onValueChange={setSelectedClientId}
            disabled={authorizedApplicationsOptions.length === 0}
            size="2"
          >
            <Select.Trigger
              className={styles.selectTrigger}
              placeholder={selectPlaceholder}
            />
            <Select.Content position="popper">
              {authorizedApplicationsOptions.map((option) => (
                <Select.Item key={option.value} value={option.value}>
                  {option.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </FormField>
        {selectedClient != null ? (
          <>
            <hr className={styles.divider} />
            <section className={styles.section}>
              <Text as="p" size="3" weight="medium" className={styles.title}>
                <FormattedMessage id="APIResourceDetailsScreen.test.accessToken.title" />
              </Text>
              <Text as="p" size="2" color="gray" className={styles.description}>
                <FormattedMessage
                  id="APIResourceDetailsScreen.test.accessToken.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    ExternalLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://www.authgear.com/tools/jwt-jwe-debugger/?utm_source=portal&utm_medium=link&utm_campaign=jwt_jwe_debugger">
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
              </Text>
              <CodeField
                className={styles.codeField}
                codeClassName="h-25 overflow-y-auto"
                placeholder={
                  <FormattedMessage id="APIResourceDetailsScreen.test.accessToken.placeholder" />
                }
              >
                {accessToken}
              </CodeField>
              <div className={styles.actions}>
                <PrimaryButton
                  size="2"
                  onClick={onGenerate}
                  disabled={isGenerating}
                  loading={isRevealing || isGenerating}
                  text={
                    <FormattedMessage id="APIResourceDetailsScreen.test.generateButton.text" />
                  }
                />
                <SecondaryButton
                  size="2"
                  onClick={onCopyToken}
                  disabled={accessToken == null}
                  text={
                    tokenCopied ? (
                      <FormattedMessage id="copied-to-clipboard" />
                    ) : (
                      <FormattedMessage id="copy" />
                    )
                  }
                />
              </div>
            </section>
            <hr className={styles.divider} />
            <section className={styles.section}>
              <Text as="p" size="3" weight="medium" className={styles.title}>
                <FormattedMessage
                  id="APIResourceDetailsScreen.test.requestToken.title"
                  values={{
                    clientName:
                      selectedClient.client_name ??
                      selectedClient.name ??
                      selectedClient.client_id,
                  }}
                />
              </Text>
              <Tabs.Root
                value={selectedCodeVariant}
                onValueChange={(value) =>
                  setSelectedCodeVariant(value as ExampleCodeVariant)
                }
              >
                <Tabs.List className={styles.codeTabsList}>
                  <Tabs.Trigger value={ExampleCodeVariant.curl}>
                    <FormattedMessage id="APIResourceDetailsScreen.test.pivot.curl.headerText" />
                  </Tabs.Trigger>
                  <Tabs.Trigger value={ExampleCodeVariant.Python}>
                    <FormattedMessage id="APIResourceDetailsScreen.test.pivot.python.headerText" />
                  </Tabs.Trigger>
                  <Tabs.Trigger value={ExampleCodeVariant.Go}>
                    <FormattedMessage id="APIResourceDetailsScreen.test.pivot.go.headerText" />
                  </Tabs.Trigger>
                  <Tabs.Trigger value={ExampleCodeVariant.NodeJS}>
                    <FormattedMessage id="APIResourceDetailsScreen.test.pivot.nodejs.headerText" />
                  </Tabs.Trigger>
                </Tabs.List>
              </Tabs.Root>
              <CodeField className={styles.codeField}>{exampleCode}</CodeField>
              <div className={styles.actions}>
                <PrimaryButton
                  size="2"
                  onClick={revealSecrets}
                  disabled={selectedClientSecret != null}
                  loading={isRevealing}
                  text={<FormattedMessage id="reveal" />}
                />
                <SecondaryButton
                  size="2"
                  onClick={onCopyCode}
                  text={
                    codeCopied ? (
                      <FormattedMessage id="copied-to-clipboard" />
                    ) : (
                      <FormattedMessage id="copy" />
                    )
                  }
                />
              </div>
            </section>
          </>
        ) : null}
      </SettingsSectionCard>
    </div>
  );
}
