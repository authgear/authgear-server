import React, { useMemo, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { FormattedMessage } from "../../intl";
import { v4 as uuidv4 } from "uuid";
import { produce } from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
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
} from "../../EditCustomAttributeForm";
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
      access_control: currentState.access_control,
    });
  });
}

interface CreateCustomAttributeContentProps {
  form: AppConfigFormModel<FormState>;
  index: number;
}

function CreateCustomAttributeContent(
  props: CreateCustomAttributeContentProps
) {
  const { index, form } = props;
  const { state, setState } = form;

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: "~/configuration/user-profile/custom-attributes",
        label: (
          <FormattedMessage id="CustomAttributesConfigurationScreen.title" />
        ),
      },
      {
        to: ".",
        label: <FormattedMessage id="CreateCustomAttributeScreen.title" />,
      },
    ];
  }, []);

  const onChangeDraft = (draft: CustomAttributeDraft) => {
    setState((prev) => {
      return {
        ...prev,
        ...draft,
      };
    });
  };

  return (
    <ScreenContent>
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
      <EditCustomAttributeForm
        className={styles.widget}
        mode="new"
        index={index}
        draft={state}
        onChangeDraft={onChangeDraft}
      />
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

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

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
      <FormContainer form={form} afterSave={afterSave}>
        <CreateCustomAttributeContent form={form} index={index} />
      </FormContainer>
    );
  };

export default CreateCustomAttributeScreen;
