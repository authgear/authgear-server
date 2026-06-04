import React, { useCallback, useMemo, useRef, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import ScreenContent from "../../ScreenContent";
import { FormattedMessage } from "../../intl";
import CodeEditor from "../../CodeEditor";

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
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";

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

interface EditConfigurationContentProps {
  rawAuthgearYAML: string | null;
  onChange: (value: string | undefined, e: unknown) => void;
  isWarningDialogVisible: boolean;
  onDismissWarning: () => void;
  onCancelWarning: () => void;
}

const EditConfigurationContent: React.VFC<EditConfigurationContentProps> =
  function EditConfigurationContent(props) {
    const {
      rawAuthgearYAML,
      onChange,
      isWarningDialogVisible,
      onDismissWarning,
      onCancelWarning,
    } = props;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    return (
      <>
        <ConfirmationDialog
          open={isWarningDialogVisible}
          onOpenChange={() => {}}
          maxWidth="500px"
          title={
            <FormattedMessage id="EditConfigurationScreen.warning.title" />
          }
          description={
            <FormattedMessage
              id="EditConfigurationScreen.warning.content"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                br: () => <br />,
              }}
            />
          }
          confirmText={
            <FormattedMessage id="EditConfigurationScreen.warning.confirm" />
          }
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={onDismissWarning}
          onCancel={onCancelWarning}
          confirmColor="indigo"
        />
        <ScreenContent
          className={cn(isDirty ? styles.contentWithSaveBar : null)}
        >
          <div
            ref={contentWidthAnchorRef}
            className={styles.contentWidthAnchor}
            aria-hidden
          />
          <div className={cn(styles.widget, styles.pageHeader)}>
            <h1 className={styles.pageTitle}>
              <FormattedMessage id="EditConfigurationScreen.title" />
            </h1>
          </div>
          <div
            className={cn(
              styles.widget,
              styles.editorCard,
              isDirty && styles.settingsCardSaveBarClearance
            )}
          >
            <div className={styles.editorCardHeader}>
              <Text as="p" size="3" weight="medium">
                <FormattedMessage id="EditConfigurationScreen.config.label" />
              </Text>
            </div>
            <CodeEditor
              className={styles.codeEditor}
              language="yaml"
              value={rawAuthgearYAML ?? ""}
              onChange={onChange}
            />
          </div>
          <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
        </ScreenContent>
      </>
    );
  };

const EditConfigurationScreen: React.VFC = function EditConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const specifiers = [AUTHGEAR_YAML_RESOURCE_SPECIFIER];
  const resourceForm = useResourceForm(appID, specifiers);
  const { refetch } = useAppAndSecretConfigQuery(appID);

  const [isWarningDialogVisible, setWarningDialogVisible] = useState(true);

  const onDismissWarning = useCallback(() => {
    setWarningDialogVisible(false);
  }, []);

  const onCancelWarning = useCallback(() => {
    navigate(-1);
  }, [navigate]);

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

  return (
    <FormContainer form={form} hideFooterComponent={true}>
      <EditConfigurationContent
        rawAuthgearYAML={rawAuthgearYAML}
        onChange={onChange}
        isWarningDialogVisible={isWarningDialogVisible}
        onDismissWarning={onDismissWarning}
        onCancelWarning={onCancelWarning}
      />
    </FormContainer>
  );
};

export default EditConfigurationScreen;
