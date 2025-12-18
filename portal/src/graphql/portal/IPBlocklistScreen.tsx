import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenTitle from "../../ScreenTitle";

const IPBlocklistScreen: React.FC = function IPBlocklistScreen() {
  return (
    <div>
      <ScreenTitle>
        <FormattedMessage id="IPBlocklistScreen.title" />
      </ScreenTitle>
    </div>
  );
};

export default IPBlocklistScreen;
