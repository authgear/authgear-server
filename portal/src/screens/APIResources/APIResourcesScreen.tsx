import React from "react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import ScreenContentHeader from "../../ScreenContentHeader";
import { FormattedMessage } from "@oursky/react-messageformat";
import { ResourceListEmptyView } from "../../components/api-resources/ResourceListEmptyView";

const APIResourcesScreen: React.VFC = function APIResourcesScreen() {
  return (
    <ScreenContent>
      <ScreenContentHeader
        title={
          <ScreenTitle>
            <FormattedMessage id="APIResourcesScreen.title" />
          </ScreenTitle>
        }
        description={
          <ScreenDescription>
            <FormattedMessage id="APIResourcesScreen.description" />
          </ScreenDescription>
        }
      />
      <div className="col-span-full p-8">
        <ResourceListEmptyView />
      </div>
    </ScreenContent>
  );
};

export default APIResourcesScreen;
