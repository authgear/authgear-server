import React, { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, PrimaryButton, Text, TextField } from "@fluentui/react";
import ShowError from "../../ShowError";
import ScreenHeader from "../../ScreenHeader";
import NavBreadcrumb from "../../NavBreadcrumb";
import { useCreateAppMutation } from "./mutations/createAppMutation";
import { useTextField } from "../../hook/useInput";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useGenericError } from "../../error/useGenericError";

import styles from "./CreateAppScreen.module.scss";

interface CreateAppProps {
  isCreating: boolean;
  createApp: (appID: string) => Promise<string | null>;
  errorMessage: string | undefined;
}

interface CreateAppFormData {
  appID: string;
}

const APP_ID_SCHEME = "https://";

const CreateApp: React.FC<CreateAppProps> = function CreateApp(
  props: CreateAppProps
) {
  const { isCreating, createApp, errorMessage } = props;
  const navigate = useNavigate();
  const systemConfig = useSystemConfig();

  const [formData, setFormData] = useState<CreateAppFormData>({
    appID: "",
  });
  const { appID } = formData;
  const { onChange: onAppIDChange } = useTextField((value) =>
    setFormData((prev) => ({ ...prev, appID: value }))
  );

  const onCreateClick = useCallback(() => {
    createApp(appID)
      .then((id) => {
        if (id) {
          navigate("/app/" + encodeURIComponent(id));
        }
      })
      .catch(() => {});
  }, [appID, createApp, navigate]);

  return (
    <main className={styles.body}>
      <Label className={styles.fieldLabel}>
        <FormattedMessage id="CreateAppScreen.app-id.label" />
      </Label>
      <Text className={styles.fieldDesc}>
        <FormattedMessage id="CreateAppScreen.app-id.desc" />
      </Text>
      <TextField
        className={styles.appIDField}
        value={appID}
        disabled={isCreating}
        onChange={onAppIDChange}
        prefix={APP_ID_SCHEME}
        suffix={systemConfig.appHostSuffix}
        errorMessage={errorMessage}
      />
      <PrimaryButton
        onClick={onCreateClick}
        disabled={appID.length === 0 || isCreating}
      >
        <FormattedMessage id="create" />
      </PrimaryButton>
    </main>
  );
};

const CreateAppScreen: React.FC = function CreateAppScreen() {
  const { loading, error, createApp } = useCreateAppMutation();

  const navBreadcrumbItems = React.useMemo(() => {
    return [
      { to: "..", label: <FormattedMessage id="AppsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="CreateAppScreen.title" /> },
    ];
  }, []);

  const { errorMessage, unrecognizedError } = useGenericError(
    error,
    [],
    [
      {
        reason: "DuplicatedAppID",
        errorMessageID: "CreateAppScreen.error.duplicated-app-id",
      },
    ],

    "CreateAppScreen.error.generic"
  );

  return (
    <div className={styles.root}>
      <ScreenHeader />
      {unrecognizedError && <ShowError error={unrecognizedError} />}
      <section className={styles.content}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <CreateApp
          isCreating={loading}
          createApp={createApp}
          errorMessage={errorMessage}
        />
      </section>
    </div>
  );
};

export default CreateAppScreen;
