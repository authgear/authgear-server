import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import deepEqual from "deep-equal";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";

import styles from "./UserInterfaceScreen.module.scss";
import {
  AppRawTemplatesUpdater,
  useUpdateAppRawTemplatesMutation,
} from "./mutations/updateAppRawTemplatesMutation";
import { STATIC_AUTHGEAR_CSS } from "../../templates";
import { useAppRawTemplatesQuery } from "./query/appRawTemplatesQuery";

interface UserInterfaceScreenState {
  customCss: string;
}

interface UserInterfaceProps {
  templates: Record<typeof STATIC_AUTHGEAR_CSS, string | null>;
  updateTemplates: AppRawTemplatesUpdater<typeof STATIC_AUTHGEAR_CSS>;
  isUpdatingTemplates: boolean;
}

function constructState(
  templates: UserInterfaceProps["templates"]
): UserInterfaceScreenState {
  return {
    customCss: templates[STATIC_AUTHGEAR_CSS] ?? "",
  };
}

const UserInterface: React.FC<UserInterfaceProps> = function UserInterface(
  props: UserInterfaceProps
) {
  const { templates, updateTemplates, isUpdatingTemplates } = props;

  const initialState = useMemo(() => constructState(templates), [templates]);

  const [state, setState] = useState<UserInterfaceScreenState>(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onCustomCssChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        customCss: value,
      }));
    },
    []
  );

  const onSaveButtonClicked = useCallback(() => {
    const updates: Partial<Record<typeof STATIC_AUTHGEAR_CSS, string>> = {
      [STATIC_AUTHGEAR_CSS]: state.customCss,
    };
    updateTemplates(updates).catch(() => {});
  }, [state, updateTemplates]);

  return (
    <div className={styles.form}>
      <Label className={styles.label}>
        <FormattedMessage id="UserInterfaceScreen.custom-css.label" />
      </Label>
      <CodeEditor
        className={styles.codeEditor}
        language="css"
        value={state.customCss}
        onChange={onCustomCssChange}
      />

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          disabled={!isFormModified}
          onClick={onSaveButtonClicked}
          loading={isUpdatingTemplates}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>

      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </div>
  );
};

const UserInterfaceScreen: React.FC = function UserInterfaceScreen() {
  const { appID } = useParams();

  const {
    updateAppRawTemplates,
    loading: isUpdatingTemplates,
    error: updateTemplatesError,
  } = useUpdateAppRawTemplatesMutation<typeof STATIC_AUTHGEAR_CSS>(appID);

  const {
    templates,
    loading: isLoadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppRawTemplatesQuery(appID, STATIC_AUTHGEAR_CSS);

  if (isLoadingTemplates) {
    return <ShowLoading />;
  }

  if (loadTemplatesError) {
    return <ShowError error={loadTemplatesError} onRetry={refetchTemplates} />;
  }

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: isUpdatingTemplates,
      })}
    >
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="UserInterfaceScreen.title" />
        </Text>
        <UserInterface
          templates={templates}
          updateTemplates={updateAppRawTemplates}
          isUpdatingTemplates={isUpdatingTemplates}
        />
      </div>
    </main>
  );
};

export default UserInterfaceScreen;
