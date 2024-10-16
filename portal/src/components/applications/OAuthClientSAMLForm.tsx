import React, { useCallback, useContext, useMemo } from "react";
import Toggle from "../../Toggle";
import {
  FormattedMessage,
  Context as MessageFormatContext,
  ContextValue as MessageFormatContextValue,
} from "@oursky/react-messageformat";
import HorizontalDivider from "../../HorizontalDivider";
import WidgetTitle from "../../WidgetTitle";
import ScreenTitle from "../../ScreenTitle";
import { IChoiceGroupOption, ChoiceGroup } from "@fluentui/react";
import { SAMLNameIDFormat, SAMLNameIDAttributePointer } from "../../types";

export interface OAuthClientSAMLFormState {
  isSAMLEnabled: boolean;
  nameIDFormat: SAMLNameIDFormat;
  nameIDAttributePointer: SAMLNameIDAttributePointer | undefined;
}

export interface OAuthClientSAMLFormProps {
  formState: OAuthClientSAMLFormState;
  onFormStateChange: (newState: OAuthClientSAMLFormState) => void;
}

const nameIDFormatOptions: IChoiceGroupOption[] = [
  { key: SAMLNameIDFormat.Unspecified, text: SAMLNameIDFormat.Unspecified },
  { key: SAMLNameIDFormat.EmailAddress, text: SAMLNameIDFormat.EmailAddress },
];

function makeNameIDAttributePointerOptions(
  renderToString: MessageFormatContextValue["renderToString"]
): IChoiceGroupOption[] {
  return [
    {
      key: SAMLNameIDAttributePointer.Sub,
      text: renderToString(
        "OAuthClientSAMLForm.nameIDAttribute.options.userID"
      ),
    },
    {
      key: SAMLNameIDAttributePointer.Email,
      text: renderToString("OAuthClientSAMLForm.nameIDAttribute.options.email"),
    },
    {
      key: SAMLNameIDAttributePointer.PhoneNumber,
      text: renderToString("OAuthClientSAMLForm.nameIDAttribute.options.phone"),
    },
    {
      key: SAMLNameIDAttributePointer.PreferredUsername,
      text: renderToString(
        "OAuthClientSAMLForm.nameIDAttribute.options.username"
      ),
    },
  ];
}

export function OAuthClientSAMLForm({
  formState,
  onFormStateChange,
}: OAuthClientSAMLFormProps): React.ReactElement {
  const { renderToString } = useContext(MessageFormatContext);

  const onIsSAMLEnabledChange = useCallback(
    (_, checked?: boolean) => {
      onFormStateChange({ ...formState, isSAMLEnabled: Boolean(checked) });
    },
    [formState, onFormStateChange]
  );

  const onNameIDFormatChange = useCallback(
    (_, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      onFormStateChange({
        ...formState,
        nameIDFormat: option.key as SAMLNameIDFormat,
      });
    },
    [formState, onFormStateChange]
  );

  const onNameIDAttributePointerChange = useCallback(
    (_, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      onFormStateChange({
        ...formState,
        nameIDAttributePointer: option.key as SAMLNameIDAttributePointer,
      });
    },
    [formState, onFormStateChange]
  );

  const nameIDAttributePointerOptions = useMemo(
    () => makeNameIDAttributePointerOptions(renderToString),
    [renderToString]
  );

  return (
    <div>
      <Toggle
        label={renderToString("OAuthClientSAMLForm.enable.label")}
        description={renderToString("OAuthClientSAMLForm.enable.description")}
        checked={formState.isSAMLEnabled}
        onChange={onIsSAMLEnabledChange}
      />
      {formState.isSAMLEnabled ? (
        <>
          <HorizontalDivider className="my-12" />
          <div className="grid gap-y-12 grid-cols-1">
            <ScreenTitle>
              <FormattedMessage id="OAuthClientSAMLForm.screen.title" />
            </ScreenTitle>
            <div className="grid gap-y-3 grid-cols-1">
              <WidgetTitle className="mb-3" id="basic">
                <FormattedMessage id="OAuthClientSAMLForm.basic.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <ChoiceGroup
                  label={renderToString(
                    "OAuthClientSAMLForm.nameIDFormat.label"
                  )}
                  options={nameIDFormatOptions}
                  selectedKey={formState.nameIDFormat}
                  onChange={onNameIDFormatChange}
                />
                <ChoiceGroup
                  label={renderToString(
                    "OAuthClientSAMLForm.nameIDAttribute.label"
                  )}
                  disabled={
                    formState.nameIDFormat !== SAMLNameIDFormat.Unspecified
                  }
                  options={nameIDAttributePointerOptions}
                  selectedKey={
                    formState.nameIDFormat !== SAMLNameIDFormat.Unspecified
                      ? null
                      : formState.nameIDAttributePointer
                  }
                  onChange={onNameIDAttributePointerChange}
                />
              </div>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
}
