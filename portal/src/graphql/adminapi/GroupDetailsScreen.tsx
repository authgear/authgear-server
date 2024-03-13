import React from "react";
import { useParams } from "react-router-dom";
import { useGroupQuery } from "./query/groupQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { GroupQueryNodeFragment } from "./query/groupQuery.generated";

function GroupDetailsScreenLoaded(_props: {
  group: GroupQueryNodeFragment;
  reload: ReturnType<typeof useGroupQuery>["refetch"];
}) {
  return <></>;
}

const GroupDetailsScreen: React.VFC = function GroupDetailsScreen() {
  const { groupID } = useParams() as { groupID: string };
  const { group, loading, error, refetch } = useGroupQuery(groupID);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (group == null) {
    return <ShowLoading />;
  }

  return <GroupDetailsScreenLoaded group={group} reload={refetch} />;
};

export default GroupDetailsScreen;
