import React, { useCallback, useEffect, useState } from "react";
import { Dialog, Flex, RadioGroup, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { Callout } from "../v2/Callout/Callout";
import { TextField } from "../v2/TextField/TextField";
import { InitialAccessTokenType } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./CreateInitialAccessTokenDialog.module.css";

const PRESET_EXPIRES_IN_OPTIONS = [3600, 86400, 604800, 2592000] as const;

type ExpiresInSelection = "3600" | "86400" | "604800" | "2592000" | "custom";

export interface CreateInitialAccessTokenDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreate: (type: InitialAccessTokenType, expiresIn: number) => void;
  loading: boolean;
}

export function CreateInitialAccessTokenDialog({
  open,
  onOpenChange,
  onCreate,
  loading,
}: CreateInitialAccessTokenDialogProps): React.ReactElement {
  const [tokenType, setTokenType] = useState<InitialAccessTokenType>(
    InitialAccessTokenType.ThirdParty
  );
  const [expiresInSelection, setExpiresInSelection] =
    useState<ExpiresInSelection>("3600");
  // Value of the datetime-local input, e.g. "2026-09-30T18:00".
  const [customExpiryValue, setCustomExpiryValue] = useState("");

  useEffect(() => {
    if (open) {
      // Reset the selection every time the dialog is (re)opened.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setTokenType(InitialAccessTokenType.ThirdParty);
      setExpiresInSelection("3600");
      setCustomExpiryValue("");
    }
  }, [open]);

  const onTokenTypeChange = useCallback((newValue: string) => {
    if (
      newValue === InitialAccessTokenType.ThirdParty ||
      newValue === InitialAccessTokenType.FirstParty
    ) {
      setTokenType(newValue);
    }
  }, []);

  const onExpiresInChange = useCallback((newValue: string) => {
    switch (newValue) {
      case "3600":
      case "86400":
      case "604800":
      case "2592000":
      case "custom":
        setExpiresInSelection(newValue);
        break;
      default:
        break;
    }
  }, []);

  // Seconds until the given custom expiry, or null when the value is empty,
  // unparsable, or not in the future. datetime-local values are interpreted
  // in the browser's local time zone, matching what the admin picked in the
  // native picker.
  const computeCustomExpiresInSeconds = useCallback(
    (value: string): number | null => {
      if (value === "") {
        return null;
      }
      const expiryMillis = new Date(value).getTime();
      if (Number.isNaN(expiryMillis)) {
        return null;
      }
      const seconds = Math.floor((expiryMillis - Date.now()) / 1000);
      if (seconds <= 0) {
        return null;
      }
      return seconds;
    },
    []
  );

  const [isCustomExpiryValid, setIsCustomExpiryValid] = useState(false);

  const onCustomExpiryChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setCustomExpiryValue(value);
      setIsCustomExpiryValid(computeCustomExpiresInSeconds(value) != null);
    },
    [computeCustomExpiresInSeconds]
  );

  const isCustomInvalid =
    expiresInSelection === "custom" && !isCustomExpiryValid;

  const onSubmit = useCallback(() => {
    if (expiresInSelection === "custom") {
      // Recompute at submit time: the expiry may have slipped into the past
      // while the dialog sat open.
      const seconds = computeCustomExpiresInSeconds(customExpiryValue);
      if (seconds == null) {
        setIsCustomExpiryValid(false);
        return;
      }
      onCreate(tokenType, seconds);
      return;
    }
    onCreate(tokenType, Number(expiresInSelection));
  }, [
    onCreate,
    tokenType,
    expiresInSelection,
    customExpiryValue,
    computeCustomExpiresInSeconds,
  ]);

  const onCancel = useCallback(() => {
    onOpenChange(false);
  }, [onOpenChange]);

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage id="CreateInitialAccessTokenDialog.title" />
        </Dialog.Title>
        <Dialog.Description size="2" color="gray">
          <FormattedMessage id="CreateInitialAccessTokenDialog.description" />
        </Dialog.Description>
        <Flex direction="column" gap="4" mt="4">
          <div className="flex flex-col gap-2">
            <Text as="p" size="2" weight="medium">
              <FormattedMessage id="CreateInitialAccessTokenDialog.type.label" />
            </Text>
            <RadioGroup.Root
              value={tokenType}
              onValueChange={onTokenTypeChange}
            >
              <Flex direction="column" gap="3">
                <Text as="label" size="2" className={styles.radioOption}>
                  <Flex gap="2" align="start">
                    <RadioGroup.Item
                      value={InitialAccessTokenType.ThirdParty}
                      className={styles.radioItem}
                    />
                    <div className={styles.radioContent}>
                      <Text as="span" size="2">
                        <FormattedMessage id="CreateInitialAccessTokenDialog.type.third-party.title" />
                      </Text>
                      <Text as="p" size="1" color="gray">
                        <FormattedMessage id="CreateInitialAccessTokenDialog.type.third-party.description" />
                      </Text>
                    </div>
                  </Flex>
                </Text>
                <Text as="label" size="2" className={styles.radioOption}>
                  <Flex gap="2" align="start">
                    <RadioGroup.Item
                      value={InitialAccessTokenType.FirstParty}
                      className={styles.radioItem}
                    />
                    <div className={styles.radioContent}>
                      <Text as="span" size="2">
                        <FormattedMessage id="CreateInitialAccessTokenDialog.type.first-party.title" />
                      </Text>
                      <Text as="p" size="1" color="gray">
                        <FormattedMessage id="CreateInitialAccessTokenDialog.type.first-party.description" />
                      </Text>
                    </div>
                  </Flex>
                </Text>
              </Flex>
            </RadioGroup.Root>
            {tokenType === InitialAccessTokenType.FirstParty ? (
              <Callout
                type="warning"
                text={
                  <FormattedMessage id="CreateInitialAccessTokenDialog.first-party.warning" />
                }
              />
            ) : null}
          </div>
          <div className="flex flex-col gap-2">
            <Text as="p" size="2" weight="medium">
              <FormattedMessage id="CreateInitialAccessTokenDialog.expires-in.label" />
            </Text>
            <RadioGroup.Root
              value={expiresInSelection}
              onValueChange={onExpiresInChange}
            >
              <Flex direction="column" gap="2">
                {PRESET_EXPIRES_IN_OPTIONS.map((option) => (
                  <Text
                    key={option}
                    as="label"
                    size="2"
                    className={styles.radioOption}
                  >
                    <Flex gap="2" align="center">
                      <RadioGroup.Item
                        value={option.toFixed(0)}
                        className={styles.radioItem}
                      />
                      <FormattedMessage
                        id={`CreateInitialAccessTokenDialog.expires-in.option.${option.toFixed(
                          0
                        )}`}
                      />
                    </Flex>
                  </Text>
                ))}
                <Text as="label" size="2" className={styles.radioOption}>
                  <Flex gap="2" align="center">
                    <RadioGroup.Item
                      value="custom"
                      className={styles.radioItem}
                    />
                    <FormattedMessage id="CreateInitialAccessTokenDialog.expires-in.option.custom" />
                  </Flex>
                </Text>
              </Flex>
            </RadioGroup.Root>
            {expiresInSelection === "custom" ? (
              <TextField
                size="2"
                labelSize="2"
                type="datetime-local"
                label={
                  <FormattedMessage id="CreateInitialAccessTokenDialog.expires-in.custom.label" />
                }
                hint={
                  <FormattedMessage id="CreateInitialAccessTokenDialog.expires-in.custom.hint" />
                }
                value={customExpiryValue}
                onChange={onCustomExpiryChange}
              />
            ) : null}
          </div>
        </Flex>
        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            disabled={loading}
            text={
              <FormattedMessage id="CreateInitialAccessTokenDialog.cancel" />
            }
            onClick={onCancel}
          />
          <PrimaryButton
            size="2"
            loading={loading}
            disabled={isCustomInvalid}
            text={
              <FormattedMessage id="CreateInitialAccessTokenDialog.create" />
            }
            onClick={onSubmit}
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
