import React from "react";
import ScreenContent from "../../ScreenContent";
import ScreenContentHeader from "../../ScreenContentHeader";
import { FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";

const CreateAPIResourceScreen: React.VFC = function CreateAPIResourceScreen() {
  return (
    <ScreenContent className="flex-1" layout="list">
      <ScreenContentHeader
        title={
          <NavBreadcrumb
            items={[
              {
                to: "~/api-resources",
                label: <FormattedMessage id="APIResourcesScreen.title" />,
              },
              {
                to: "",
                label: <FormattedMessage id="CreateAPIResourceScreen.title" />,
              },
            ]}
          />
        }
      />
    </ScreenContent>
  );
};

export default CreateAPIResourceScreen;
