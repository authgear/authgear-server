import React, { useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { FormattedMessage } from "../../intl";
import { Heading } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { v4 as uuidv4 } from "uuid";
import { produce } from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import Link from "../../Link";
import ScreenContent from "../../ScreenContent";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import {
  PortalAPIAppConfig,
  UserProfileAttributesAccessControl,
} from "../../types";
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
  patternErrorKey,
} from "../../util/stringConstraints";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import styles from "./CreateCustomAttributeScreen.module.css";

interface FormState extends CustomAttributeDraft {
  id: string;
  access_control: UserProfileAttributesAccessControl;
}

function constructFormState(): FormState {
  return {
    id: uuidv4(),
    pointer: "",
    type: "string",
    minimum: "",
    maximum: "",
    enum: [],
    constraints: EMPTY_CONSTRAINTS,
    access_control: {
      portal_ui: "readwrite",
      bearer: "hidden",
      end_user: "hidden",
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.user_profile ??= {};
    config.user_profile.custom_attributes ??= {};
    config.user_profile.custom_attributes.attributes ??= [];

    const minimum = parseFloat(currentState.minimum);
    const maximum = parseFloat(currentState.maximum);

    config.user_profile.custom_attributes.attributes.push({
      id: currentState.id,
      pointer: currentState.pointer,
      type: currentState.type as any,
      minimum: !isNaN(minimum) ? minimum : undefined,
      maximum: !isNaN(maximum) ? maximum : undefined,
      enum: currentState.type === "enum" ? currentState.enum : undefined,
      ...(currentState.type === "string"
        ? constraintsFromDraft(currentState.constraints)
        : {}),
      access_control: currentState.access_control,
    });
  });
}

interface CreateCustomAttributeContentProps {
  form: AppConfigFormModel<FormState>;
  translations: TranslationValuesForm;
  index: number;
}

function CreateCustomAttributeContent(
  props: CreateCustomAttributeContentProps
) {
  const { index, form, translations } = props;
  const { appID } = useParams() as { appID: string };
  const { state, setState } = form;
  const { canSave, isUpdating } = useFormContainerBaseContext();
  const navigate = useNavigate();

  const backURL = `/project/${appID}/configuration/user-profile/custom-attributes`;

  const patternKey = patternErrorKey("custom", state.pointer);
  const onChangeDraft = (draft: CustomAttributeDraft) => {
    syncPatternMessage(state, draft, translations);
    setState((prev) => {
      return {
        ...prev,
        ...draft,
      };
    });
  };

  const onCancel = useCallback(() => {
    navigate(backURL);
  }, [backURL, navigate]);

  return (
    <ScreenContent layout="list">
      <div className={styles.widget}>
        <Link to={backURL} className={styles.backLink}>
          <ChevronLeftIcon className={styles.backLinkIcon} />
          <span>
            <FormattedMessage id="CustomAttributesConfigurationScreen.title" />
          </span>
        </Link>
        <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="CreateCustomAttributeScreen.title" />
        </Heading>
      </div>
      <EditCustomAttributeForm
        className={styles.widget}
        mode="new"
        index={index}
        draft={state}
        onChangeDraft={onChangeDraft}
        patternMessage={translations.getValue(patternKey)}
        onPatternMessageChange={(m) => translations.setValue(patternKey, m)}
      />
      <div className={styles.actions}>
        <PrimaryButton
          size="2"
          disabled={!canSave || isUpdating}
          loading={isUpdating}
          type="submit"
          text={<FormattedMessage id="create" />}
        />
        <SecondaryButton
          size="2"
          disabled={isUpdating}
          type="button"
          onClick={onCancel}
          text={<FormattedMessage id="cancel" />}
        />
      </div>
    </ScreenContent>
  );
}

const CreateCustomAttributeScreen: React.VFC =
  function CreateCustomAttributeScreen() {
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();

    const afterSave = useCallback(() => {
      navigate("./..");
    }, [navigate]);

    const configForm = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
    const translations = useTranslationValuesForm(
      appID,
      configForm.isLoading
        ? null
        : configForm.effectiveConfig.localization?.fallback_language ?? "en"
    );
    const form = combineForms(configForm, translations.form);

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    const index =
      form.effectiveConfig.user_profile?.custom_attributes?.attributes
        ?.length ?? 0;

    return (
      <FormContainer form={form} afterSave={afterSave} canSave={true}>
        <CreateCustomAttributeContent
          form={form}
          translations={translations}
          index={index}
        />
      </FormContainer>
    );
  };

export default CreateCustomAttributeScreen;
