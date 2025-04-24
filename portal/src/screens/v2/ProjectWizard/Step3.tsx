import React, { useCallback, useState } from "react";
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
          {/* TODO */}
        </FormField>{" "}
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step3.fields.buttonLabelColor.label" />
          }
        >
          {/* TODO */}
        </FormField>
      </div>
      <div className="grid grid-flow-col grid-rows-1 gap-8 items-center justify-start">
        <ProjectWizardBackButton onClick={form.toPreviousStep} />
        <PrimaryButton
          type="submit"
          size="3"
          text={<FormattedMessage id="ProjectWizardScreen.actions.next" />}
          onClick={form.toNextStep}
          disabled={!form.canNavigateToNextStep}
        />
      </div>
    </div>
  );
}
