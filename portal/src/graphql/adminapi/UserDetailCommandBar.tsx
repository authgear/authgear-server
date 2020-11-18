import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  CommandBar,
  CommandButton,
  ICommandBarItemProps,
} from "@fluentui/react";

import { Context } from "@oursky/react-messageformat";

import styles from "./UserDetailCommandBar.module.scss";
import SetUserDisabledDialog from "./SetUserDisabledDialog";
import { extractUserInfoFromIdentities, Identity } from "../../util/user";

interface CommandBarUser {
  id: string;
  isDisabled: boolean;
}

interface UserDetailCommandBarProps {
  className?: string;
  user: CommandBarUser | null;
  identities: Identity[];
}

const UserDetailCommandBar: React.FC<UserDetailCommandBarProps> = function UserDetailCommandBar(
  props: UserDetailCommandBarProps
) {
  const { className, user, identities } = props;
  const { renderToString } = useContext(Context);

  interface DisableUserDialogData {
    isDisablingUser: boolean;
    userID: string;
    username: string | null;
  }
  const [
    disableUserDialogData,
    setDisableUserDialogData,
  ] = useState<DisableUserDialogData | null>(null);
  const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] = useState(
    true
  );
  const dismissDisableUserDialog = useCallback(() => {
    setIsDisableUserDialogHidden(true);
  }, []);

  const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
    if (!user) {
      return [];
    }
    const { id, isDisabled } = user;
    const { username, email, phone } = extractUserInfoFromIdentities(
      identities
    );
    return [
      /* TODO: to be implemented
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
      {
        key: "loginAsUser",
        text: renderToString("UserDetails.command-bar.login-as-user"),
        iconProps: { iconName: "FollowUser" },
      },
      */
      {
        key: "setDisabledStatus",
        text: user.isDisabled
          ? renderToString("enable")
          : renderToString("disable"),
        iconProps: { iconName: user.isDisabled ? "Play" : "CircleStop" },
        onRender: (props) => {
          return (
            <CommandButton
              {...props}
              onClick={() => {
                setDisableUserDialogData({
                  isDisablingUser: !isDisabled,
                  userID: id,
                  username: username ?? email ?? phone,
                });
                setIsDisableUserDialogHidden(false);
              }}
            />
          );
        },
      },
    ];
  }, [user, identities, renderToString]);

  return (
    <>
      <CommandBar
        className={cn(styles.commandBar, className)}
        items={[]}
        farItems={commandBarItems}
      />
      {disableUserDialogData != null && (
        <SetUserDisabledDialog
          isHidden={isDisableUserDialogHidden}
          onDismiss={dismissDisableUserDialog}
          {...disableUserDialogData}
        />
      )}
    </>
  );
};

export default UserDetailCommandBar;
