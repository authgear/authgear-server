import React, { useCallback, useEffect, useState } from "react";
import { Dialog, Flex, RadioGroup, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { Callout } from "../v2/Callout/Callout";
import { InitialAccessTokenType } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./CreateInitialAccessTokenDialog.module.css";

const EXPIRES_IN_OPTIONS = [3600, 86400, 604800, 2592000] as const;
export type ExpiresInSeconds = (typeof EXPIRES_IN_OPTIONS)[number];

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
  const [expiresIn, setExpiresIn] = useState<ExpiresInSeconds>(3600);

  useEffect(() => {
    if (open) {
      // Reset the selection every time the dialog is (re)opened.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setTokenType(InitialAccessTokenType.ThirdParty);
      setExpiresIn(3600);
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
    const parsed = Number(newValue);
    for (const option of EXPIRES_IN_OPTIONS) {
      if (parsed === option) {
        setExpiresIn(option);
      }
    }
  }, []);

  const onSubmit = useCallback(() => {
    onCreate(tokenType, expiresIn);
  }, [onCreate, tokenType, expiresIn]);

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
              value={expiresIn.toFixed(0)}
              onValueChange={onExpiresInChange}
            >
              <Flex direction="column" gap="2">
                {EXPIRES_IN_OPTIONS.map((option) => (
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
              </Flex>
            </RadioGroup.Root>
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
