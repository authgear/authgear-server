import React, { useCallback, useContext, useMemo, useState } from "react";
import { TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import deepEqual from "deep-equal";

import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";

import styles from "./GeneralSettings.module.scss";

interface State {
  appName: string;
}

const GeneralSettings: React.FC = function GeneralSettings() {
  const { renderToString } = useContext(Context);

  const initialState = useMemo<State>(() => {
    return {
      appName: "",
    };
  }, []);

  const [state, setState] = useState<State>(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onAppNameChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((prevState) => ({
      ...prevState,
      appName: value,
    }));
  }, []);

  const onFormSubmit = useCallback((ev: React.SyntheticEvent<HTMLElement>) => {
    ev.preventDefault();
    ev.stopPropagation();

    alert("App name cannot be saved currently");
  }, []);

  return (
    <form className={styles.root} onSubmit={onFormSubmit}>
      <TextField
        className={styles.textField}
        label={renderToString("GeneralSettings.app-name.label")}
        value={state.appName}
        onChange={onAppNameChange}
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
