import React, { useMemo } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { makeEnvironment } from "./graphql/adminapi/relay";
import AppContext from "./AppContext";
import ScreenLayout from "./ScreenLayout";
import UsersScreen from "./graphql/adminapi/UsersScreen";

const AppRoot: React.FC = function AppRoot() {
  const { appID } = useParams();
  const environment = useMemo(() => {
    return makeEnvironment(appID);
  }, [appID]);
  return (
    <AppContext.Provider value={environment}>
      <ScreenLayout>
        <Routes>
          <Route path="/" element={<Navigate to="users" replace={true} />} />
          <Route path="/users" element={<UsersScreen />} />
        </Routes>
      </ScreenLayout>
    </AppContext.Provider>
  );
};

export default AppRoot;
