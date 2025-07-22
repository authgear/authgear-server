import React from "react";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { ResourceListEmptyView } from "./ResourceListEmptyView";

interface ResourceListProps {
  resources: Resource[];
  loading: boolean;
}

export const ResourceList: React.VFC<ResourceListProps> = function ResourceList(
  props
) {
  const { resources, loading } = props;

  if (loading) {
    return <div>Loading...</div>;
  }

  if (resources.length === 0) {
    return <ResourceListEmptyView />;
  }

  return (
    <>
      {resources.map((resource) => (
        <div key={resource!.id}>
          <h3>{resource!.name || resource!.resourceURI}</h3>
          <p>URI: {resource!.resourceURI}</p>
          <p>Created: {resource!.createdAt}</p>
          <p>Updated: {resource!.updatedAt}</p>
        </div>
      ))}
    </>
  );
};
