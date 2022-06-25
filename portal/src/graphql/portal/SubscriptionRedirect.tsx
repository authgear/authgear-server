import React, { useEffect } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useReconcileCheckoutSessionMutation } from "./mutations/reconcileCheckoutSessionMutation";

const SubscriptionRedirect: React.FC = function SubscriptionRedirect() {
  const navigate = useNavigate();
  const { appID } = useParams() as { appID: string };
  const [searchParams] = useSearchParams();
  const stripeCheckoutSessionID = searchParams.get("session_id");
  const { reconcileCheckoutSession, error } =
    useReconcileCheckoutSessionMutation();

  useEffect(() => {
    if (stripeCheckoutSessionID) {
      reconcileCheckoutSession(appID, stripeCheckoutSessionID)
        .then(() => {
          navigate("./../billing", { replace: false });
        })
        .catch(() => {})
        .finally(() => {});
    } else {
      navigate("./../billing", { replace: false });
    }
  }, [navigate, reconcileCheckoutSession, stripeCheckoutSessionID, appID]);

  if (error) {
    return <ShowError error={error} />;
  }

  return <ShowLoading />;
};

export default SubscriptionRedirect;
