import React, { useCallback, useMemo } from "react";
import { Text } from "../../../components/project-wizard/Text";
import { FormattedMessage } from "../../../intl";
import { PrimaryButton } from "../../../components/v2/Button/PrimaryButton/PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { ProjectWizardStepper } from "../../../components/project-wizard/ProjectWizardStepper";
import { ProjectWizardFormModel } from "./form";
import { FormField } from "../../../components/v2/FormField/FormField";
import { ProjectWizardBackButton } from "../../../components/project-wizard/ProjectWizardBackButton";
import {
  ImageInput,
  ImageValue,
} from "../../../components/v2/ImageInput/ImageInput";
import { produce } from "immer";
import {
  ColorHex,
  ColorPickerField,
} from "../../../components/v2/ColorPickerField/ColorPickerField";
import { useCapture } from "../../../gtm_v2";

export function Step3(): React.ReactElement {
  const capture = useCapture();
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();

  const handleImageChange = useCallback(
    (imageValue: ImageValue | null) => {
      form.setState((prev) =>
        produce(prev, (draft) => {
          draft.logo = imageValue;
          return draft;
        })
      );
    },
    [form]
  );

  const handleColorChange = useMemo(() => {
    const fnFactory = (stateKey: "buttonAndLinkColor" | "buttonLabelColor") => {
      const fn = (newColor: ColorHex) => {
        form.setState((prev) =>
          produce(prev, (draft) => {
            draft[stateKey] = newColor;
            return draft;
          })
        );
      };
      return fn;
    };
    return {
      buttonAndLinkColor: fnFactory("buttonAndLinkColor"),
      buttonLabelColor: fnFactory("buttonLabelColor"),
    };
  }, [form]);

  const trackButtonColorEvent = useCallback(() => {
    capture("projectWizard.clicked-button-color");
  }, [capture]);

  const trackLabelColorEvent = useCallback(() => {
    capture("projectWizard.clicked-label-color");
  }, [capture]);

  return (
    <div className="grid grid-cols-1 gap-12 text-left self-stretch">
      <ProjectWizardStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-6 max-w-[400px]">
        <div className="grid grid-cols-1 gap-1">
          <Text.Heading>
            <FormattedMessage
              id="ProjectWizardScreen.step3.header"
              values={{ projectName: form.state.projectName }}
            />
          </Text.Heading>
          <Text.Subheading>
            <FormattedMessage
              id="ProjectWizardScreen.step3.subheader"
              values={{ projectName: form.state.projectName }}
            />
          </Text.Subheading>
        </div>
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.logo.label" />
          }
        >
          <ImageInput
            sizeLimitInBytes={100 * 1000}
            value={form.state.logo}
            onValueChange={handleImageChange}
            onClickUpload={useCallback(() => {
              capture("projectWizard.clicked-upload");
            }, [capture])}
          />
        </FormField>
        <ColorPickerField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.buttonAndLinkColor.label" />
          }
          value={form.state.buttonAndLinkColor}
          onValueChange={handleColorChange.buttonAndLinkColor}
          onOpenPicker={trackButtonColorEvent}
          onFocus={trackButtonColorEvent}
        />
        <ColorPickerField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.buttonLabelColor.label" />
          }
          value={form.state.buttonLabelColor}
          onValueChange={handleColorChange.buttonLabelColor}
          onOpenPicker={trackLabelColorEvent}
          onFocus={trackLabelColorEvent}
        />
      </div>
      <div className="grid grid-flow-col grid-rows-1 gap-8 items-center justify-start">
        <ProjectWizardBackButton onClick={form.toPreviousStep} />
        <PrimaryButton
          type="submit"
          size="3"
          text={<FormattedMessage id="ProjectWizardScreen.actions.done" />}
          loading={form.isUpdating}
          onClick={form.save}
          disabled={!form.canSave}
        />
      </div>
    </div>
  );
}
