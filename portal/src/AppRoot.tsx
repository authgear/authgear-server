import React, { useMemo } from "react";
import { Routes, Route, useParams } from "react-router-dom";
import { makeEnvironment } from "./graphql/adminapi/relay";
import AppContext from "./AppContext";
import AppScreen from "./graphql/portal/AppScreen";
import ScreenLayout from "./ScreenLayout";

const DummyScreen: React.FC = function DummyScreen() {
  return <p>This is dummy</p>;
};

const AppRoot: React.FC = function AppRoot() {
  const { appID } = useParams();
  const environment = useMemo(() => {
    return makeEnvironment(appID);
  }, [appID]);
  return (
    <AppContext.Provider value={environment}>
      <ScreenLayout>
        <Routes>
          <Route path="/" element={<AppScreen />} />
          <Route path="/dummy" element={<DummyScreen />} />
        </Routes>
      </ScreenLayout>
    </AppContext.Provider>
  );
};

export default AppRoot;
