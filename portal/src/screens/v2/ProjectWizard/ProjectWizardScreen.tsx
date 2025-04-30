import React, { useCallback } from "react";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../../FormContainerBase";
import {
  FormState,
  ProjectWizardFormModel,
  ProjectWizardStep,
  useProjectWizardForm,
} from "./form";
import { ProjectWizardLayout } from "../../../components/project-wizard/ProjectWizardLayout";
import { Step1 } from "./Step1";
import { Step2 } from "./Step2";
import { Step3 } from "./Step3";
import { useOptionalAppContext } from "../../../context/AppContext";
import { usePortalClient } from "../../../graphql/portal/apollo";
import { useQuery } from "@apollo/client";
import {
  ScreenNavQueryDocument,
  ScreenNavQueryQuery,
  ScreenNavQueryQueryVariables,
} from "../../../graphql/portal/query/screenNavQuery.generated";
import ShowLoading from "../../../ShowLoading";
import ShowError from "../../../ShowError";

function Loaded({
  initialState,
}: {
  initialState: FormState | null;
}): React.ReactElement {
  const form = useProjectWizardForm(initialState);

  const handleFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();

      if (form.canSave) {
        form.save();
      }
    },
    [form]
  );

  if (form.isInitializing) {
    return <ShowLoading />;
  }

  if (form.initializeError) {
    return <ShowError error={form.initializeError} />;
  }

  return (
    <FormContainerBase form={form}>
      <ProjectWizardLayout>
        <form className="contents" onSubmit={handleFormSubmit}>
          <ProjectWizardScreenContent />
        </form>
      </ProjectWizardLayout>
    </FormContainerBase>
  );
}

function ProjectWizardScreenContent() {
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();
  switch (form.state.step) {
    case ProjectWizardStep.step1:
      return <Step1 />;
    case ProjectWizardStep.step2:
      return <Step2 />;
    case ProjectWizardStep.step3:
      return <Step3 />;
  }
}

function ProjectWizardScreen(): React.ReactElement {
  const appContext = useOptionalAppContext();
  const existingAppNodeID = appContext?.appNodeID;

  const client = usePortalClient();
  const skipAppSpecificQuery = existingAppNodeID == null;
  const screenNavQuery = useQuery<
    ScreenNavQueryQuery,
    ScreenNavQueryQueryVariables
  >(ScreenNavQueryDocument, {
    client,
    variables: {
      id: existingAppNodeID!,
    },
    fetchPolicy: "cache-first",
    skip: skipAppSpecificQuery,
  });

  if (skipAppSpecificQuery) {
    return <Loaded initialState={null} />;
  }

  if (screenNavQuery.loading || screenNavQuery.data == null) {
    return <ShowLoading />;
  }

  const initialState =
    screenNavQuery.data.node?.__typename === "App"
      ? screenNavQuery.data.node.tutorialStatus.data.project_wizard
      : null;

  return <Loaded initialState={initialState} />;
}

export default ProjectWizardScreen;
