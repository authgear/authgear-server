import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Select, Tabs, Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import {
  OAuthClientConfig,
  OAuthClientSecret,
  PortalAPIAppConfig,
} from "../../types";
import { useResourcesQueryQuery } from "../adminapi/query/resourcesQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import {
  ExampleCodeVariant,
  useExampleCode,
} from "../../components/api-resources/useExampleCode";
import { useEndpoints } from "../../hook/useEndpoints";
import { CodeField } from "../../components/common/CodeField";
import { copyToClipboard } from "../../util/clipboard";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { FormField } from "../../components/v2/FormField/FormField";
import { useNavigate } from "react-router-dom";
import PortalLink from "../../Link";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import { Resource } from "../adminapi/globalTypes.generated";
import { useSearchParamsState } from "../../hook/useSearchParamsState";
import { LocationState } from "./EditOAuthClientScreen";
import ExternalLink from "../../ExternalLink";
import { QuickStartStep } from "./QuickStartStep";

interface EditOAuthClientFormQuickStartContentProps {
  className?: string;
  appConfig: PortalAPIAppConfig;
  client: OAuthClientConfig;
  clientSecrets?: OAuthClientSecret | null;
}

export const EditOAuthClientFormQuickStartContent: React.VFC<EditOAuthClientFormQuickStartContentProps> =
  function EditOAuthClientFormQuickStartContent(props) {
    const { client } = props;

    const { data, loading, error, refetch } = useResourcesQueryQuery({
      variables: {
        first: 20,
        clientID: client.client_id,
      },
      fetchPolicy: "cache-and-network",
    });

    const resources = useMemo(() => {
      const resources =
        data?.resources?.edges
          ?.map((edge) => edge?.node)
          .filter((node) => !!node) ?? [];
      return resources;
    }, [data?.resources?.edges]);

    return useLoadableView({
      loadables: [
        {
          isLoading: loading,
          loadError: error,
          reload: refetch,
        },
      ],
      render: () => (
        <EditOAuthClientFormQuickStartContentLoaded
          {...props}
          resources={resources}
        />
      ),
    });
  };

interface EditOAuthClientFormQuickStartContentLoadedProps
  extends EditOAuthClientFormQuickStartContentProps {
  resources: Pick<Resource, "id" | "resourceURI" | "name">[];
}

function EditOAuthClientFormQuickStartContentLoaded(
  props: EditOAuthClientFormQuickStartContentLoadedProps
) {
  const { className, resources, appConfig, client, clientSecrets } = props;
  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();
  const { startReauthentication, isRevealing } =
    useStartReauthentication<LocationState>();
  const firstClientSecret =
    (clientSecrets?.keys?.length ?? 0) > 0 ? clientSecrets!.keys![0] : null;

  const isEmpty = resources.length === 0;

  const [selectedResourceURI, setSelectedResourceURI] =
    useSearchParamsState<string>(
      "resource",
      resources.length > 0 ? resources[0].resourceURI : ""
    );
  const [selectedCodeVariant, setSelectedCodeVariant] =
    useState<ExampleCodeVariant>(ExampleCodeVariant.curl);

  const handleCodeVariantChange = useCallback((value: string) => {
    setSelectedCodeVariant(value as ExampleCodeVariant);
  }, []);

  const { token: tokenEndpoint } = useEndpoints(
    appConfig.http?.public_origin ?? "",
    client.x_application_type
  );

  const exampleCode = useExampleCode({
    variant: selectedCodeVariant,
    tokenEndpoint,
    resourceURI: selectedResourceURI,
    clientSecret: firstClientSecret?.key ? firstClientSecret.key : null,
    clientID: client.client_id,
  });

  const [codeCopied, setCodeCopied] = useState(false);
  const copiedTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    return () => {
      if (copiedTimerRef.current != null) {
        clearTimeout(copiedTimerRef.current);
      }
    };
  }, []);
  const onCopyCode = useCallback(() => {
    copyToClipboard(exampleCode);
    setCodeCopied(true);
    if (copiedTimerRef.current != null) {
      clearTimeout(copiedTimerRef.current);
    }
    copiedTimerRef.current = setTimeout(() => {
      setCodeCopied(false);
    }, 2000);
  }, [exampleCode]);

  const revealSecrets = useCallback(() => {
    startReauthentication(navigate, {
      isClientSecretRevealed: true,
    }).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate, startReauthentication]);

  const resourceOptions = useMemo(() => {
    return resources.map((resource) => {
      return {
        value: resource.resourceURI,
        label: resource.name ?? resource.resourceURI,
      };
    });
  }, [resources]);

  return (
    <div className={className}>
      <Text as="p" size="4" weight="bold">
        <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.title" />
      </Text>
      <Text as="p" size="2" className="mt-2">
        <FormattedMessage
          id="EditOAuthClientForm.quick-start.m2m.description"
          values={{
            // eslint-disable-next-line react/no-unstable-nested-components
            docLink: (chunks: React.ReactNode) => (
              <ExternalLink href="https://docs.authgear.com/get-started/m2m-applications">
                {chunks}
              </ExternalLink>
            ),
          }}
        />
      </Text>
      <QuickStartStep
        className="mt-6"
        stepNumber="1"
        title={
          <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.step1.title" />
        }
      >
        <FormField
          size="2"
          label={renderToString(
            "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource"
          )}
        >
          <Select.Root
            value={selectedResourceURI === "" ? undefined : selectedResourceURI}
            onValueChange={setSelectedResourceURI}
            disabled={isEmpty}
            size="2"
          >
            <Select.Trigger
              placeholder={renderToString(
                isEmpty
                  ? "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource.empty.placeholder"
                  : "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource.placeholder"
              )}
            />
            <Select.Content position="popper">
              {resourceOptions.map((option) => (
                <Select.Item key={option.value} value={option.value}>
                  {option.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </FormField>
        {isEmpty ? (
          <Text as="p" size="2" className="mt-2">
            <FormattedMessage
              id="EditOAuthClientForm.quick-start.m2m.step1.no-api-resource-yet"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                reactRouterLink: (chunks: React.ReactNode) => (
                  <PortalLink to="?tab=api-resources">{chunks}</PortalLink>
                ),
              }}
            />
          </Text>
        ) : null}
        {selectedResourceURI ? (
          <div>
            <Tabs.Root
              className="mt-2"
              value={selectedCodeVariant}
              onValueChange={handleCodeVariantChange}
            >
              <Tabs.List>
                <Tabs.Trigger value={ExampleCodeVariant.curl}>
                  <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.pivot.curl.headerText" />
                </Tabs.Trigger>
                <Tabs.Trigger value={ExampleCodeVariant.Python}>
                  <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.pivot.python.headerText" />
                </Tabs.Trigger>
                <Tabs.Trigger value={ExampleCodeVariant.Go}>
                  <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.pivot.go.headerText" />
                </Tabs.Trigger>
                <Tabs.Trigger value={ExampleCodeVariant.NodeJS}>
                  <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.pivot.nodejs.headerText" />
                </Tabs.Trigger>
              </Tabs.List>
            </Tabs.Root>
            <CodeField className="mt-1">{exampleCode}</CodeField>
            <div className="mt-4 flex space-x-4">
              <PrimaryButton
                size="2"
                onClick={revealSecrets}
                disabled={!!firstClientSecret?.key}
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
          </div>
        ) : null}
      </QuickStartStep>
      <QuickStartStep
        className="mt-6"
        stepNumber="2"
        title={
          <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.step2.title" />
        }
      >
        <Text as="p" size="2">
          <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.step2.description" />
        </Text>
        <CodeField className="mt-1">{`Authorization: Bearer <token>`}</CodeField>
      </QuickStartStep>
    </div>
  );
}
