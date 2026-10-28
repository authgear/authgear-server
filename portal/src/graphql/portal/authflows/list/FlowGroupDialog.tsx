import React, { useContext, useState } from "react";
import { Button, Dialog, Select, Text, TextField } from "@radix-ui/themes";
import { Context } from "../../../../intl";
import {
  AuthenticationFlowsConfig,
  UIAuthenticationFlowGroup,
} from "../../../../types";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import { FlowType, displayNameOf, flowTypes } from "../model";
import { flowDisplayName, flowTypeMessageIDs } from "../labels";
import { NameError } from "./actions";
import { GroupSelection, flowNamesOfType, groupSelection } from "./groups";
import styles from "./FlowGroupDialog.module.css";

export const groupNameErrorMessageIDs: Record<NameError, string> = {
  invalid: "AuthFlowsScreen.list.name.invalid",
  taken: "AuthFlowsScreen.groups.name.taken",
  reserved: "AuthFlowsScreen.groups.name.reserved",
};

export interface FlowGroupDialogProps {
  // null for a new group, which also asks for a name.
  group: UIAuthenticationFlowGroup | null;
  flows: AuthenticationFlowsConfig | undefined;
  validateName: (name: string) => NameError | null;
  onSubmit: (name: string, selection: GroupSelection) => void;
  onCancel: () => void;
}

// New group, or a group's flows. Mounted per use, so it starts from the
// group each time.
export function FlowGroupDialog({
  group,
  flows,
  validateName,
  onSubmit,
  onCancel,
}: FlowGroupDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const isNew = group == null;
  const [name, setName] = useState(group?.name ?? "");
  const [selection, setSelection] = useState(() => groupSelection(group));
  const nameErr = isNew ? validateName(name) : null;
  const setFlow = (type: FlowType, flowName: string) =>
    setSelection((s) => ({ ...s, [type]: flowName }));
  return (
    <Dialog.Root
      open={true}
      onOpenChange={(o) => {
        if (!o) {
          onCancel();
        }
      }}
    >
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          {isNew
            ? renderToString("AuthFlowsScreen.groups.new.title")
            : renderToString("AuthFlowsScreen.groups.edit.title", {
                name: group.name,
              })}
        </Dialog.Title>
        <Dialog.Description size="2" color="gray">
          {renderToString("AuthFlowsScreen.groups.dialog.description")}
        </Dialog.Description>
        <form
          className={styles.form}
          onSubmit={(e) => {
            e.preventDefault();
            if (nameErr == null) {
              onSubmit(name, selection);
            }
          }}
        >
          {isNew ? (
            <label className={styles.field}>
              <Text size="2" weight="medium">
                {renderToString("AuthFlowsScreen.list.name")}
              </Text>
              <TextField.Root
                value={name}
                onChange={(e) => setName(e.target.value)}
                autoFocus={true}
              />
              {nameErr != null && name !== "" ? (
                <Text size="1" color="red">
                  {renderToString(groupNameErrorMessageIDs[nameErr])}
                </Text>
              ) : null}
            </label>
          ) : null}
          <div className={styles.flows}>
            {flowTypes.map((type) => {
              const names = flowNamesOfType(flows, type);
              const value = selection[type];
              return (
                <React.Fragment key={type}>
                  <Text size="2">
                    {renderToString(flowTypeMessageIDs[type])}
                  </Text>
                  <Select.Root
                    size="2"
                    value={value}
                    onValueChange={(v) => setFlow(type, v)}
                  >
                    <Select.Trigger />
                    <Select.Content>
                      {/* A flow the group names but that no longer exists
                          stays visible, so it can be changed. */}
                      {(names.includes(value) ? names : [...names, value]).map(
                        (n) => (
                          <Select.Item key={n} value={n}>
                            {flowDisplayName(
                              n,
                              renderToString,
                              displayNameOf(flows, type, n)
                            )}
                          </Select.Item>
                        )
                      )}
                    </Select.Content>
                  </Select.Root>
                </React.Fragment>
              );
            })}
          </div>
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              text={renderToString("AuthFlowsScreen.edit.cancel")}
              onClick={onCancel}
            />
            <Button type="submit" size="2" disabled={nameErr != null}>
              {renderToString(
                isNew
                  ? "AuthFlowsScreen.groups.create"
                  : "AuthFlowsScreen.groups.save"
              )}
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
