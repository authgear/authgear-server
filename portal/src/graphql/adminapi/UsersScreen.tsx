import React, { useContext, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { CommandBar, ICommandBarItemProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";

import styles from "./UsersScreen.module.scss";

const UsersScreen: React.FC = function UsersScreen() {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="UsersScreen.title" /> }];
  }, []);

  const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "addUser",
        text: renderToString("UsersScreen.add-user"),
        iconProps: { iconName: "CirclePlus" },
        onClick: () => navigate("./add-user"),
      },
    ];
  }, [navigate, renderToString]);

  return (
    <main className={styles.root}>
      <CommandBar
        className={styles.commandBar}
        items={[]}
        farItems={commandBarItems}
      />
      <NavBreadcrumb items={items} />
      <UsersList className={styles.usersList} />
    </main>
  );
};

export default UsersScreen;
