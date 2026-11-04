import React, { useContext, useState } from "react";
import { Button, Dialog, Tabs } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { StringConstraintsFields } from "../../StringConstraintsFields";
import { StandardAttributesValidationConfig } from "../../types";
import {
  ADDRESS_VALIDATION_POINTERS,
  StringConstraints,
  StringConstraintsDraft,
  constraintsFromDraft,
  draftFromConstraints,
} from "../../util/stringConstraints";
import styles from "./StandardAttributeValidationDialog.module.css";

export interface ValidationUpdate {
  pointer: string;
  constraints: StringConstraints;
  // The pattern's message; "" when there is none.
  message: string;
}

// The address attribute is validated part by part.
export function validationPointersOf(rowPointer: string): readonly string[] {
  return rowPointer === "/address" ? ADDRESS_VALIDATION_POINTERS : [rowPointer];
}

const addressPartMessageIDs: Record<string, string> = {
  "/address/formatted": "StandardAttributeValidationDialog.address.formatted",
  "/address/street_address":
    "StandardAttributeValidationDialog.address.street_address",
  "/address/locality": "StandardAttributeValidationDialog.address.locality",
  "/address/region": "StandardAttributeValidationDialog.address.region",
  "/address/postal_code":
    "StandardAttributeValidationDialog.address.postal_code",
};

export interface StandardAttributeValidationDialogProps {
  // The row's pointer: an attribute, or /address for its parts.
  rowPointer: string;
  validation: StandardAttributesValidationConfig[];
  getMessage: (pointer: string) => string;
  onApply: (updates: ValidationUpdate[]) => void;
  onCancel: () => void;
}

// Edits the row's constraints locally; Apply hands them to the page, which
// saves them with the Save bar. Mounted per use.
export function StandardAttributeValidationDialog({
  rowPointer,
  validation,
  getMessage,
  onApply,
  onCancel,
}: StandardAttributeValidationDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const pointers = validationPointersOf(rowPointer);
  const [drafts, setDrafts] = useState<Record<string, StringConstraintsDraft>>(
    () =>
      Object.fromEntries(
        pointers.map((p) => [
          p,
          draftFromConstraints(validation.find((v) => v.pointer === p)),
        ])
      )
  );
  const [messages, setMessages] = useState<Record<string, string>>(() =>
    Object.fromEntries(pointers.map((p) => [p, getMessage(p)]))
  );
  const [tab, setTab] = useState(pointers[0]);
  const fieldName = rowPointer.slice(1);

  const fields = (p: string) => (
    <StringConstraintsFields
      draft={drafts[p]}
      onChange={(d) => setDrafts((prev) => ({ ...prev, [p]: d }))}
      patternMessage={messages[p]}
      onPatternMessageChange={(m) =>
        setMessages((prev) => ({ ...prev, [p]: m }))
      }
      // A part with no entry yet has no errors to show.
      parentJSONPointer={`/user_profile/standard_attributes/validation/${
        validation.findIndex((v) => v.pointer === p) >= 0
          ? validation.findIndex((v) => v.pointer === p)
          : "-"
      }`}
    />
  );

  return (
    <Dialog.Root
      open={true}
      onOpenChange={(o) => {
        if (!o) {
          onCancel();
        }
      }}
    >
      <Dialog.Content maxWidth="560px" size="3">
        <Dialog.Title>
          <FormattedMessage
            id="StandardAttributeValidationDialog.title"
            values={{
              attribute: renderToString("standard-attribute." + fieldName),
            }}
          />
        </Dialog.Title>
        <Dialog.Description
          size="2"
          color="gray"
          className={styles.description}
        >
          <FormattedMessage id="StringConstraintsFields.scope" />
        </Dialog.Description>
        {pointers.length > 1 ? (
          <Tabs.Root value={tab} onValueChange={setTab}>
            <Tabs.List size="1">
              {pointers.map((p) => (
                <Tabs.Trigger key={p} value={p}>
                  {renderToString(addressPartMessageIDs[p])}
                </Tabs.Trigger>
              ))}
            </Tabs.List>
            {pointers.map((p) => (
              <Tabs.Content key={p} value={p} className={styles.tab}>
                {fields(p)}
              </Tabs.Content>
            ))}
          </Tabs.Root>
        ) : (
          fields(pointers[0])
        )}
        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            text={renderToString("cancel")}
            onClick={onCancel}
          />
          <Button
            size="2"
            onClick={() =>
              onApply(
                pointers.map((p) => {
                  const constraints = constraintsFromDraft(drafts[p]);
                  return {
                    pointer: p,
                    constraints,
                    message:
                      (constraints.pattern ?? "") !== "" ? messages[p] : "",
                  };
                })
              )
            }
          >
            <FormattedMessage id="StandardAttributeValidationDialog.apply" />
          </Button>
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
