import React, { useCallback, useMemo, useState } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
  useErrorMessageBarContext,
} from "../../ErrorMessageBar";
import { parseRawError } from "../../error/parse";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { InitialAccessTokenType } from "../../graphql/adminapi/globalTypes.generated";
import { useInitialAccessTokensQueryQuery } from "../../graphql/adminapi/query/initialAccessTokensQuery.generated";
import { useCreateInitialAccessTokenMutationMutation } from "../../graphql/adminapi/mutations/createInitialAccessTokenMutation.generated";
import { useRevokeInitialAccessTokenMutationMutation } from "../../graphql/adminapi/mutations/revokeInitialAccessTokenMutation.generated";
import {
  InitialAccessTokenList,
  InitialAccessTokenListItem,
} from "./InitialAccessTokenList";
import { CreateInitialAccessTokenDialog } from "./CreateInitialAccessTokenDialog";
import { InitialAccessTokenRevealDialog } from "./InitialAccessTokenRevealDialog";
import styles from "./InitialAccessTokenSection.module.css";

function InitialAccessTokenSectionContent(): React.ReactElement {
  const { setErrors } = useErrorMessageBarContext();

  const { data, refetch } = useInitialAccessTokensQueryQuery();

  const [createInitialAccessToken, { loading: isCreating }] =
    useCreateInitialAccessTokenMutationMutation();
  const [revokeInitialAccessToken, { loading: isRevoking }] =
    useRevokeInitialAccessTokenMutationMutation();

  const [isCreateDialogVisible, setIsCreateDialogVisible] = useState(false);
  const [revealedToken, setRevealedToken] = useState<string | null>(null);
  const [tokenToRevoke, setTokenToRevoke] =
    useState<InitialAccessTokenListItem | null>(null);

  const tokens = useMemo(
    (): InitialAccessTokenListItem[] =>
      (data?.initialAccessTokens ?? []).map((token) => ({
        id: token.id,
        createdAt: String(token.createdAt),
        expiresAt: String(token.expiresAt),
        type: token.type,
      })),
    [data?.initialAccessTokens]
  );

  const onOpenCreateDialog = useCallback(() => {
    setIsCreateDialogVisible(true);
  }, []);

  const onCreate = useCallback(
    (type: InitialAccessTokenType, expiresIn: number) => {
      createInitialAccessToken({
        variables: { input: { type, expiresIn } },
      })
        .then(async (result) => {
          const token = result.data?.createInitialAccessToken.token;
          if (token != null) {
            setRevealedToken(token);
          }
          return refetch();
        })
        .catch((e: unknown) => {
          setErrors(parseRawError(e));
        })
        .finally(() => {
          // Close the dialog on failure too — the error message bar renders
          // behind the modal overlay and would otherwise be invisible.
          setIsCreateDialogVisible(false);
        });
    },
    [createInitialAccessToken, refetch, setErrors]
  );

  const onDismissReveal = useCallback(() => {
    setRevealedToken(null);
  }, []);

  const onRequestRevoke = useCallback((token: InitialAccessTokenListItem) => {
    setTokenToRevoke(token);
  }, []);

  const onCancelRevoke = useCallback(() => {
    setTokenToRevoke(null);
  }, []);

  const onRevokeDialogOpenChange = useCallback((open: boolean) => {
    if (!open) {
      setTokenToRevoke(null);
    }
  }, []);

  const onConfirmRevoke = useCallback(() => {
    if (tokenToRevoke == null) {
      return;
    }
    revokeInitialAccessToken({
      variables: { input: { id: tokenToRevoke.id } },
    })
      .then(async () => refetch())
      .catch((e: unknown) => {
        setErrors(parseRawError(e));
      })
      .finally(() => {
        // Close the dialog on failure too — the error message bar renders
        // behind the modal overlay and would otherwise be invisible.
        setTokenToRevoke(null);
      });
  }, [tokenToRevoke, revokeInitialAccessToken, refetch, setErrors]);

  return (
    <div className={styles.section}>
      <ErrorMessageBar />
      <div className={styles.header}>
        <div className={styles.headerText}>
          <Text as="p" size="2" weight="medium">
            <FormattedMessage id="InitialAccessTokenSection.title" />
          </Text>
          <Text as="p" size="1" color="gray">
            <FormattedMessage id="InitialAccessTokenSection.description" />
          </Text>
        </div>
        <SecondaryButton
          size="2"
          text={<FormattedMessage id="InitialAccessTokenSection.create" />}
          onClick={onOpenCreateDialog}
        />
      </div>
      <InitialAccessTokenList tokens={tokens} onRevoke={onRequestRevoke} />
      <CreateInitialAccessTokenDialog
        open={isCreateDialogVisible}
        onOpenChange={setIsCreateDialogVisible}
        onCreate={onCreate}
        loading={isCreating}
      />
      <InitialAccessTokenRevealDialog
        token={revealedToken}
        onDismiss={onDismissReveal}
      />
      <ConfirmationDialog
        open={tokenToRevoke != null}
        onOpenChange={onRevokeDialogOpenChange}
        title={
          <FormattedMessage id="InitialAccessTokenSection.revoke.confirm.title" />
        }
        description={
          <FormattedMessage id="InitialAccessTokenSection.revoke.confirm.description" />
        }
        confirmText={
          <FormattedMessage id="InitialAccessTokenSection.revoke.confirm.confirm" />
        }
        cancelText={
          <FormattedMessage id="InitialAccessTokenSection.revoke.confirm.cancel" />
        }
        onConfirm={onConfirmRevoke}
        onCancel={onCancelRevoke}
        loading={isRevoking}
      />
    </div>
  );
}

export function InitialAccessTokenSection(): React.ReactElement {
  return (
    <ErrorMessageBarContextProvider>
      <InitialAccessTokenSectionContent />
    </ErrorMessageBarContextProvider>
  );
}
