import React, { useMemo, useCallback, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { Nav, INavLink, INavLinkGroup, INavProps } from "@fluentui/react";

const ScreenNav: React.FC = function ScreenNav() {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const labelUsers = renderToString("nav.users");

  const navGroups: INavLinkGroup[] = useMemo(() => {
    return [
      {
        links: [
          {
            key: "users",
            name: labelUsers,
            url: "users",
            icon: "People",
          },
        ],
      },
    ];
  }, [labelUsers]);

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
