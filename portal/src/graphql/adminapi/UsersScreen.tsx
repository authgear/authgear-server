import React, { useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";
import styles from "./UsersScreen.module.scss";

const UsersScreen: React.FC = function UsersScreen() {
  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="UsersScreen.title" /> }];
  }, []);
  return (
    <main className={styles.root}>
      <NavBreadcrumb items={items} />
      <UsersList className={styles.usersList} />
    </main>
  );
};

export default UsersScreen;
