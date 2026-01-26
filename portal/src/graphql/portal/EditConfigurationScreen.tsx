import React, { useCallback, useMemo, useState } from "react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { FormattedMessage } from "../../intl";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react/lib/Dialog";

import styles from "./EditConfigurationScreen.module.css";
import { useParams, useNavigate } from "react-router-dom";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import {
  ResourcesFormState,
  useResourceForm,
} from "../../hook/useResourceForm";
import {
  expandSpecifier,
  Resource,
  ResourceSpecifier,
  specifierId,
} from "../../util/resource";
import { RESOURCE_AUTHGEAR_YAML } from "../../resources";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import DefaultButton from "../../DefaultButton";
import PrimaryButton from "../../PrimaryButton";

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;
}

interface FormState extends ResourcesFormState {}
const AUTHGEAR_YAML_RESOURCE_SPECIFIER: ResourceSpecifier = {
  def: RESOURCE_AUTHGEAR_YAML,
  locale: null,
  extension: null,
};

const EditConfigurationScreen: React.VFC = function EditConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const specifiers = [AUTHGEAR_YAML_RESOURCE_SPECIFIER];
  const resourceForm = useResourceForm(appID, specifiers);
  const { refetch } = useAppAndSecretConfigQuery(appID);

  const [isWarningDialogVisible, setWarningDialogVisible] = useState(true);

  const onDismiss = useCallback(() => {
    setWarningDialogVisible(false);
  }, []);

  const onCancel = useCallback(() => {
    navigate(-1);
  }, [navigate]);

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="EditConfigurationScreen.warning.title" />
      ) as unknown as string,
      subText: (
        <FormattedMessage id="EditConfigurationScreen.warning.content" />
      ) as unknown as string,
      showCloseButton: false,
    };
  }, []);

  const form: FormModel = useMemo(
    () => ({
      ...resourceForm,
      save: async (...args: Parameters<(typeof resourceForm)["save"]>) => {
        await resourceForm.save(...args);
        await refetch();
      },
    }),
    [refetch, resourceForm]
  );

  const rawAuthgearYAML = useMemo(() => {
    const resource =
      form.state.resources[specifierId(AUTHGEAR_YAML_RESOURCE_SPECIFIER)];
    if (resource == null) {
      return null;
    }
    if (resource.nullableValue == null) {
      return null;
    }
    return resource.nullableValue;
  }, [form.state.resources]);

  const onChange = useCallback(
    (value: string | undefined, _e: unknown) => {
      const resource: Resource = {
        specifier: AUTHGEAR_YAML_RESOURCE_SPECIFIER,
        path: expandSpecifier(AUTHGEAR_YAML_RESOURCE_SPECIFIER),
        nullableValue: value,
        effectiveData: value,
      };
      const updatedResources = {
        [specifierId(AUTHGEAR_YAML_RESOURCE_SPECIFIER)]: resource,
      };
      form.setState(() => {
        return {
          resources: updatedResources,
        };
      });
    },
    [form]
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  const authgearYAMLSections: [EditTemplatesWidgetSection] = [
    {
      key: "authgear.yaml",
      title: null,
      items: [
        {
          key: "authgear.yaml",
          title: null,
          editor: "code",
          language: "yaml",

          value: rawAuthgearYAML ?? "",
          onChange,
        },
      ],
    },
  ];

  return (
    <FormContainer form={form}>
      <Dialog
        hidden={!isWarningDialogVisible}
        dialogContentProps={dialogContentProps}
        modalProps={{
          isBlocking: true,
          styles: {
            main: { "@media (min-width: 480px)": { maxWidth: "500px" } },
          },
        }}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <DefaultButton
            onClick={onCancel}
            text={<FormattedMessage id="cancel" />}
          />
          <PrimaryButton
            onClick={onDismiss}
            text={
              <FormattedMessage id="EditConfigurationScreen.warning.confirm" />
            }
          />
        </DialogFooter>
      </Dialog>
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="EditConfigurationScreen.title" />
        </ScreenTitle>
        <EditTemplatesWidget
          className={styles.widget}
          codeEditorClassname={styles.codeEditor}
          sections={authgearYAMLSections}
        />
      </ScreenContent>
    </FormContainer>
  );
};

export default EditConfigurationScreen;
