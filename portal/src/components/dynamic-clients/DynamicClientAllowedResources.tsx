import React, { useMemo } from "react";
import { Spinner, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { useDynamicAccessResourcesQueryQuery } from "../../graphql/adminapi/query/dynamicAccessResourcesQuery.generated";
import styles from "./DynamicClientAllowedResources.module.css";

// The portal fetches at most this many resources / scopes per resource for
// this informational view. Projects rarely exceed it; the authoritative
// list lives on the API Resources screen.
const RESOURCES_FIRST = 100;
const SCOPES_FIRST = 100;

interface AllowedResource {
  id: string;
  name: string | null;
  resourceURI: string;
  scopes: string[];
}

export const DynamicClientAllowedResources: React.VFC =
  function DynamicClientAllowedResources() {
    const { data, loading } = useDynamicAccessResourcesQueryQuery({
      variables: { first: RESOURCES_FIRST, scopesFirst: SCOPES_FIRST },
    });

    const allowedResources = useMemo((): AllowedResource[] => {
      return (
        data?.resources?.edges
          ?.map((edge) => edge?.node)
          .filter((node): node is NonNullable<typeof node> => !!node)
          .filter(
            (node) => node.accessPolicy.allowDynamicThirdPartyClientAccess
          )
          .map((node) => ({
            id: node.id,
            name: node.name ?? null,
            resourceURI: node.resourceURI,
            scopes:
              node.scopes?.edges
                ?.map((edge) => edge?.node)
                .filter((scope): scope is NonNullable<typeof scope> => !!scope)
                .filter(
                  (scope) =>
                    scope.accessPolicy.allowDynamicThirdPartyClientAccess
                )
                .map((scope) => scope.scope) ?? [],
          })) ?? []
      );
    }, [data]);

    if (loading) {
      return <Spinner />;
    }

    if (allowedResources.length === 0) {
      return (
        <Text as="p" size="2" color="gray">
          <FormattedMessage id="DynamicClientAllowedResources.empty" />
        </Text>
      );
    }

    return (
      <div className={styles.resources}>
        {allowedResources.map((resource) => (
          <div key={resource.id} className={styles.resource}>
            <Text as="p" size="2">
              {resource.name != null && resource.name !== ""
                ? resource.name
                : resource.resourceURI}
            </Text>
            {resource.name != null && resource.name !== "" ? (
              <Text as="p" size="1" color="gray" className="break-all">
                {resource.resourceURI}
              </Text>
            ) : null}
            {resource.scopes.length > 0 ? (
              <div className={styles.scopes}>
                {resource.scopes.map((scope) => (
                  <span key={scope} className={styles.scopeChip}>
                    <Text size="1">{scope}</Text>
                  </span>
                ))}
              </div>
            ) : (
              <Text as="p" size="1" color="gray">
                <FormattedMessage id="DynamicClientAllowedResources.no-scopes" />
              </Text>
            )}
          </div>
        ))}
      </div>
    );
  };
