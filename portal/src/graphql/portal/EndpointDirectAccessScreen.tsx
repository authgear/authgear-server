import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
import { produce } from "immer";
import WidgetTitle from "../../WidgetTitle";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { useId } from "../../hook/useId";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { PortalAPIAppConfig } from "../../types";
import styles from "./CustomDomainListScreen.module.css";
import { useParams } from "react-router-dom";
import { useDomainsQuery } from "./query/domainsQuery";
import { nullishCoalesce, or_ } from "../../util/operators";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import { Domain } from "./globalTypes.generated";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

interface RedirectURLFormState {
  postLoginURL: string;
  postLogoutURL: string;
}

function constructRedirectURLFormState(
  config: PortalAPIAppConfig
): RedirectURLFormState {
  return {
    postLoginURL: config.ui?.default_redirect_uri ?? "",
    postLogoutURL: config.ui?.default_post_logout_redirect_uri ?? "",
  };
}
function constructConfigFromRedirectURLFormState(
  config: PortalAPIAppConfig,
  _initialState: RedirectURLFormState,
  currentState: RedirectURLFormState
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    draft.ui ??= {};
    draft.ui.default_redirect_uri = currentState.postLoginURL || undefined;
    draft.ui.default_post_logout_redirect_uri =
      currentState.postLogoutURL || undefined;
  });
}

interface RedirectURLTextFieldProps {
  className?: string;
  fieldName: string;
  label: string;
  description: string;
  value: string;
  onChangeValue: (value: string) => void;
  disabled?: boolean;
}
const RedirectURLTextField: React.VFC<RedirectURLTextFieldProps> =
  function RedirectURLTextField(props) {
    const {
      fieldName,
      className,
      label,
      description,
      value,
      onChangeValue,
      disabled,
    } = props;
    const id = useId();
    const onChange = useCallback(
      (_e: React.FormEvent<any>, value?: string) => {
        onChangeValue(value ?? "");
      },
      [onChangeValue]
    );
    return (
      <div className={className}>
        <FormTextField
          id={id}
          fieldName={fieldName}
          parentJSONPointer="/ui"
          className={cn("mt-2.5")}
          label={label}
          description={description}
          value={value}
          onChange={onChange}
          disabled={disabled}
        />
      </div>
    );
  };

interface RedirectURLFormProps {
  className?: string;
  redirectURLForm: AppConfigFormModel<RedirectURLFormState>;
  disabled: boolean;
}
const RedirectURLForm: React.VFC<RedirectURLFormProps> =
  function RedirectURLForm(props) {
    const { className, redirectURLForm, disabled } = props;
    const { renderToString } = useContext(Context);

    const { canSave, onSubmit } = useFormContainerBaseContext();

    const onChangePostLoginURL = useCallback(
      (url: string) => {
        redirectURLForm.setState((prev) =>
          produce(prev, (draft) => {
            draft.postLoginURL = url;
          })
        );
      },
      [redirectURLForm]
    );

    const onChangePostLogoutURL = useCallback(
      (url: string) => {
        redirectURLForm.setState((prev) =>
          produce(prev, (draft) => {
            draft.postLogoutURL = url;
          })
        );
      },
      [redirectURLForm]
    );

    return (
      <form className={className} onSubmit={onSubmit}>
        <WidgetTitle>
          <FormattedMessage id="CustomDomainListScreen.redirectURLSection.title" />
        </WidgetTitle>
        {disabled ? (
          <FeatureDisabledMessageBar
            className="mt-4"
            messageID="CustomDomainListScreen.redirectURLSection.disabled.message"
          />
        ) : null}
        <RedirectURLTextField
          className={cn("mt-4")}
          fieldName="default_redirect_uri"
          label={renderToString(
            "CustomDomainListScreen.redirectURLSection.input.postLoginURL.label"
          )}
          description={renderToString(
            "CustomDomainListScreen.redirectURLSection.input.postLoginURL.description"
          )}
          value={redirectURLForm.state.postLoginURL}
          onChangeValue={onChangePostLoginURL}
          disabled={disabled}
        />
        <RedirectURLTextField
          className={cn("mt-4")}
          fieldName="default_post_logout_redirect_uri"
          label={renderToString(
            "CustomDomainListScreen.redirectURLSection.input.postLogoutURL.label"
          )}
          description={renderToString(
            "CustomDomainListScreen.redirectURLSection.input.postLogoutURL.description"
          )}
          value={redirectURLForm.state.postLogoutURL}
          onChangeValue={onChangePostLogoutURL}
          disabled={disabled}
        />
        <PrimaryButton
          className={cn("mt-12")}
          type="submit"
          disabled={!canSave || disabled}
          text={<FormattedMessage id="save" />}
        ></PrimaryButton>
      </form>
    );
  };

interface EndpointDirectAccessContentProps {
  domains: Domain[];
  redirectURLForm: AppConfigFormModel<RedirectURLFormState>;
}

const EndpointDirectAccessContent: React.VFC<EndpointDirectAccessContentProps> =
  function EndpointDirectAccessContent(props) {
    const { domains, redirectURLForm } = props;

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: ".",
          label: <FormattedMessage id="CustomDomainListScreen.title" />,
        },
      ];
    }, []);

    const hasNoCustomDomains = useMemo(() => {
      if (!domains) return true;
      const index = domains.findIndex((d) => d.isCustom === true);
      return index < 0;
    }, [domains]);

    return (
      <ScreenLayoutScrollView>
        <ScreenContent>
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <RedirectURLForm
            className={cn(styles.widget)}
            redirectURLForm={redirectURLForm}
            disabled={hasNoCustomDomains}
          />
        </ScreenContent>
      </ScreenLayoutScrollView>
    );
  };

const EndpointDirectAccessScreen: React.VFC =
  function EndpointDirectAccessScreen() {
    const { appID } = useParams() as { appID: string };
    const {
      domains,
      loading: fetchingDomains,
      error: fetchDomainsError,
      refetch: refetchDomains,
    } = useDomainsQuery(appID);
    const redirectURLForm = useAppConfigForm({
      appID,
      constructFormState: constructRedirectURLFormState,
      constructConfig: constructConfigFromRedirectURLFormState,
    });

    const isloading = or_(fetchingDomains, redirectURLForm.isLoading);
    const error = nullishCoalesce(fetchDomainsError, redirectURLForm.loadError);
    const retry = useCallback(() => {
      refetchDomains().catch((e) => console.log(e));
      redirectURLForm.reload();
    }, [refetchDomains, redirectURLForm]);

    if (isloading) return <ShowLoading />;
    if (error) return <ShowError error={error} onRetry={retry} />;

    return (
      <FormContainerBase form={redirectURLForm}>
        <EndpointDirectAccessContent
          domains={domains ?? []}
          redirectURLForm={redirectURLForm}
        />
      </FormContainerBase>
    );
  };

export default EndpointDirectAccessScreen;
