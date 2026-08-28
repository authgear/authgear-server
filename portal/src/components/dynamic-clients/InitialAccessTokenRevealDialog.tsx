import React, { useCallback, useMemo } from "react";
import { Dialog, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { Callout } from "../v2/Callout/Callout";
import { TextField } from "../v2/TextField/TextField";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { CodeField } from "../common/CodeField";
import styles from "./InitialAccessTokenRevealDialog.module.css";

// Minimal shell-snippet highlighter for the fixed curl example below — not a
// general shell tokenizer. It colors single-quoted strings, flags, and the
// leading curl command; everything else keeps the default code color.
function highlightCurl(code: string): React.ReactNode {
  // Split on single-quoted strings; odd indices are the quoted strings.
  // No `s` flag (the TS target predates it); [^'\\] already spans newlines,
  // which the multi-line -d '{...}' body needs.
  const parts = code.split(/('(?:[^'\\]|\\[^])*')/);
  return parts.map((part, index) => {
    if (index % 2 === 1) {
      return (
        <span key={index} className={styles.tokenString}>
          {part}
        </span>
      );
    }
    // Within plain shell text, color the curl command and -X flags.
    const subparts = part.split(/(^curl\b|\s-[A-Za-z]\b)/m);
    return subparts.map((subpart, subindex) => {
      const key = `${index}-${subindex}`;
      if (/^curl\b/.test(subpart)) {
        return (
          <span key={key} className={styles.tokenCommand}>
            {subpart}
          </span>
        );
      }
      if (/^\s-[A-Za-z]\b/.test(subpart)) {
        return (
          <span key={key} className={styles.tokenFlag}>
            {subpart}
          </span>
        );
      }
      return <React.Fragment key={key}>{subpart}</React.Fragment>;
    });
  });
}

export interface InitialAccessTokenRevealDialogProps {
  // The plaintext token to reveal; the dialog is open while this is non-null.
  token: string | null;
  // The project's client registration endpoint, for the example request.
  registrationEndpoint: string;
  onDismiss: () => void;
}

export function InitialAccessTokenRevealDialog({
  token,
  registrationEndpoint,
  onDismiss,
}: InitialAccessTokenRevealDialogProps): React.ReactElement {
  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        onDismiss();
      }
    },
    [onDismiss]
  );

  const exampleCurl = useMemo(() => {
    return `curl ${registrationEndpoint} \\
  -H 'Content-Type: application/json' \\
  -H 'Authorization: Bearer ${token ?? ""}' \\
  -d '{
    "client_name": "My client",
    "redirect_uris": ["https://example.com/callback"]
  }'`;
  }, [registrationEndpoint, token]);

  // Derived from the very string the copy button copies, so the highlighted
  // view can never drift from what is copied.
  const highlightedCurl = useMemo(
    () => highlightCurl(exampleCurl),
    [exampleCurl]
  );

  return (
    <Dialog.Root open={token != null} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="640px" size="3">
        <Dialog.Title>
          <FormattedMessage id="InitialAccessTokenRevealDialog.title" />
        </Dialog.Title>
        <Dialog.Description size="2" color="gray">
          <FormattedMessage id="InitialAccessTokenRevealDialog.description" />
        </Dialog.Description>
        <div className={styles.content}>
          <TextField
            size="2"
            label={
              <FormattedMessage id="InitialAccessTokenRevealDialog.token.label" />
            }
            value={token ?? ""}
            readOnly={true}
            suffixPlain={true}
            suffix={<CopyIconButton textToCopy={token ?? ""} />}
          />
          <Callout
            type="warning"
            showCloseButton={false}
            text={
              <FormattedMessage id="InitialAccessTokenRevealDialog.warning" />
            }
          />
          <div>
            <div className={styles.exampleHeader}>
              <Text as="p" size="2" weight="medium">
                <FormattedMessage id="InitialAccessTokenRevealDialog.example.label" />
              </Text>
              <CopyIconButton textToCopy={exampleCurl} />
            </div>
            <CodeField codeClassName={styles.exampleCode}>
              {highlightedCurl}
            </CodeField>
          </div>
        </div>
        <div className={styles.actions}>
          <PrimaryButton
            size="2"
            text={<FormattedMessage id="InitialAccessTokenRevealDialog.done" />}
            onClick={onDismiss}
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
