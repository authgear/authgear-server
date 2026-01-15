import React, { ReactElement, ReactNode } from "react";
import { MessageBar } from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import { useParams, Link } from "react-router-dom";
import { PortalAPIAppConfig } from "../../types";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

export interface ShowOnlyIfSIWEIsDisabledProps {
  className?: string;
  children?: ReactNode;
}

interface FormState {
  siweChecked: boolean;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const siweIndex = config.authentication?.identities?.indexOf("siwe");
  const siweChecked = siweIndex != null && siweIndex >= 0;
  return {
    siweChecked,
  };
}

function constructConfig(config: PortalAPIAppConfig): PortalAPIAppConfig {
  return config;
}

export default function ShowOnlyIfSIWEIsDisabled(
  props: ShowOnlyIfSIWEIsDisabledProps
): ReactElement {
  const { className, children } = props;
  const { appID } = useParams() as { appID: string };

  const form = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError != null) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (form.state.siweChecked) {
    return (
      <MessageBar className={className}>
        <FormattedMessage
          id="SIWE.disable-first"
          values={{
            reactRouterLink: (chunks: React.ReactNode) => (
              <Link to={`/project/${appID}/configuration/authentication/web3`}>
                {chunks}
              </Link>
            ),
          }}
        />
      </MessageBar>
    );
  }

  return <>{children}</>;
}
