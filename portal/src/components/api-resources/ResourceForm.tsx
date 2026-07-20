import React, { useContext, useCallback, useEffect, useId } from "react";
import { Flex, IconButton, Text, TextField as RadixTextField } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import cn from "classnames";
import styles from "./ResourceForm.module.css";
import { Context, FormattedMessage } from "../../intl";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useFormTopErrors } from "../../form";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useLoading } from "../../hook/loading";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { TextField } from "../v2/TextField/TextField";
import { FormField } from "../v2/FormField/FormField";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";

export interface ResourceFormState {
  name: string;
  resourceURI: string;
}

export interface ResourceFormProps {
  className?: string;
  mode: "create" | "edit";
  state: ResourceFormState;
  setState: (fn: (state: ResourceFormState) => ResourceFormState) => void;
}

const URI_SCHEME_RE = /^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//;

export function sanitizeFormState(state: ResourceFormState): ResourceFormState {
  const resourceURI = state.resourceURI.trim();
  const hasScheme = URI_SCHEME_RE.test(resourceURI);

  return {
    name: state.name.trim(),
    resourceURI: resourceURI
      ? hasScheme
        ? resourceURI
        : `https://${resourceURI}`
      : "",
  };
}

function isFormComplete(state: ResourceFormState): boolean {
  const s = sanitizeFormState(state);
  return Boolean(s.name && s.resourceURI);
}

export const ResourceForm: React.VFC<ResourceFormProps> =
  function ResourceForm({ className, state, setState, mode }) {
    const { renderToString } = useContext(Context);
    const resourceURIId = useId();
    const handleNameChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        setState((s) => ({ ...s, name: e.target.value })),
      [setState]
    );
    const handleResourceURIChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        setState((s) => {
          let resourceURI = e.target.value;
          resourceURI = resourceURI.replace(/^(\s*)https:\/\//, "");
          return { ...s, resourceURI };
        }),
      [setState]
    );
    const { onSubmit, canSave, isUpdating } = useFormContainerBaseContext();

    useLoading(isUpdating);

    const errors = useFormTopErrors();
    const { setErrors } = useErrorMessageBarContext();
    useEffect(() => {
      setErrors(errors);
    }, [errors, setErrors]);

    const resourceURILabel = (
      <Flex display="inline-flex" align="center" gap="1">
        <span>
          <FormattedMessage id="ResourceForm.resourceURI.label" />
        </span>
        <span className={styles.requiredMark} aria-hidden="true">
          *
        </span>
        <Tooltip
          content={
            <FormattedMessage id="ResourceForm.resourceURI.tooltip" />
          }
        >
          <IconButton
            type="button"
            variant="ghost"
            color="gray"
            size="1"
            aria-label={renderToString("ResourceForm.resourceURI.tooltip")}
          >
            <InfoCircledIcon width="1rem" height="1rem" />
          </IconButton>
        </Tooltip>
      </Flex>
    );

    return (
      <form className={cn(styles.root, className)} onSubmit={onSubmit}>
        <div className={styles.formFields}>
          <TextField
            size="2"
            required={true}
            label={<FormattedMessage id="ResourceForm.name.label" />}
            hint={<FormattedMessage id="ResourceForm.name.description" />}
            fieldName="name"
            parentJSONPointer=""
            type="text"
            value={state.name}
            onChange={handleNameChange}
          />
          {mode === "edit" ? (
            <TextField
              size="2"
              label={<FormattedMessage id="ResourceForm.resourceURI.label" />}
              value={state.resourceURI}
              readOnly={true}
              type="text"
              suffixPlain={true}
              suffix={<CopyIconButton textToCopy={state.resourceURI} />}
            />
          ) : (
            <FormField
              size="2"
              label={resourceURILabel}
              htmlFor={resourceURIId}
              hint={
                <FormattedMessage id="ResourceForm.resourceURI.description" />
              }
              fieldName="resourceURI"
              parentJSONPointer=""
            >
              <RadixTextField.Root
                id={resourceURIId}
                size="2"
                variant="surface"
                type="text"
                value={state.resourceURI}
                onChange={handleResourceURIChange}
              >
                <RadixTextField.Slot side="left">
                  <Text size="2" color="gray">
                    https://
                  </Text>
                </RadixTextField.Slot>
              </RadixTextField.Root>
            </FormField>
          )}
        </div>
        {mode !== "edit" ? (
          <PrimaryButton
            size="2"
            type="submit"
            text={renderToString("create")}
            disabled={!canSave || !isFormComplete(state)}
            loading={isUpdating}
          />
        ) : null}
      </form>
    );
  };
