import React, { useMemo, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { FormattedMessage } from "../../intl";
import { produce } from "immer";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
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
} from "../../EditCustomAttributeForm";

type FormState = CustomAttributeDraft;

const EMPTY_STATE: FormState = {
  pointer: "",
  type: "",
  minimum: "",
  maximum: "",
  enum: [],
};

interface EditCustomAttributeContentProps {
  form: AppConfigFormModel<FormState>;
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
    });
  };
}

function EditCustomAttributeContent(props: EditCustomAttributeContentProps) {
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
        label: <FormattedMessage id="EditCustomAttributeScreen.title" />,
      },
    ];
  }, []);

  const onChangeDraft = (draft: FormState) => {
    setState(() => draft);
  };

  return (
    <ScreenContent>
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
      <EditCustomAttributeForm
        className={styles.widget}
        mode="edit"
        index={index}
        draft={state}
        onChangeDraft={onChangeDraft}
      />
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

    const form = useAppConfigForm({
      appID,
      constructFormState: makeConstructFormState(index),
      constructConfig: makeConstructConfig(index),
    });

    const afterSave = useCallback(() => {
      navigate("./../..");
    }, [navigate]);

    if (isNaN(index)) {
      return null;
    }

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form} afterSave={afterSave}>
        <EditCustomAttributeContent form={form} index={index} />
      </FormContainer>
    );
  };

export default EditCustomAttributeScreen;
