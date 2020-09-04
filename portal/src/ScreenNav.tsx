import React, { useMemo, useCallback, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { Nav, INavLink, INavLinkGroup, INavProps } from "@fluentui/react";

const ScreenNav: React.FC = function ScreenNav() {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const label = renderToString("ScreenNav.label");

  const navGroups: INavLinkGroup[] = useMemo(() => {
    return [
      {
        links: [
          {
            key: "users",
            name: renderToString("ScreenNav.users"),
            url: "users",
            icon: "People",
          },
          {
            key: "authentication",
            name: renderToString("ScreenNav.authentication"),
            url: "configuration/authentication",
            icon: "Shield",
          },
        ],
      },
    ];
  }, [renderToString]);

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

  return <Nav ariaLabel={label} groups={navGroups} onLinkClick={onLinkClick} />;
};

export default ScreenNav;
