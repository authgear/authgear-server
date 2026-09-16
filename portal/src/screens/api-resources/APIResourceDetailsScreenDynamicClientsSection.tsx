import React from "react";
import { FormattedMessage } from "../../intl";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { DynamicClientsAccessRow } from "../../components/api-resources/DynamicClientsAccessRow";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";

export function APIResourceDetailsScreenDynamicClientsSection({
  resource,
}: {
  resource: Resource;
}): JSX.Element {
  return (
    <SettingsSectionCard
      title={
        <FormattedMessage id="APIResourceDetailsScreen.section.dynamic-clients" />
      }
      description={
        <FormattedMessage id="APIResourceDetailsScreen.dynamic-clients.description" />
      }
    >
      <DynamicClientsAccessRow resource={resource} />
    </SettingsSectionCard>
  );
}
