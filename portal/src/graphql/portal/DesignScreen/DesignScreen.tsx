import React from "react";
import { DefaultEffects } from "@fluentui/react";
import cn from "classnames";

import { useParams } from "react-router-dom";
import FormContainer from "../../../FormContainer";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { useBrandDesignForm } from "./form";

const ConfigurationPanel: React.VFC = function ConfigurationPanel() {
  return <div></div>;
};

const DesignScreen: React.VFC = function DesignScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useBrandDesignForm(appID);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer className={cn("h-full")} form={form} canSave={true}>
      <div className={cn("h-full", "flex")}>
        <div className={cn("flex-1", "h-full", "p-6")}>
          <div
            className={cn("flex-1", "rounded-xl", "h-full")}
            style={{
              boxShadow: DefaultEffects.elevation4,
            }}
          >
            Preview
          </div>
        </div>
        <div className={cn("w-80", "p-6", "overflow-auto")}>
          <ConfigurationPanel />
        </div>
      </div>
    </FormContainer>
  );
};

export default DesignScreen;
