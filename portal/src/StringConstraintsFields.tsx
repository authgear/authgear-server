import React, { useContext, useMemo } from "react";
import { CheckboxCards, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "./intl";
import { TextField } from "./components/v2/TextField/TextField";
import { checkNonNegativeIntegerInput } from "./util/input";
import { makeValidationErrorCustomMessageIDRule } from "./error/parse";
import {
  ALLOWED_CHARACTERS,
  AllowedCharacter,
  StringConstraintsDraft,
} from "./util/stringConstraints";
import styles from "./StringConstraintsFields.module.css";

const characterExampleMessageIDs: Record<AllowedCharacter, string> = {
  letters: "StringConstraintsFields.characters.letters.example",
  digits: "StringConstraintsFields.characters.digits.example",
  spaces: "StringConstraintsFields.characters.spaces.example",
  symbols: "StringConstraintsFields.characters.symbols.example",
};

const characterMessageIDs: Record<AllowedCharacter, string> = {
  letters: "StringConstraintsFields.characters.letters",
  digits: "StringConstraintsFields.characters.digits",
  spaces: "StringConstraintsFields.characters.spaces",
  symbols: "StringConstraintsFields.characters.symbols",
};

// A minimum and a maximum side by side, each labelled; empty is no limit.
function MinMaxFields({
  label,
  hint,
  minimum,
  maximum,
  onMinimumChange,
  onMaximumChange,
  minimumFieldName,
  maximumFieldName,
  parentJSONPointer,
}: {
  label: string;
  hint: string;
  minimum: string;
  maximum: string;
  onMinimumChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onMaximumChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  minimumFieldName: string;
  maximumFieldName: string;
  parentJSONPointer: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  return (
    <div className={styles.field}>
      <Text as="p" size="2" weight="medium">
        {label}
      </Text>
      <div className={styles.minMax}>
        <TextField
          size="2"
          label={renderToString("StringConstraintsFields.minimum")}
          value={minimum}
          onChange={onMinimumChange}
          parentJSONPointer={parentJSONPointer}
          fieldName={minimumFieldName}
          placeholder={renderToString("StringConstraintsFields.no-limit")}
        />
        <TextField
          size="2"
          label={renderToString("StringConstraintsFields.maximum")}
          value={maximum}
          onChange={onMaximumChange}
          parentJSONPointer={parentJSONPointer}
          fieldName={maximumFieldName}
          placeholder={renderToString("StringConstraintsFields.no-limit")}
        />
      </div>
      <Text as="p" size="1" color="gray">
        {hint}
      </Text>
    </div>
  );
}

export interface StringConstraintsFieldsProps {
  draft: StringConstraintsDraft;
  onChange: (draft: StringConstraintsDraft) => void;
  // The message shown when the pattern does not match, in the fallback
  // language.
  patternMessage: string;
  onPatternMessageChange: (message: string) => void;
  // The config object the constraint keys are on, for its server errors.
  parentJSONPointer: string;
}

// Length, allowed characters and a pattern for a string attribute.
export function StringConstraintsFields({
  draft,
  onChange,
  patternMessage,
  onPatternMessageChange,
  parentJSONPointer,
}: StringConstraintsFieldsProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  // The server's error for a regex it can't compile is a format error; say
  // so here only, so other format errors keep their wording.
  const patternErrorRules = useMemo(
    () => [
      makeValidationErrorCustomMessageIDRule(
        "format",
        /\/pattern$/,
        "StringConstraintsFields.pattern.invalid"
      ),
    ],
    []
  );
  const set = (patch: Partial<StringConstraintsDraft>) =>
    onChange({ ...draft, ...patch });
  const onLength =
    (key: "minLength" | "maxLength") =>
    (e: React.ChangeEvent<HTMLInputElement>) => {
      // A length is never negative, so "-" is refused rather than kept
      // and then saved as no limit.
      if (checkNonNegativeIntegerInput(e.target.value)) {
        set({ [key]: e.target.value });
      }
    };
  // Every kind on is any character, which writes no constraint.
  const selected: string[] = draft.charactersLimited
    ? draft.allowedCharacters
    : [...ALLOWED_CHARACTERS];
  const isLastOn = (c: AllowedCharacter): boolean =>
    selected.length === 1 && selected[0] === c;
  const onCharacters = (values: string[]) => {
    const next = ALLOWED_CHARACTERS.filter((c) => values.includes(c));
    if (next.length === 0) {
      return;
    }
    set(
      next.length === ALLOWED_CHARACTERS.length
        ? { charactersLimited: false, allowedCharacters: [] }
        : { charactersLimited: true, allowedCharacters: next }
    );
  };

  return (
    <div className={styles.root}>
      <MinMaxFields
        label={renderToString("StringConstraintsFields.length")}
        hint={renderToString("StringConstraintsFields.length.hint")}
        minimum={draft.minLength}
        maximum={draft.maxLength}
        onMinimumChange={onLength("minLength")}
        onMaximumChange={onLength("maxLength")}
        minimumFieldName="min_length"
        maximumFieldName="max_length"
        parentJSONPointer={parentJSONPointer}
      />

      <div className={styles.field}>
        <Text as="p" size="2" weight="medium">
          <FormattedMessage id="StringConstraintsFields.characters" />
        </Text>
        <CheckboxCards.Root
          size="1"
          columns={{ initial: "2", sm: "4" }}
          value={selected}
          onValueChange={onCharacters}
        >
          {ALLOWED_CHARACTERS.map((c) => (
            <CheckboxCards.Item
              key={c}
              value={c}
              // The last kind on stays: none would allow nothing.
              disabled={isLastOn(c)}
            >
              <span className={styles.characterCard}>
                <Text size="2" weight="medium">
                  {renderToString(characterMessageIDs[c])}
                </Text>
                <Text size="1" color="gray">
                  {renderToString(characterExampleMessageIDs[c])}
                </Text>
              </span>
            </CheckboxCards.Item>
          ))}
        </CheckboxCards.Root>
        <Text as="p" size="1" color="gray">
          <FormattedMessage id="StringConstraintsFields.characters.hint" />
        </Text>
      </div>

      <div className={styles.field}>
        <TextField
          size="2"
          label={renderToString("StringConstraintsFields.pattern")}
          hint={renderToString("StringConstraintsFields.pattern.hint")}
          value={draft.pattern}
          onChange={(e) => set({ pattern: e.target.value })}
          parentJSONPointer={parentJSONPointer}
          fieldName="pattern"
          errorRules={patternErrorRules}
          inputClassName={styles.code}
          placeholder={renderToString(
            "StringConstraintsFields.pattern.placeholder"
          )}
        />
        {draft.pattern !== "" ? (
          <TextField
            size="2"
            label={renderToString("StringConstraintsFields.pattern-message")}
            hint={renderToString(
              "StringConstraintsFields.pattern-message.hint"
            )}
            value={patternMessage}
            onChange={(e) => onPatternMessageChange(e.target.value)}
          />
        ) : null}
      </div>
    </div>
  );
}

export interface NumberRangeFieldsProps {
  minimum: string;
  maximum: string;
  onChange: (minimum: string, maximum: string) => void;
  // Whether a typed value may stay in the field.
  checkInput: (value: string) => boolean;
  // The config object minimum and maximum are on, for their server errors.
  parentJSONPointer: string;
}

// The range of a number or integer attribute, laid out like a string's
// length.
export function NumberRangeFields({
  minimum,
  maximum,
  onChange,
  checkInput,
  parentJSONPointer,
}: NumberRangeFieldsProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  return (
    <div className={styles.root}>
      <MinMaxFields
        label={renderToString("StringConstraintsFields.range")}
        hint={renderToString("StringConstraintsFields.range.hint")}
        minimum={minimum}
        maximum={maximum}
        onMinimumChange={(e) => {
          if (checkInput(e.target.value)) {
            onChange(e.target.value, maximum);
          }
        }}
        onMaximumChange={(e) => {
          if (checkInput(e.target.value)) {
            onChange(minimum, e.target.value);
          }
        }}
        minimumFieldName="minimum"
        maximumFieldName="maximum"
        parentJSONPointer={parentJSONPointer}
      />
    </div>
  );
}

// The validation rules of an attribute, in a panel of their own. They are
// optional, and the note says when they are checked.
export function ValidationSection({
  children,
}: {
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <section className={styles.section}>
      <div className={styles.sectionHeader}>
        <div className={styles.sectionTitle}>
          <Text size="3" weight="bold">
            <FormattedMessage id="StringConstraintsFields.title" />
          </Text>
          <Text size="2" color="gray">
            <FormattedMessage id="StringConstraintsFields.optional" />
          </Text>
        </div>
        <Text as="p" size="1" color="gray">
          <FormattedMessage id="StringConstraintsFields.scope" />
        </Text>
      </div>
      {children}
    </section>
  );
}
