import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { CommandBar, ICommandBarItemProps } from "@fluentui/react";

import { Context } from "@oursky/react-messageformat";

import styles from "./UserDetailCommandBar.module.scss";

interface UserDetailCommandBarProps {
  className?: string;
}

const UserDetailCommandBar: React.FC<UserDetailCommandBarProps> = function UserDetailCommandBar(
  props: UserDetailCommandBarProps
) {
  const { className } = props;
  const { renderToString } = useContext(Context);

  const commandBarItems: ICommandBarItemProps[] = useMemo(
    () => [
      {
        key: "remove",
        text: renderToString("remove"),
        iconProps: { iconName: "Delete" },
      },

      {
        key: "loginAsUser",
        text: renderToString("UserDetails.command-bar.login-as-user"),
        iconProps: { iconName: "FollowUser" },
      },

      {
        key: "invalidateSessions",
        text: renderToString("UserDetails.command-bar.invalidate-sessions"),
        iconProps: {
          iconName: "CircleAddition",
          className: styles.invalidateIcon,
        },
      },
      {
        key: "disable",
        text: renderToString("disable"),
        iconProps: { iconName: "CircleStop" },
      },
    ],
    [renderToString]
  );

  return (
    <CommandBar
      className={cn(styles.commandBar, className)}
      items={[]}
      farItems={commandBarItems}
    />
  );
};

export default UserDetailCommandBar;
