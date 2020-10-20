import React, { useCallback, useContext, useMemo, useState } from "react";
import { TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import deepEqual from "deep-equal";

import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";

import styles from "./GeneralSettings.module.scss";

interface State {
  appID: string;
}

const GeneralSettings: React.FC = function GeneralSettings() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();

  const initialState = useMemo<State>(() => {
    return {
      appID: appID,
    };
  }, [appID]);

  const [state] = useState<State>(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onFormSubmit = useCallback((ev: React.SyntheticEvent<HTMLElement>) => {
    ev.preventDefault();
    ev.stopPropagation();

    alert("App ID cannot be saved currently");
  }, []);

  return (
    <form className={styles.root} onSubmit={onFormSubmit}>
      <TextField
        className={styles.textField}
        disabled={true}
        label={renderToString("GeneralSettings.app-id.label")}
        value={state.appID}
      />
      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified}
          loading={false}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </form>
  );
};

export default GeneralSettings;
