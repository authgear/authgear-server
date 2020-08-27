import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";

const ShowLoading: React.FC = function ShowLoading() {
  return (
    <p>
      <FormattedMessage id="loading" />
    </p>
  );
};

export default ShowLoading;
