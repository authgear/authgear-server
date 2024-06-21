import React from "react";
import { useParams } from "react-router-dom";
import FormContainer from "../../../FormContainer";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { useBrandDesignForm } from "./form";

const DesignScreen: React.VFC = function DesignScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useBrandDesignForm(appID);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return <FormContainer form={form} canSave={true}></FormContainer>;
};

export default DesignScreen;
