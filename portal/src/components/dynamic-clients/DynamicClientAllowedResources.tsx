import React, { useMemo } from "react";
import { Spinner, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { Callout } from "../v2/Callout/Callout";
import { useDynamicAccessResourcesQueryQuery } from "../../graphql/adminapi/query/dynamicAccessResourcesQuery.generated";
import styles from "./DynamicClientAllowedResources.module.css";

// The portal fetches at most this many resources / scopes per resource for
// this informational view. Projects rarely exceed it; the authoritative
// list lives on the API Resources screen. When a project does exceed it the
// UI says so rather than presenting a partial list as complete.
const RESOURCES_FIRST = 100;
const SCOPES_FIRST = 100;

interface AllowedResource {
  id: string;
  name: string | null;
  resourceURI: string;
  scopes: string[];
  // Total scopes on the resource, vs. the SCOPES_FIRST we actually inspected.
  scopesTotalCount: number | null;
  scopesFetchedCount: number;
}

export const DynamicClientAllowedResources: React.VFC =
  function DynamicClientAllowedResources() {
    const { data, loading } = useDynamicAccessResourcesQueryQuery({
      variables: { first: RESOURCES_FIRST, scopesFirst: SCOPES_FIRST },
    });

    const resourcesTotalCount = data?.resources?.totalCount ?? null;
    const resourcesFetchedCount = data?.resources?.edges?.length ?? 0;

    const allowedResources = useMemo((): AllowedResource[] => {
      return (
        data?.resources?.edges
          ?.map((edge) => edge?.node)
          .filter((node): node is NonNullable<typeof node> => !!node)
          .filter(
            (node) => node.accessPolicy.allowDynamicThirdPartyClientAccess
          )
          .map((node) => {
            const scopeNodes =
              node.scopes?.edges
                ?.map((edge) => edge?.node)
                .filter(
                  (scope): scope is NonNullable<typeof scope> => !!scope
                ) ?? [];
            return {
              id: node.id,
              name: node.name ?? null,
              resourceURI: node.resourceURI,
              scopes: scopeNodes
                .filter(
                  (scope) =>
                    scope.accessPolicy.allowDynamicThirdPartyClientAccess
                )
                .map((scope) => scope.scope),
              scopesTotalCount: node.scopes?.totalCount ?? null,
              scopesFetchedCount: scopeNodes.length,
            };
          }) ?? []
      );
    }, [data]);

    if (loading) {
      return <Spinner />;
    }

    // Only the first RESOURCES_FIRST resources were inspected, so a resource
    // that allows dynamic access may be missing from the list below.
    const resourcesTruncated =
      resourcesTotalCount != null &&
      resourcesTotalCount > resourcesFetchedCount;

    const truncationNotice = resourcesTruncated ? (
      <Callout
        type="warning"
        text={
          <FormattedMessage
            id="DynamicClientAllowedResources.resources-truncated"
            values={{
              shown: resourcesFetchedCount,
              total: resourcesTotalCount,
            }}
          />
        }
      />
    ) : null;

    if (allowedResources.length === 0) {
      return (
        <>
          {truncationNotice}
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="DynamicClientAllowedResources.empty" />
          </Text>
        </>
      );
    }

    return (
      <>
        {truncationNotice}
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
              {resource.scopesTotalCount != null &&
              resource.scopesTotalCount > resource.scopesFetchedCount ? (
                <Text as="p" size="1" color="gray">
                  <FormattedMessage
                    id="DynamicClientAllowedResources.scopes-truncated"
                    values={{
                      shown: resource.scopesFetchedCount,
                      total: resource.scopesTotalCount,
                    }}
                  />
                </Text>
              ) : null}
            </div>
          ))}
        </div>
      </>
    );
  };
