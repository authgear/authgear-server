import React, { useCallback } from "react";
import FormContainer from "../../FormContainer";
import { useSimpleForm } from "../../hook/useSimpleForm";

interface FormState {}

const defaultState: FormState = {};

const AddRolesScreen: React.VFC = function AddRolesScreen() {
  const submit = useCallback(async () => {}, []);

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
  });

  return <FormContainer form={form} hideCommandBar={true}></FormContainer>;
};

export default AddRolesScreen;
