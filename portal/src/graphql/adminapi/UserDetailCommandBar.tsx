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
      {
        key: "remove",
        text: renderToString("remove"),
        iconProps: { iconName: "Delete" },
        onRender: (props) => {
          return (
            <TodoButtonWrapper>
              <CommandButton disabled={true} {...props} />
            </TodoButtonWrapper>
          );
        },
      },
      /* TODO: to be implemented
      {
        key: "loginAsUser",
        text: renderToString("UserDetails.command-bar.login-as-user"),
        iconProps: { iconName: "FollowUser" },
      },
      */
      {
        key: "disable",
        text: renderToString("disable"),
        iconProps: { iconName: "CircleStop" },
        onRender: (props) => {
          return (
            <TodoButtonWrapper>
              <CommandButton disabled={true} {...props} />
            </TodoButtonWrapper>
          );
        },
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
