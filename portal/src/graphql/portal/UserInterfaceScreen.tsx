import React, { useCallback, useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label } from "@fluentui/react";
import { useParams } from "react-router-dom";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import FormContainer from "../../FormContainer";
import { PATH_AUTHGEAR_CSS, RESOURCE_AUTHGEAR_CSS } from "../../resources";
import { Resource, ResourceSpecifier } from "../../util/resource";
import { ResourceFormModel, useResourceForm } from "../../hook/useResourceForm";
import styles from "./UserInterfaceScreen.module.scss";

const specifier: ResourceSpecifier = { def: RESOURCE_AUTHGEAR_CSS };

interface FormState {
  customCSS: string;
}

function constructFormState(resources: Resource[]): FormState {
  return {
    customCSS:
      resources.find((r) => r.specifier.def === RESOURCE_AUTHGEAR_CSS)?.value ??
      "",
  };
}

function constructResources(state: FormState): Resource[] {
  return [{ specifier, path: PATH_AUTHGEAR_CSS, value: state.customCSS }];
}

interface UserInterfaceContentProps {
  form: ResourceFormModel<FormState>;
}

const UserInterfaceContent: React.FC<UserInterfaceContentProps> = function UserInterfaceContent(
  props
) {
  const { state, setState } = props.form;

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: ".", label: <FormattedMessage id="UserInterfaceScreen.title" /> },
    ];
  }, []);

  const onCustomCSSChange = useCallback(
    (_, value?: string) => {
      setState((state) => ({
        ...state,
        customCSS: value ?? "",
      }));
    },
    [setState]
  );

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <Label className={styles.title}>
        <FormattedMessage id="UserInterfaceScreen.custom-css.label" />
      </Label>
      <CodeEditor
        className={styles.codeEditor}
        language="css"
        value={state.customCSS}
        onChange={onCustomCSSChange}
      />
    </div>
  );
};

const UserInterfaceScreen: React.FC = function UserInterfaceScreen() {
  const { appID } = useParams();
  const specifiers = useMemo(() => [specifier], []);
  const form = useResourceForm(
    appID,
    specifiers,
    constructFormState,
    constructResources
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <UserInterfaceContent form={form} />
    </FormContainer>
  );
};

export default UserInterfaceScreen;
