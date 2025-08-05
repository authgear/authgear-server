import React, { useContext, useMemo } from "react";
import { Dropdown, IDropdownOption, Text } from "@fluentui/react";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import WidgetTitle from "../../WidgetTitle";
import { OAuthClientConfig } from "../../types";
import { useResourcesQueryQuery } from "../adminapi/query/resourcesQuery.generated";
import styles from "./EditOAuthClientFormQuickStartContent.module.css";
import { useLoadableView } from "../../hook/useLoadableView";

interface EditOAuthClientFormQuickStartContentProps {
  className?: string;
  client: OAuthClientConfig;
}

export const EditOAuthClientFormQuickStartContent: React.VFC<EditOAuthClientFormQuickStartContentProps> =
  function EditOAuthClientFormQuickStartContent(props) {
    const { className, client } = props;

    const { data, loading, error, refetch } = useResourcesQueryQuery({
      variables: {
        first: 20,
        clientID: client.client_id,
      },
      fetchPolicy: "cache-and-network",
    });

    const resources = useMemo((): IDropdownOption[] => {
      const resources =
        data?.resources?.edges
          ?.map((edge) => edge?.node)
          .filter((node) => !!node) ?? [];
      return resources.map((resource) => {
        return {
          key: resource.id,
          text: resource.name ?? resource.resourceURI,
        };
      });
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
          className={className}
          client={client}
          resources={resources}
        />
      ),
    });
  };

interface EditOAuthClientFormQuickStartContentLoadedProps
  extends EditOAuthClientFormQuickStartContentProps {
  resources: IDropdownOption[];
}

function EditOAuthClientFormQuickStartContentLoaded(
  props: EditOAuthClientFormQuickStartContentLoadedProps
) {
  const { className, resources } = props;
  const { renderToString } = useContext(MessageContext);

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
          options={resources}
          placeholder={renderToString(
            "EditOAuthClientForm.quick-start.m2m.step1.select-api-resource.placeholder"
          )}
        />
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
