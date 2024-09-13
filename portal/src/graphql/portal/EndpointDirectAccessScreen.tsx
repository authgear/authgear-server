import React, { useCallback, useContext, useMemo, useState } from "react";
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
import { useId } from "../../hook/useId";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { PortalAPIAppConfig } from "../../types";
import styles from "./EndpointDirectAccessScreen.module.css";
import { useParams } from "react-router-dom";
import { useDomainsQuery } from "./query/domainsQuery";
import { nullishCoalesce, or_ } from "../../util/operators";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import Toggle from "../../Toggle";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import { Domain } from "./globalTypes.generated";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ScreenDescription from "../../ScreenDescription";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupStyles,
  ITextFieldStyles,
  MessageBar,
  useTheme,
} from "@fluentui/react";

interface RedirectURLFormState {
  directAccessDisabled: boolean | undefined;
  brandingPageURL: string;
  postLoginURL: string | undefined;
  postLogoutURL: string;
}

function constructRedirectURLFormState(
  config: PortalAPIAppConfig
): RedirectURLFormState {
  return {
    directAccessDisabled: config.ui?.direct_access_disabled,
    brandingPageURL: config.ui?.branding_page_uri ?? "",
    postLoginURL: config.ui?.default_redirect_uri,
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
    draft.ui.direct_access_disabled = currentState.directAccessDisabled;
    draft.ui.branding_page_uri = currentState.brandingPageURL || undefined;
    draft.ui.default_redirect_uri = currentState.postLoginURL;
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
  required?: boolean;
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
      required,
    } = props;
    const id = useId();
    const theme = useTheme();
    const onChange = useCallback(
      (_e: React.FormEvent<any>, value?: string) => {
        onChangeValue(value ?? "");
      },
      [onChangeValue]
    );

    const textFieldStyles: Partial<ITextFieldStyles> = {
      description: {
        color: disabled ? theme.semanticColors.disabledText : "",
      },
    };

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
          required={required}
          styles={textFieldStyles}
          placeholder="https://"
        />
      </div>
    );
  };

enum DirectAccessOptions {
  BrandPage = "BrandPage",
  RedirectURL = "RedirectURL",
}

interface EndpointDirectAccessConfigOptionSelectorProps {
  className?: string;
  redirectURLForm: AppConfigFormModel<RedirectURLFormState>;
  disabled: boolean;
  selectedDirectAccessOption: DirectAccessOptions | undefined;
  onChangeDirectAccessOption: (key: DirectAccessOptions) => void;
  onChangebrandingPageURL: (url: string) => void;
  onChangePostLoginURL: (url: string | undefined) => void;
  onChangePostLogoutURL: (url: string) => void;
}

const EndpointDirectAccessConfigOptionSelector: React.VFC<EndpointDirectAccessConfigOptionSelectorProps> =
  function EndpointDirectAccessConfigOptionSelector(props) {
    const {
      className,
      redirectURLForm,
      disabled,
      selectedDirectAccessOption,
      onChangeDirectAccessOption,
      onChangebrandingPageURL,
      onChangePostLoginURL,
      onChangePostLogoutURL,
    } = props;
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const onOptionsChange = useCallback(
      (
        _?: React.FormEvent<HTMLElement | HTMLInputElement>,
        option?: IChoiceGroupOption
      ) => {
        if (!option?.key) return;
        onChangeDirectAccessOption(option.key as DirectAccessOptions);
      },
      [onChangeDirectAccessOption]
    );

    const onPostRedirectToSettingToggleClick = useCallback(() => {
      if (redirectURLForm.state.postLoginURL != null) {
        onChangePostLoginURL(undefined);
      } else {
        onChangePostLoginURL("");
      }
    }, [onChangePostLoginURL, redirectURLForm.state.postLoginURL]);

    const ChoiceGroupStyles: Partial<IChoiceGroupStyles> = {
      flexContainer: {
        selectors: {
          ".ms-ChoiceField": {
            display: "block",
          },
        },
      },
    };

    const options: IChoiceGroupOption[] = [
      {
        key: DirectAccessOptions.BrandPage,
        text: renderToString(
          "EndpointDirectAccessScreen.configuration.options.brand-page.label"
        ),
        // Workaround for hidden input of ChoiceGroupOption
        // ref: https://github.com/microsoft/fluentui/issues/21252#issuecomment-1168690443
        styles: {
          input: {
            zIndex: -1,
          },
        },
        onRenderField: useCallback(
          (props, render) => {
            return (
              <div>
                {render(props)}
                <div className={cn(styles.optionsChild)}>
                  <FormattedMessage id="EndpointDirectAccessScreen.configuration.options.brand-page.description" />
                </div>
                <RedirectURLTextField
                  className={cn(styles.optionsChild, styles["options--last"])}
                  fieldName="branding_page_uri"
                  label=""
                  description=""
                  value={redirectURLForm.state.brandingPageURL}
                  onChangeValue={onChangebrandingPageURL}
                  disabled={!props?.checked}
                />
                {disabled ? (
                  // https://github.com/microsoft/fluentui/issues/17688
                  <MessageBar className={cn("my-3", "z-10")}>
                    <FormattedMessage
                      id="EndpointDirectAccessScreen.redirect-custom-domain.message"
                      values={{
                        href: `/project/${appID}/branding/custom-domains`,
                      }}
                    />
                  </MessageBar>
                ) : null}
              </div>
            );
          },
          [
            redirectURLForm.state.brandingPageURL,
            onChangebrandingPageURL,
            disabled,
            appID,
          ]
        ),
      },
      {
        key: DirectAccessOptions.RedirectURL,
        text: renderToString(
          "EndpointDirectAccessScreen.configuration.options.redirect-url.label"
        ),
        styles: {
          input: {
            zIndex: -1,
          },
        },
        onRenderField: useCallback(
          (props, render) => {
            return (
              <div>
                {render(props)}
                <div
                  className={cn(
                    styles.optionsChild,
                    styles["options--last"],
                    disabled && styles.disabledText
                  )}
                >
                  <FormattedMessage id="EndpointDirectAccessScreen.configuration.options.redirect-url.description" />
                  <div className={cn("my-4")}>
                    <p className={cn("font-semibold", "my-1.5")}>
                      <FormattedMessage id="EndpointDirectAccessScreen.configuration.options.redirect-to-setting-page.label" />
                    </p>
                    <Toggle
                      inlineLabel={true}
                      disabled={!props?.checked || disabled}
                      checked={redirectURLForm.state.postLoginURL == null}
                      onChange={onPostRedirectToSettingToggleClick}
                      label={renderToString(
                        redirectURLForm.state.postLoginURL == null
                          ? "EndpointDirectAccessScreen.configuration.options.redirect-to-setting-page.on"
                          : "EndpointDirectAccessScreen.configuration.options.redirect-to-setting-page.off"
                      )}
                    />
                  </div>
                  <RedirectURLTextField
                    fieldName="default_redirect_uri"
                    label={renderToString(
                      "EndpointDirectAccessScreen.redirect-url-section.input.post-login-url.label"
                    )}
                    description={renderToString(
                      "EndpointDirectAccessScreen.redirect-url-section.input.post-login-url.description"
                    )}
                    value={redirectURLForm.state.postLoginURL ?? ""}
                    onChangeValue={onChangePostLoginURL}
                    required={
                      !(
                        !props?.checked ||
                        disabled ||
                        redirectURLForm.state.postLoginURL == null
                      )
                    }
                    disabled={
                      !props?.checked ||
                      disabled ||
                      redirectURLForm.state.postLoginURL == null
                    }
                  />
                  <RedirectURLTextField
                    fieldName="default_post_logout_redirect_uri"
                    label={renderToString(
                      "EndpointDirectAccessScreen.redirect-url-section.input.post-logout-url.label"
                    )}
                    description={renderToString(
                      "EndpointDirectAccessScreen.redirect-url-section.input.post-logout-url.description"
                    )}
                    required={true}
                    value={redirectURLForm.state.postLogoutURL}
                    onChangeValue={onChangePostLogoutURL}
                    disabled={!props?.checked || disabled}
                  />
                </div>
              </div>
            );
          },
          [
            disabled,
            redirectURLForm.state.postLoginURL,
            redirectURLForm.state.postLogoutURL,
            onChangePostLoginURL,
            onChangePostLogoutURL,
            onPostRedirectToSettingToggleClick,
            renderToString,
          ]
        ),
        disabled,
      },
    ];

    return (
      <div className={className}>
        <ChoiceGroup
          options={options}
          styles={ChoiceGroupStyles}
          selectedKey={selectedDirectAccessOption}
          onChange={onOptionsChange}
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

    const { canSave, onSubmit } = useFormContainerBaseContext();

    const initSelectedDirectAccessOption = useMemo(() => {
      if (redirectURLForm.initialState.directAccessDisabled) {
        return DirectAccessOptions.BrandPage;
      }
      return DirectAccessOptions.RedirectURL;
    }, [redirectURLForm.initialState.directAccessDisabled]);

    const [selectedDirectAccessOption, setSelectedDirectAccessOption] =
      useState<DirectAccessOptions>(initSelectedDirectAccessOption);

    const onChangebrandingPageURL = useCallback(
      (url: string) => {
        redirectURLForm.setState((prev) =>
          produce(prev, (draft) => {
            draft.directAccessDisabled = true;
            draft.brandingPageURL = url;
          })
        );
      },
      [redirectURLForm]
    );

    const onChangePostLoginURL = useCallback(
      (url: string | undefined) => {
        redirectURLForm.setState((prev) =>
          produce(prev, (draft) => {
            if (url != null) {
              draft.postLoginURL = url;
            } else {
              draft.postLoginURL = undefined;
            }
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

    const onChangeDirectAccessOption = useCallback(
      (directAccessOption: DirectAccessOptions) => {
        setSelectedDirectAccessOption(directAccessOption);
        if (directAccessOption === DirectAccessOptions.BrandPage) {
          redirectURLForm.setState((prev) =>
            produce(prev, (draft) => {
              draft.directAccessDisabled = true;
            })
          );
        } else {
          redirectURLForm.setState((prev) =>
            produce(prev, (draft) => {
              draft.directAccessDisabled = undefined;
            })
          );
        }
      },
      [redirectURLForm]
    );

    return (
      <form className={className} onSubmit={onSubmit}>
        <EndpointDirectAccessConfigOptionSelector
          className={cn(styles.widget, styles.selector)}
          redirectURLForm={redirectURLForm}
          disabled={disabled}
          selectedDirectAccessOption={selectedDirectAccessOption}
          onChangeDirectAccessOption={onChangeDirectAccessOption}
          onChangebrandingPageURL={onChangebrandingPageURL}
          onChangePostLoginURL={onChangePostLoginURL}
          onChangePostLogoutURL={onChangePostLogoutURL}
        />
        <PrimaryButton
          className={cn("mt-12")}
          type="submit"
          disabled={!canSave}
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
          label: <FormattedMessage id="EndpointDirectAccessScreen.title" />,
        },
      ];
    }, []);

    const hasNoCustomDomains = useMemo(() => {
      const index = domains.findIndex((d) => d.isCustom && d.isVerified);
      return index < 0;
    }, [domains]);

    return (
      <ScreenLayoutScrollView>
        <ScreenContent>
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <ScreenDescription className={styles.widget}>
            <FormattedMessage id="EndpointDirectAccessScreen.desc" />
          </ScreenDescription>
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
      refetchDomains().catch((e) => console.error(e));
      redirectURLForm.reload();
    }, [refetchDomains, redirectURLForm]);

    const canSave =
      redirectURLForm.state.directAccessDisabled != null ||
      (redirectURLForm.state.postLogoutURL.length > 0 &&
        (redirectURLForm.state.postLoginURL == null ||
          redirectURLForm.state.postLoginURL.length > 0));

    if (isloading) return <ShowLoading />;
    if (error) return <ShowError error={error} onRetry={retry} />;

    return (
      <FormContainerBase form={redirectURLForm} canSave={canSave}>
        <EndpointDirectAccessContent
          domains={domains ?? []}
          redirectURLForm={redirectURLForm}
        />
      </FormContainerBase>
    );
  };

export default EndpointDirectAccessScreen;
