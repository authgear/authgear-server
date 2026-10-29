import React, { useCallback, useRef, useMemo } from "react";
import { Navigate, useParams, useNavigate } from "react-router-dom";
import cn from "classnames";
import { FormattedMessage } from "../../intl";
import { Heading } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { produce } from "immer";
import ScreenContent from "../../ScreenContent";
import Link from "../../Link";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { PortalAPIAppConfig } from "../../types";
import styles from "./EditCustomAttributeScreen.module.css";
import EditCustomAttributeForm, {
  CustomAttributeDraft,
  EMPTY_CONSTRAINTS,
  syncPatternMessage,
} from "../../EditCustomAttributeForm";
import {
  TranslationValuesForm,
  combineForms,
  useTranslationValuesForm,
} from "../../hook/useTranslationValuesForm";
import {
  constraintsFromDraft,
  draftFromConstraints,
  patternErrorKey,
} from "../../util/stringConstraints";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";

type FormState = CustomAttributeDraft;

const EMPTY_STATE: FormState = {
  pointer: "",
  type: "",
  minimum: "",
  maximum: "",
  enum: [],
  constraints: EMPTY_CONSTRAINTS,
};

interface EditCustomAttributeContentProps {
  form: AppConfigFormModel<FormState>;
  translations: TranslationValuesForm;
  index: number;
}

function makeConstructFormState(
  index: number
): (config: PortalAPIAppConfig) => FormState {
  return (config: PortalAPIAppConfig): FormState => {
    const c = config.user_profile?.custom_attributes?.attributes?.[index];
    if (c == null) {
      return EMPTY_STATE;
    }
    return {
      pointer: c.pointer,
      type: c.type,
      minimum: c.minimum != null ? String(c.minimum) : "",
      maximum: c.maximum != null ? String(c.maximum) : "",
      enum: c.enum ?? [],
      constraints: draftFromConstraints(c),
    };
  };
}

function makeConstructConfig(
  index: number
): (
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
) => PortalAPIAppConfig {
  return (
    config: PortalAPIAppConfig,
    _initialState: FormState,
    currentState: FormState
  ): PortalAPIAppConfig => {
    return produce(config, (config) => {
      const c = config.user_profile?.custom_attributes?.attributes?.[index];
      if (c == null) {
        return;
      }

      c.pointer = currentState.pointer;
      if (currentState.type !== "") {
        c.type = currentState.type;
      }

      if (currentState.minimum === "") {
        c.minimum = undefined;
      } else {
        const minimum = parseFloat(currentState.minimum);
        if (!isNaN(minimum)) {
          c.minimum = minimum;
        }
      }

      if (currentState.maximum === "") {
        c.maximum = undefined;
      } else {
        const maximum = parseFloat(currentState.maximum);
        if (!isNaN(maximum)) {
          c.maximum = maximum;
        }
      }

      if (c.type === "enum") {
        c.enum = currentState.enum;
      }

      // The four keys are written fresh, so a cleared one goes.
      delete c.min_length;
      delete c.max_length;
      delete c.allowed_characters;
      delete c.pattern;
      if (c.type === "string") {
        Object.assign(c, constraintsFromDraft(currentState.constraints));
      }
    });
  };
}

function EditCustomAttributeContent(props: EditCustomAttributeContentProps) {
  const { index, form, translations } = props;
  const { appID } = useParams() as { appID: string };
  const { state, setState } = form;
  const { getIsDirty } = useFormContainerBaseContext();
  const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
  const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

  const backURL = `/project/${appID}/configuration/user-profile/custom-attributes`;

  const patternKey = patternErrorKey("custom", state.pointer);
  const onChangeDraft = (draft: FormState) => {
    syncPatternMessage(state, draft, translations);
    setState(() => draft);
  };

  return (
    <ScreenContent
      layout="list"
      className={cn(isDirty ? styles.contentWithSaveBar : null)}
    >
      <div ref={contentWidthAnchorRef} className={styles.widget}>
        <Link to={backURL} className={styles.backLink}>
          <ChevronLeftIcon className={styles.backLinkIcon} />
          <span>
            <FormattedMessage id="CustomAttributesConfigurationScreen.title" />
          </span>
        </Link>
        <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="EditCustomAttributeScreen.title" />
        </Heading>
      </div>
      <EditCustomAttributeForm
        className={styles.widget}
        mode="edit"
        index={index}
        draft={state}
        onChangeDraft={onChangeDraft}
        patternMessage={translations.getValue(patternKey)}
        onPatternMessageChange={(m) => translations.setValue(patternKey, m)}
      />
      <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
    </ScreenContent>
  );
}

const EditCustomAttributeScreen: React.VFC =
  function EditCustomAttributeScreen() {
    const { appID, index: indexString } = useParams() as {
      appID: string;
      index: string;
    };
    const navigate = useNavigate();

    const index = parseInt(indexString, 10);

    const configForm = useAppConfigForm({
      appID,
      constructFormState: makeConstructFormState(index),
      constructConfig: makeConstructConfig(index),
    });
    const translations = useTranslationValuesForm(
      appID,
      configForm.isLoading
        ? null
        : configForm.effectiveConfig.localization?.fallback_language ?? "en"
    );
    const form = combineForms(configForm, translations.form);

    const afterSave = useCallback(() => {
      navigate("./../..");
    }, [navigate]);

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (
      isNaN(index) ||
      form.effectiveConfig.user_profile?.custom_attributes?.attributes?.[
        index
      ] == null
    ) {
      // The attribute does not exist in this project (e.g. after switching
      // projects); fall back to the custom attribute list.
      return (
        <Navigate
          to={`/project/${appID}/configuration/user-profile/custom-attributes`}
          replace={true}
        />
      );
    }

    return (
      <FormContainer form={form} afterSave={afterSave}>
        <EditCustomAttributeContent
          form={form}
          translations={translations}
          index={index}
        />
      </FormContainer>
    );
  };

export default EditCustomAttributeScreen;
