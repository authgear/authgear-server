import React from "react";
import { FormattedMessage } from "../../../intl";
import { useParams } from "react-router-dom";

import { RolesAndGroupsEmptyView } from "./RolesAndGroupsEmptyView";

import iconBadge from "../../../images/badge.svg";
import { ReactRouterLinkComponent } from "../../../ReactRouterLink";

export const RolesEmptyView: React.VFC<{ className?: string }> =
  function RolesEmptyView({ className }) {
    const { appID } = useParams<{ appID: string }>();

    return (
      <RolesAndGroupsEmptyView
        className={className}
        icon={<img src={iconBadge} />}
        title={<FormattedMessage id="RolesEmptyView.title" />}
        description={<FormattedMessage id="RolesEmptyView.description" />}
        button={
          <ReactRouterLinkComponent
            component={RolesAndGroupsEmptyView.CreateButton}
            to={`/project/${appID}/user-management/roles/add-role`}
            text={<FormattedMessage id="RolesEmptyView.button.text" />}
          />
        }
      />
    );
  };
