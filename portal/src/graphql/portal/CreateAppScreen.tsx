import React, { useCallback, useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Context as LocaleContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import { PrimaryButton, Text, TextField } from "@fluentui/react";
import ShowError from "../../ShowError";
import ScreenHeader from "../../ScreenHeader";
import styles from "./CreateAppScreen.module.scss";
import { useCreateAppMutation } from "./mutations/createAppMutation";
import { useTextField } from "../../hook/useInput";

interface CreateAppProps {
  isCreating: boolean;
  createApp: (appID: string) => Promise<string | null>;
}

interface CreateAppFormData {
  appID: string;
}

const CreateApp: React.FC<CreateAppProps> = function CreateApp(
  props: CreateAppProps
) {
  const { isCreating, createApp } = props;
  const navigate = useNavigate();
  const { renderToString } = useContext(LocaleContext);

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
      <Text as="h1" variant="xLarge" block={true}>
        <FormattedMessage id="CreateAppScreen.title" />
      </Text>
      <TextField
        className={styles.appIDField}
        label={renderToString("CreateAppScreen.app-id.label")}
        value={appID}
        disabled={isCreating}
        onChange={onAppIDChange}
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
  return (
    <div className={styles.root}>
      <ScreenHeader />
      {error && <ShowError error={error} />}
      <CreateApp isCreating={loading} createApp={createApp} />
    </div>
  );
};

export default CreateAppScreen;
