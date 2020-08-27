import React, { useMemo } from "react";
import { Routes, Route, useParams } from "react-router-dom";
import { makeEnvironment } from "./graphql/adminapi/relay";
import AppContext from "./AppContext";
import AppScreen from "./graphql/portal/AppScreen";

const AppRoot: React.FC = function AppRoot() {
  const { appID } = useParams();
  const environment = useMemo(() => {
    return makeEnvironment(appID);
  }, [appID]);
  return (
    <AppContext.Provider value={environment}>
      <Routes>
        <Route path="/" element={<AppScreen />} />
      </Routes>
    </AppContext.Provider>
  );
};

export default AppRoot;
