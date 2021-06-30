import React, { useEffect } from "react";
import ShowLoading from "../../ShowLoading";

const SubscriptionScreen: React.FC = function SubscriptionScreen() {
  useEffect(() => {
    // TODO(portal): implement subscription page
    window.location.href = "https://oursky.typeform.com/to/PecQiGfc";
  }, []);

  return <ShowLoading />;
};

export default SubscriptionScreen;
