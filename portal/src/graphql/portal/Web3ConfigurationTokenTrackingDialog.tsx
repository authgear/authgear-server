import React, { useCallback, useContext, useMemo, useState } from "react";
import { Dialog, DialogFooter, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import PrimaryButton from "../../PrimaryButton";
import TextField from "../../TextField";
import { convertToHexstring, parseHexstring } from "../../util/hex";

const TOKEN_ID_REGEX = /^0x[a-fA-F0-9]+$/;
const DIALOG_MAX_WIDTH = "33%";
const MAX_TOKEN_COUNT = 10;

interface Web3ConfigurationTokenTrackingDialogProps {
  isVisible: boolean;
  initialValue?: string[];
  onContinue: (tokenIDs: string[]) => void;
  onDismiss: () => void;
}
const Web3ConfigurationTokenTrackingDialog: React.VFC<
  Web3ConfigurationTokenTrackingDialogProps
> = (props) => {
  const { isVisible, initialValue, onContinue, onDismiss } = props;
  const [validationErrorId, setValidationErrorId] = useState<string | null>(
    null
  );

  const { renderToString } = useContext(Context);

  const defaultValue = useMemo(() => {
    if (!initialValue) {
      return [];
    }

    return initialValue.map((i) => parseHexstring(i));
  }, [initialValue]);

  const [value, setValue] = useState<string[]>(defaultValue);

  const dialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="Web3ConfigurationScreen.collection-list.add-collection.toke-tracking-dialog.title" />
      ),
      subText: renderToString(
        "Web3ConfigurationScreen.collection-list.add-collection.toke-tracking-dialog.description"
      ),
    };
  }, [renderToString]);

  const tokenIDs = useMemo(() => {
    return value.join("\n");
  }, [value]);

  const onTextChange = useCallback(
    (_, newValue?: string) => {
      setValidationErrorId(null);
      const tokens = newValue?.split("\n") ?? [];
      if (
        tokens.length <= 1 ||
        (tokens.length <= MAX_TOKEN_COUNT && tokens[tokens.length - 2])
      ) {
        setValue(tokens);
      }
    },
    [setValue]
  );

  const onProceed = useCallback(() => {
    const hexValues = value.map((v) => convertToHexstring(v));
    if (!hexValues.every((t) => TOKEN_ID_REGEX.test(t))) {
      setValidationErrorId("errors.invalid-token-id");
      return;
    }

    onContinue(hexValues);
  }, [value, onContinue]);

  const onCancel = useCallback(() => {
    setValue(defaultValue);

    onDismiss();
  }, [defaultValue, onDismiss]);

  return (
    <Dialog
      hidden={!isVisible}
      dialogContentProps={dialogContentProps}
      onDismiss={onDismiss}
      maxWidth={DIALOG_MAX_WIDTH}
    >
      <TextField
        placeholder={renderToString(
          "Web3ConfigurationScreen.collection-list.add-collection.toke-tracking-dialog.placeholder"
        )}
        multiline={true}
        resizable={false}
        rows={20}
        value={tokenIDs}
        errorMessage={
          validationErrorId != null
            ? renderToString(validationErrorId)
            : undefined
        }
        onChange={onTextChange}
      />
      <Text variant="small">
        <FormattedMessage
          id="Web3ConfigurationScreen.collection-list.add-collection.toke-tracking-dialog.item-count"
          values={{
            count: value.length,
            max: MAX_TOKEN_COUNT,
          }}
        />
      </Text>
      <DialogFooter>
        <PrimaryButton
          onClick={onProceed}
          text={<FormattedMessage id="continue" />}
        />
        <DefaultButton
          onClick={onCancel}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
};

export default Web3ConfigurationTokenTrackingDialog;
