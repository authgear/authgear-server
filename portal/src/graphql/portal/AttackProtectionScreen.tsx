import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenTitle from "../../ScreenTitle";

const AttackProtectionScreen: React.FC = function AttackProtectionScreen() {
  return (
    <div>
      <ScreenTitle>
        <FormattedMessage id="AttackProtectionScreen.title" />
      </ScreenTitle>
    </div>
  );
};

export default AttackProtectionScreen;
