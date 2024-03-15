import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";

import { RolesAndGroupsEmptyView } from "./RolesAndGroupsEmptyView";

import iconGroups from "../../../images/material-groups.svg";
import { ReactRouterLinkComponent } from "../../../ReactRouterLink";

export const GroupsEmptyView: React.VFC<{ className?: string }> =
  function GroupsEmptyView({ className }) {
    const { appID } = useParams<{ appID: string }>();

    return (
      <RolesAndGroupsEmptyView
        className={className}
        icon={<img src={iconGroups} />}
        title={<FormattedMessage id="GroupsEmptyView.title" />}
        description={<FormattedMessage id="GroupsEmptyView.description" />}
        button={
          <ReactRouterLinkComponent
            component={RolesAndGroupsEmptyView.CreateButton}
            to={`/project/${appID}/user-management/groups/add-group`}
            text={<FormattedMessage id="GroupsEmptyView.button.text" />}
          />
        }
      />
    );
  };
