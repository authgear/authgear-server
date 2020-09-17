import React, { useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

import styles from "./AddUserScreen.module.scss";

const AddUserScreen: React.FC = function AddUserScreen() {
  const navBreadcrumbItems: BreadcrumbItem = useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
    </main>
  );
};

export default AddUserScreen;
