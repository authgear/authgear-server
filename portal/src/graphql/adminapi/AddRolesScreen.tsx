import React, { useCallback } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";

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

  return <FormContainerBase form={form}></FormContainerBase>;
};

export default AddRolesScreen;
