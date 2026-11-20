import React, { useContext } from "react";
import { createIntl } from "react-intl";
import { useParams } from "react-router-dom";
import {
  Checkbox,
  Flex,
  Select,
  Text,
  TextArea,
  TextField,
} from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../../../intl";
import Link from "../../../../Link";
import { unescapeMessageFormatText } from "../../../../util/messageFormat";
import { FlowType } from "../model";
import { useFormTextContext } from "./FormTextContext";
import { FormText } from "./useFormText";
import { formTextKey, isEditableText } from "./formText";
import { parseLinkLabel } from "./linkLabel";
import styles from "./StepForm.module.css";

const DEFAULT_FORM = "default";

// A stored translation as the end-user reads it; values such as FlowType
// pick the variant of a built-in text.
export function formatFormText(
  pattern: string,
  language: string,
  values: Record<string, string>
): string {
  try {
    return createIntl({
      locale: language,
      messages: { text: pattern },
      onError: () => {},
    }).formatMessage({ id: "text" }, values);
  } catch {
    return unescapeMessageFormatText(pattern);
  }
}

// A form field's label in the selected language, as text: the form's own
// label, else fallback (the attribute's label).
export function useFormFieldLabel(
  formName: string | undefined
): (pointer: string, fallback: string) => string {
  const ctx = useFormTextContext();
  return (pointer, fallback) => {
    if (ctx == null || formName == null || formName === "") {
      return fallback;
    }
    const own = ctx.text.ownText(
      ctx.language,
      formTextKey(formName, { label: pointer })
    );
    if (own == null || own === "") {
      return fallback;
    }
    const plain = unescapeMessageFormatText(own);
    const segments = parseLinkLabel(plain);
    return segments != null ? segments.map((x) => x.text).join("") : plain;
  };
}

// A form page's title or description as the end-user reads it in a
// language: the form's own text, else the default form's, else the built-in
// text. "" when the description is hidden.
export function formPageText(
  text: FormText,
  language: string,
  formName: string | undefined,
  flowType: FlowType,
  part: "title" | "description"
): string | undefined {
  const own =
    formName != null && formName !== ""
      ? text.effectiveText(language, formTextKey(formName, part))
      : undefined;
  const pattern =
    own ?? text.effectiveText(language, formTextKey(DEFAULT_FORM, part));
  return pattern != null
    ? formatFormText(pattern, language, { FlowType: flowType })
    : undefined;
}

// The language the form's text is edited in. The list labels languages
// with text of the form's own. The Authflow API returns no text, so the
// note says a Custom UI is not affected.
export function FormLanguageSelect({
  formName,
}: {
  formName: string | undefined;
}): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const ctx = useFormTextContext();
  if (ctx == null) {
    return null;
  }
  const named = formName != null && formName !== "";
  return (
    <div className={styles.field}>
      <Text as="div" size="2" weight="medium" className={styles.label}>
        {renderToString("AuthFlowsScreen.form-text.language")}
      </Text>
      <Select.Root value={ctx.language} onValueChange={ctx.setLanguage}>
        {/* The language alone; the list says which ones are edited. */}
        <Select.Trigger>
          {renderToString(`Locales.${ctx.language}`)}
        </Select.Trigger>
        <Select.Content>
          {ctx.languages.map((l) => (
            <Select.Item key={l} value={l}>
              {renderToString(`Locales.${l}`)}
              {named && ctx.text.hasText(l, formName) ? (
                <Text size="1" color="gray" className={styles.edited}>
                  {renderToString("AuthFlowsScreen.form-text.language.edited")}
                </Text>
              ) : null}
            </Select.Item>
          ))}
        </Select.Content>
      </Select.Root>
      <Text as="p" size="1" color="gray">
        {renderToString("AuthFlowsScreen.form-text.auth-ui-only")}
      </Text>
    </div>
  );
}

// Shown in place of a text box for text the builder cannot edit safely.
export function NotEditable(): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  return (
    <Text as="p" size="1" color="gray">
      <FormattedMessage
        id="AuthFlowsScreen.form-text.not-editable"
        values={{
          // eslint-disable-next-line react/no-unstable-nested-components
          link: (chunks: React.ReactNode) => (
            <Link
              to={`/project/${encodeURIComponent(appID)}/branding/custom-text`}
            >
              {chunks}
            </Link>
          ),
        }}
      />
    </Text>
  );
}

export interface FormTextInputProps {
  label: string;
  // The key under the form's current name; null while the form has none.
  textKey: string | null;
  // Grey text when the box is empty: the text in effect.
  placeholder: string | undefined;
  multiline?: boolean;
  disabled?: boolean;
  // Names the form on the first edit, and returns the key to write.
  keyForEdit: () => string;
}

// A plain-text translation of the selected language.
export function FormTextInput({
  label,
  textKey,
  placeholder,
  multiline,
  disabled,
  keyForEdit,
}: FormTextInputProps): React.ReactElement | null {
  const ctx = useFormTextContext();
  if (ctx == null) {
    return null;
  }
  const { text, language, readOnly } = ctx;
  const own = textKey != null ? text.ownText(language, textKey) : undefined;
  // From the saved value: what is being typed never locks the box.
  const saved = textKey != null ? text.savedText(language, textKey) : undefined;
  const editable = saved == null || isEditableText(saved, false);
  const value = own != null ? unescapeMessageFormatText(own) : "";
  const onChange = (v: string) =>
    // An empty box returns to the text in effect.
    text.setText(language, keyForEdit(), v === "" ? null : v);
  return (
    <div className={styles.field}>
      <Text as="label" size="2" weight="medium" className={styles.textLabel}>
        {label}
        {!editable ? (
          <NotEditable />
        ) : multiline ? (
          <TextArea
            size="2"
            value={value}
            placeholder={placeholder}
            disabled={readOnly || disabled}
            onChange={(e) => onChange(e.target.value)}
          />
        ) : (
          <TextField.Root
            size="2"
            value={value}
            placeholder={placeholder}
            disabled={readOnly || disabled}
            onChange={(e) => onChange(e.target.value)}
          />
        )}
      </Text>
    </div>
  );
}

export interface FormTextFieldsProps {
  formName: string | undefined;
  flowType: FlowType;
  // The form's name, naming it first when it has none.
  ensureName: () => string;
}

// A form page's title and description, in the selected language.
export function FormTextFields({
  formName,
  flowType,
  ensureName,
}: FormTextFieldsProps): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const ctx = useFormTextContext();
  if (ctx == null) {
    return null;
  }
  const { text, language, readOnly } = ctx;
  const named = formName != null && formName !== "";
  const keyOf = (part: "title" | "description") =>
    named ? formTextKey(formName, part) : null;
  const inEffect = (part: "title" | "description") =>
    formPageText(text, language, formName, flowType, part);
  const descriptionKey = keyOf("description");
  const hidden =
    descriptionKey != null && text.ownText(language, descriptionKey) === "";
  return (
    <>
      <FormTextInput
        label={renderToString("AuthFlowsScreen.form-text.title")}
        textKey={keyOf("title")}
        placeholder={inEffect("title")}
        keyForEdit={() => formTextKey(ensureName(), "title")}
      />
      <FormTextInput
        label={renderToString("AuthFlowsScreen.form-text.description")}
        textKey={hidden ? null : descriptionKey}
        placeholder={hidden ? undefined : inEffect("description")}
        multiline={true}
        disabled={hidden}
        keyForEdit={() => formTextKey(ensureName(), "description")}
      />
      <Text as="label" size="2">
        <Flex gap="2" align="center">
          <Checkbox
            checked={hidden}
            disabled={readOnly}
            onCheckedChange={(checked) => {
              const key = formTextKey(ensureName(), "description");
              if (checked === true) {
                // An empty description hides it.
                text.setText(language, key, "");
              } else if (text.savedText(language, key) === "") {
                text.setText(language, key, null);
              } else {
                // Brings back the saved description.
                text.clearEdit(language, key);
              }
            }}
          />
          {renderToString("AuthFlowsScreen.form-text.hide-description")}
        </Flex>
      </Text>
    </>
  );
}
