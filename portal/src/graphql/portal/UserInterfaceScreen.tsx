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
import { RESOURCE_AUTHGEAR_CSS, PATH_AUTHGEAR_CSS } from "../../resources";
import {
  AppTemplatesUpdater,
  useUpdateAppTemplatesMutation,
} from "./mutations/updateAppTemplatesMutation";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import { Resource } from "../../util/resource";

interface UserInterfaceScreenState {
  customCss: string;
}

interface UserInterfaceProps {
  resources: Resource[];
  updateTemplates: AppTemplatesUpdater;
  isUpdatingTemplates: boolean;
}

function constructState(resources: Resource[]): UserInterfaceScreenState {
  return {
    customCss:
      resources.find((r) => r.specifier.def === RESOURCE_AUTHGEAR_CSS)?.value ??
      "",
  };
}

const UserInterface: React.FC<UserInterfaceProps> = function UserInterface(
  props: UserInterfaceProps
) {
  const { resources, updateTemplates, isUpdatingTemplates } = props;

  const initialState = useMemo(() => constructState(resources), [resources]);

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
    const specifier = { def: RESOURCE_AUTHGEAR_CSS };
    updateTemplates(
      [specifier],
      [
        {
          specifier,
          path: PATH_AUTHGEAR_CSS,
          value: state.customCss.length === 0 ? null : state.customCss,
        },
      ]
    ).catch(() => {});
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
    updateAppTemplates,
    loading: isUpdatingTemplates,
    error: updateTemplatesError,
  } = useUpdateAppTemplatesMutation(appID);

  const {
    resources,
    loading: isLoadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery(appID, [{ def: RESOURCE_AUTHGEAR_CSS }]);

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
          resources={resources}
          updateTemplates={updateAppTemplates}
          isUpdatingTemplates={isUpdatingTemplates}
        />
      </div>
    </main>
  );
};

export default UserInterfaceScreen;
