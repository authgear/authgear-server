import React from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";

import styles from "./UserDetailsScreen.module.scss";

const UserDetailsScreen: React.FC = function UserDetailsScreen() {
  const { userID } = useParams();

  const navBreadcrumItems = React.useMemo(() => {
    return [
      { to: "../../", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="UserDetailsScreen.title" /> },
    ];
  }, []);

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumItems} />
    </div>
  );
};

export default UserDetailsScreen;
