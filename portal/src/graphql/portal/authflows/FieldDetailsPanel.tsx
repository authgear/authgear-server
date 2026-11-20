import React, { useContext } from "react";
import {
  Code,
  Flex,
  Heading,
  IconButton,
  SegmentedControl,
  Text,
} from "@radix-ui/themes";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Context } from "../../../intl";
import Link from "../../../Link";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowStepConfig,
  CustomAttributesAttributeConfig,
} from "../../../types";
import { unescapeMessageFormatText } from "../../../util/messageFormat";
import { formFieldLabel } from "./sketches";
import { EditResult, updateStep } from "./edit/edits";
import { getAt } from "./edit/pointer";
import { useFormTextContext } from "./edit/FormTextContext";
import {
  FormTextInput,
  NotEditable,
  useFormFieldLabel,
} from "./edit/FormTextFields";
import { LinkLabelEditor } from "./edit/LinkLabelEditor";
import {
  PLACEHOLDER_FIELD_TYPES,
  formFieldType,
  formTextKey,
  isEditableText,
} from "./edit/formText";
import { parseLinkLabel } from "./edit/linkLabel";
import styles from "./FieldDetailsPanel.module.css";

// A field's type in words, by the type the Authflow API reports.
const fieldTypeMessageIDs: Partial<Record<string, string>> = {
  string: "AuthFlowsScreen.field.type.string",
  integer: "AuthFlowsScreen.field.type.integer",
  number: "AuthFlowsScreen.field.type.number",
  boolean: "AuthFlowsScreen.field.type.boolean",
  enum: "AuthFlowsScreen.field.type.enum",
  date: "AuthFlowsScreen.field.type.date",
  email: "AuthFlowsScreen.field.type.email",
  phone_number: "AuthFlowsScreen.field.type.phone_number",
  url: "AuthFlowsScreen.field.type.url",
  country_code: "AuthFlowsScreen.field.type.country_code",
  timezone: "AuthFlowsScreen.field.type.timezone",
  address: "AuthFlowsScreen.field.type.address",
};

export interface FieldDetailsPanelProps {
  appID: string;
  flow: AuthenticationFlowConfig;
  stepID: string;
  index: number;
  customAttributes: CustomAttributesAttributeConfig[];
  // Absent while viewing.
  apply?: (result: EditResult) => void;
  onClose: () => void;
}

// A form field's settings, beside the form step's panel.
export function FieldDetailsPanel({
  appID,
  flow,
  stepID,
  index,
  customAttributes,
  apply,
  onClose,
}: FieldDetailsPanelProps): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const ctx = useFormTextContext();
  const step = getAt(flow, stepID) as AuthenticationFlowStepConfig | undefined;
  const fieldLabel = useFormFieldLabel(step?.form?.name);
  const field = step?.form?.fields?.[index];
  const pointer = field?.user_profile?.pointer;
  if (ctx == null || step == null || pointer == null) {
    return null;
  }
  const { text, language } = ctx;
  const readOnly = ctx.readOnly || apply == null;
  const formName = step.form?.name ?? "";
  const custom = customAttributes.some((a) => a.pointer === pointer);
  const type = formFieldType(pointer, customAttributes);

  const ensureName = () => {
    if (formName !== "") {
      return formName;
    }
    const name = ctx.newFormName();
    apply?.(
      updateStep(flow, stepID, (s) => {
        s.form ??= {};
        s.form.name = name;
      })
    );
    return name;
  };

  // The attribute's own label, which the field shows unless the form sets
  // one.
  const attributeLabelKey = `${
    custom ? "custom" : "standard"
  }-attribute-label-${pointer}`;
  const attributeLabel = (() => {
    const pattern = text.effectiveText(language, attributeLabelKey);
    if (pattern == null) {
      return formFieldLabel(pointer, customAttributes, renderToString);
    }
    const plain = unescapeMessageFormatText(pattern);
    const segments = parseLinkLabel(plain);
    return segments != null ? segments.map((s) => s.text).join("") : plain;
  })();

  const labelKey =
    formName !== "" ? formTextKey(formName, { label: pointer }) : null;
  const ownLabel =
    labelKey != null ? text.ownText(language, labelKey) : undefined;
  const savedLabel =
    labelKey != null ? text.savedText(language, labelKey) : undefined;
  const heading = fieldLabel(pointer, attributeLabel);

  const settingsPath = `/project/${encodeURIComponent(
    appID
  )}/configuration/user-profile/${
    custom ? "custom-attributes" : "standard-attributes"
  }`;

  return (
    <aside className={styles.panel}>
      <Flex justify="between" align="center" gap="2">
        <Heading size="3" as="h2" className={styles.heading}>
          {heading}
        </Heading>
        <IconButton
          variant="ghost"
          color="gray"
          aria-label={renderToString("AuthFlowsScreen.panel.close")}
          onClick={onClose}
        >
          <Cross2Icon />
        </IconButton>
      </Flex>
      {type === "boolean" ? (
        <div className={styles.field}>
          <Text as="div" size="2" weight="medium">
            {renderToString("AuthFlowsScreen.form-text.label")}
          </Text>
          {savedLabel != null && !isEditableText(savedLabel, true) ? (
            <NotEditable />
          ) : (
            <LinkLabelEditor
              // The editor owns its content; remount it for another text.
              key={`${language}:${stepID}:${pointer}`}
              value={
                ownLabel != null ? unescapeMessageFormatText(ownLabel) : ""
              }
              placeholder={attributeLabel}
              disabled={readOnly}
              onChange={(html) =>
                text.setText(
                  language,
                  formTextKey(ensureName(), { label: pointer }),
                  html === "" ? null : html
                )
              }
            />
          )}
        </div>
      ) : (
        <FormTextInput
          label={renderToString("AuthFlowsScreen.form-text.label")}
          textKey={labelKey}
          placeholder={attributeLabel}
          disabled={readOnly}
          keyForEdit={() => formTextKey(ensureName(), { label: pointer })}
        />
      )}
      {PLACEHOLDER_FIELD_TYPES.includes(type) ? (
        <FormTextInput
          label={renderToString("AuthFlowsScreen.form-text.placeholder")}
          textKey={
            formName !== ""
              ? formTextKey(formName, { placeholder: pointer })
              : null
          }
          placeholder={renderToString(
            "AuthFlowsScreen.form-text.placeholder.hint"
          )}
          disabled={readOnly}
          keyForEdit={() => formTextKey(ensureName(), { placeholder: pointer })}
        />
      ) : null}
      <div className={styles.field}>
        <Text as="div" size="2" weight="medium">
          {renderToString("AuthFlowsScreen.form-text.answer")}
        </Text>
        <SegmentedControl.Root
          size="1"
          value={
            field?.user_profile?.required === true ? "required" : "optional"
          }
          disabled={readOnly}
          onValueChange={(v) =>
            apply?.(
              updateStep(flow, stepID, (s) => {
                const up = s.form?.fields?.[index]?.user_profile;
                if (up != null) {
                  up.required = v === "required";
                }
              })
            )
          }
        >
          <SegmentedControl.Item value="required">
            {renderToString("AuthFlowsScreen.edit.form.field.required")}
          </SegmentedControl.Item>
          <SegmentedControl.Item value="optional">
            {renderToString("AuthFlowsScreen.form-text.optional")}
          </SegmentedControl.Item>
        </SegmentedControl.Root>
      </div>
      <dl className={styles.facts}>
        <dt>{renderToString("AuthFlowsScreen.form-text.attribute")}</dt>
        <dd>
          <Link to={settingsPath}>
            <Code variant="ghost">{pointer}</Code>
          </Link>{" "}
          <Text color="gray">
            {renderToString(
              custom
                ? "AuthFlowsScreen.form-text.attribute.custom"
                : "AuthFlowsScreen.form-text.attribute.standard"
            )}
          </Text>
        </dd>
        <dt>{renderToString("AuthFlowsScreen.form-text.type")}</dt>
        <dd>
          {fieldTypeMessageIDs[type] != null
            ? renderToString(fieldTypeMessageIDs[type])
            : type}
        </dd>
      </dl>
    </aside>
  );
}
