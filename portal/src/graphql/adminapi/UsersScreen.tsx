import React, { useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";

const UsersScreen: React.FC = function UsersScreen() {
  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="UsersScreen.title" /> }];
  }, []);
  return (
    <>
      <NavBreadcrumb items={items} />
      <UsersList />
    </>
  );
};

export default UsersScreen;
