import React, { useContext, useMemo } from "react";
import cn from "classnames";
import {
  CommandBar,
  CommandButton,
  ICommandBarItemProps,
} from "@fluentui/react";
import TodoButtonWrapper from "../../TodoButtonWrapper";

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
      /* TODO: to be implemented
      {
        key: "remove",
        text: renderToString("remove"),
        iconProps: { iconName: "Delete" },
      },
      */
      {
        key: "loginAsUser",
        text: renderToString("UserDetails.command-bar.login-as-user"),
        iconProps: { iconName: "FollowUser" },
        onRender: (props) => {
          return (
            <TodoButtonWrapper>
              <CommandButton {...props} />
            </TodoButtonWrapper>
          );
        },
      },
      /* TODO: to be implemented
      {
        key: "disable",
        text: renderToString("disable"),
        iconProps: { iconName: "CircleStop" },
      },
       */
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
