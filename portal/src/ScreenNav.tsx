import React, { useMemo, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Nav, INavLink, INavLinkGroup, INavProps } from "@fluentui/react";

const ScreenNav: React.FC = function ScreenNav() {
  const { appID } = useParams();
  const navigate = useNavigate();

  const navGroups: INavLinkGroup[] = useMemo(() => {
    const appRootURL = `/apps/${encodeURIComponent(appID)}`;
    return [
      {
        links: [
          {
            key: "dashboard",
            name: "Dashboard",
            url: appRootURL,
            icon: "Rocket",
          },
          {
            key: "dummy",
            name: "dummy",
            url: appRootURL + "/dummy",
            icon: "Diamond",
          },
        ],
      },
    ];
  }, [appID]);

  const onLinkClick: INavProps["onLinkClick"] = useCallback(
    (ev?: React.MouseEvent<HTMLElement>, item?: INavLink) => {
      ev?.stopPropagation();
      ev?.preventDefault();
      if (item != null) {
        navigate(item.url);
      }
    },
    [navigate]
  );

  return <Nav groups={navGroups} onLinkClick={onLinkClick} />;
};

export default ScreenNav;
