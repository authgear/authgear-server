import React, { useCallback, useContext, useMemo, useState } from "react";
import { Dropdown, IDropdownOption, PivotItem, Text } from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
import {
  Context as MessageContext,
  FormattedMessage,
} from "../../intl";
import WidgetTitle from "../../WidgetTitle";
import {
  OAuthClientConfig,
  OAuthClientSecret,
  PortalAPIAppConfig,
} from "../../types";
import { useResourcesQueryQuery } from "../adminapi/query/resourcesQuery.generated";
import styles from "./EditOAuthClientFormQuickStartContent.module.css";
import { useLoadableView } from "../../hook/useLoadableView";
import {
  ExampleCodeVariant,
  useExampleCode,
} from "../../components/api-resources/useExampleCode";
import { useEndpoints } from "../../hook/useEndpoints";
import { CodeField } from "../../components/common/CodeField";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import DefaultButton from "../../DefaultButton";
import { useNavigate } from "react-router-dom";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import ButtonWithLoading from "../../ButtonWithLoading";
import { Resource } from "../adminapi/globalTypes.generated";
import { useSearchParamsState } from "../../hook/useSearchParamsState";
import { LocationState } from "./EditOAuthClientScreen";

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

  const handleDropdownChange = useCallback(
    (_: unknown, option?: IDropdownOption) => {
      setSelectedResourceURI(String(option?.key ?? ""));
    },
    [setSelectedResourceURI]
  );

  const handlePivotClick = useCallback((item?: PivotItem) => {
    if (item?.props.itemKey) {
      setSelectedCodeVariant(item.props.itemKey as ExampleCodeVariant);
    }
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

  const { copyButtonProps, Feedback: CopyFeedback } = useCopyFeedback({
    textToCopy: exampleCode,
  });

  const revealSecrets = useCallback(() => {
    startReauthentication(navigate, {
      isClientSecretRevealed: true,
    }).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate, startReauthentication]);

  const resourceOptions = useMemo((): IDropdownOption[] => {
    return resources.map((resource) => {
      return {
        key: resource.resourceURI,
        text: resource.name ?? resource.resourceURI,
      };
    });
  }, [resources]);

  return (
    <div className={className}>
      <WidgetTitle>
        <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.title" />
      </WidgetTitle>
      <Text as="p" variant="medium" className="mt-2" block={true}>
        <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.description" />
      </Text>
      <QuickStartStep
        className="mt-6"
        stepNumber="1"
        title={
          <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.step1.title" />
        }
      >
        <Dropdown
          label={renderToString(
            "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource"
          )}
          options={resourceOptions}
          placeholder={renderToString(
            isEmpty
              ? "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource.empty.placeholder"
              : "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource.placeholder"
          )}
          selectedKey={selectedResourceURI}
          disabled={isEmpty}
          onChange={handleDropdownChange}
        />
        {isEmpty ? (
          <Text as="p" block={true} className="mt-2">
            <FormattedMessage
              id="EditOAuthClientForm.quick-start.m2m.step1.no-api-resource-yet"
              values={{ href: "?tab=api-resources" }}
            />
          </Text>
        ) : null}
        {selectedResourceURI ? (
          <div>
            <AGPivot
              className="mt-2"
              selectedKey={selectedCodeVariant}
              onLinkClick={handlePivotClick}
            >
              <PivotItem
                headerText={renderToString(
                  "EditOAuthClientForm.quick-start.m2m.pivot.curl.headerText"
                )}
                itemKey={ExampleCodeVariant.curl}
              />
              <PivotItem
                headerText={renderToString(
                  "EditOAuthClientForm.quick-start.m2m.pivot.python.headerText"
                )}
                itemKey={ExampleCodeVariant.Python}
              />
              <PivotItem
                headerText={renderToString(
                  "EditOAuthClientForm.quick-start.m2m.pivot.go.headerText"
                )}
                itemKey={ExampleCodeVariant.Go}
              />
              <PivotItem
                headerText={renderToString(
                  "EditOAuthClientForm.quick-start.m2m.pivot.nodejs.headerText"
                )}
                itemKey={ExampleCodeVariant.NodeJS}
              />
            </AGPivot>
            <CodeField className="mt-1">{exampleCode}</CodeField>
            <div className="mt-4 flex space-x-4">
              <ButtonWithLoading
                labelId="reveal"
                onClick={revealSecrets}
                disabled={!!firstClientSecret?.key}
                loading={isRevealing}
              />
              <DefaultButton
                {...copyButtonProps}
                text={<FormattedMessage id="copy" />}
                iconProps={undefined}
              />
              <CopyFeedback />
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
        <Text as="p" block={true}>
          <FormattedMessage id="EditOAuthClientForm.quick-start.m2m.step2.description" />
        </Text>
        <CodeField className="mt-1">{`Authorization: Bearer <token>`}</CodeField>
      </QuickStartStep>
    </div>
  );
}

function QuickStartStep({
  className,
  stepNumber,
  title,
  children,
}: {
  className?: string;
  stepNumber: string;
  title: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <section className={className}>
      <header className={styles.quickStartStep__header}>
        <Text
          variant="mediumPlus"
          styles={{
            root: {
              fontWeight: 600,
              color: "var(--gray-12)",
              backgroundColor: "var(--gray-a3)",
              width: 22,
              height: 22,
              borderRadius: 999,
              textAlign: "center",
              lineHeight: 20,
            },
          }}
          block={true}
        >
          {stepNumber}
        </Text>
        <Text
          variant="mediumPlus"
          styles={{
            root: {
              fontWeight: 600,
              color: "var(--gray-12)",
            },
          }}
        >
          {title}
        </Text>
      </header>
      <div className={styles.quickStartStep__childrenContainer}>{children}</div>
    </section>
  );
}
