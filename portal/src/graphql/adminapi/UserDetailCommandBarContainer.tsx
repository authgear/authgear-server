import React, { useCallback, useContext, useMemo, useState } from "react";
import { CommandButton, ICommandBarItemProps } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import SetUserDisabledDialog from "./SetUserDisabledDialog";
import { extractUserInfoFromIdentities } from "../../util/user";
import { Identity } from "../../types";
import DeleteUserDialog from "./DeleteUserDialog";
import { useNavigate } from "react-router-dom";
import CommandBarContainer from "../../CommandBarContainer";

interface CommandBarUser {
  id: string;
  isDisabled: boolean;
}

interface UserDetailCommandBarContainerProps {
  className?: string;
  user: CommandBarUser | null;
  identities: Identity[];
}

const UserDetailCommandBarContainer: React.FC<UserDetailCommandBarContainerProps> =
  function UserDetailCommandBarContainer(props) {
    const { className, user, identities } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
    const navigate = useNavigate();

    interface DisableUserDialogData {
      isDisablingUser: boolean;
      userID: string;
      username: string | null;
    }
    const [disableUserDialogData, setDisableUserDialogData] =
      useState<DisableUserDialogData | null>(null);
    const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] =
      useState(true);
    const dismissDisableUserDialog = useCallback(() => {
      setIsDisableUserDialogHidden(true);
    }, []);

    interface DeleteUserDialogData {
      userID: string;
      username: string | null;
    }
    const [deleteUserDialogData, setDeleteUserDialogData] =
      useState<DeleteUserDialogData | null>(null);
    const [isDeleteUserDialogHidden, setIsDeleteUserDialogHidden] =
      useState(true);
    const dismissDeleteUserDialog = useCallback(
      (deletedUser: boolean) => {
        setIsDeleteUserDialogHidden(true);
        if (deletedUser) {
          setTimeout(() => navigate("../.."), 0);
        }
      },
      [navigate]
    );

    const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
      if (!user) {
        return [];
      }
      const { id, isDisabled } = user;
      const { username, email, phone } =
        extractUserInfoFromIdentities(identities);
      return [
        {
          key: "remove",
          text: renderToString("remove"),
          iconProps: { iconName: "Delete" },
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (props) => {
            return (
              <CommandButton
                {...props}
                theme={themes.destructive}
                onClick={() => {
                  setDeleteUserDialogData({
                    userID: id,
                    username: username ?? email ?? phone,
                  });
                  setIsDeleteUserDialogHidden(false);
                }}
              />
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
          key: "setDisabledStatus",
          text: user.isDisabled
            ? renderToString("enable")
            : renderToString("disable"),
          iconProps: { iconName: user.isDisabled ? "Play" : "CircleStop" },
          // eslint-disable-next-line react/no-unstable-nested-components
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
    }, [user, identities, renderToString, themes.destructive]);

    return (
      <CommandBarContainer className={className} farItems={commandBarItems}>
        {props.children}
        {disableUserDialogData != null && (
          <SetUserDisabledDialog
            isHidden={isDisableUserDialogHidden}
            onDismiss={dismissDisableUserDialog}
            {...disableUserDialogData}
          />
        )}
        {deleteUserDialogData != null && (
          <DeleteUserDialog
            isHidden={isDeleteUserDialogHidden}
            onDismiss={dismissDeleteUserDialog}
            {...deleteUserDialogData}
          />
        )}
      </CommandBarContainer>
    );
  };

export default UserDetailCommandBarContainer;
