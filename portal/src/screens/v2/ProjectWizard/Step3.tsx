import React, { useCallback, useMemo, useState } from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { ProjectWizardStepper } from "../../../components/project-wizard/ProjectWizardStepper";
import { ProjectWizardFormModel } from "./form";
import { FormField } from "../../../components/v2/FormField/FormField";
import { ProjectWizardBackButton } from "../../../components/project-wizard/ProjectWizardBackButton";
import {
  ImageInput,
  ImageInputError,
  ImageInputErrorCode,
} from "../../../components/v2/ImageInput/ImageInput";
import { produce } from "immer";
import {
  ColorHex,
  ColorInput,
} from "../../../components/v2/ColorInput/ColorInput";

export function Step3(): React.ReactElement {
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();

  const [imageFieldError, setImageFieldError] =
    useState<React.ReactNode | null>(null);

  const handleImageChange = useCallback(
    (imageBase64DataURL: string) => {
      setImageFieldError(null);
      form.setState((prev) =>
        produce(prev, (draft) => {
          draft.logoBase64DataURL = imageBase64DataURL;
          return draft;
        })
      );
    },
    [form]
  );

  const handleImageError = useCallback((err: ImageInputError) => {
    switch (err.code) {
      case ImageInputErrorCode.FILE_TOO_LARGE:
        setImageFieldError(<FormattedMessage id="errors.image-too-large" />);
        break;

      default:
        setImageFieldError(
          <FormattedMessage
            id="errors.unknown"
            values={{ message: String(err.internalError ?? "") }}
          />
        );
        break;
    }
  }, []);

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

  return (
    <div className="grid grid-cols-1 gap-12 text-left self-stretch">
      <ProjectWizardStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-6 max-w-[400px]">
        <Text.Heading>
          <FormattedMessage
            id="ProjectWizardScreen.step3.header"
            values={{ projectName: form.state.projectName }}
          />
        </Text.Heading>
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.logo.label" />
          }
          error={imageFieldError}
        >
          <ImageInput
            value={form.state.logoBase64DataURL ?? null}
            onValueChange={handleImageChange}
            onError={handleImageError}
          />
        </FormField>
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.buttonAndLinkColor.label" />
          }
        >
          <ColorInput
            value={form.state.buttonAndLinkColor}
            onValueChange={handleColorChange.buttonAndLinkColor}
          />
        </FormField>
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.buttonLabelColor.label" />
          }
        >
          <ColorInput
            value={form.state.buttonLabelColor}
            onValueChange={handleColorChange.buttonLabelColor}
          />
        </FormField>
      </div>
      <div className="grid grid-flow-col grid-rows-1 gap-8 items-center justify-start">
        <ProjectWizardBackButton onClick={form.toPreviousStep} />
        <PrimaryButton
          type="submit"
          size="3"
          text={<FormattedMessage id="ProjectWizardScreen.actions.done" />}
          onClick={form.save}
          disabled={!form.canSave}
        />
      </div>
    </div>
  );
}
