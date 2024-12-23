import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import {
  Text,
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  IChoiceGroupStyles,
  ITextFieldStyles,
  useTheme,
} from "@fluentui/react";
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
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import { Domain } from "./globalTypes.generated";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import HorizontalDivider from "../../HorizontalDivider";
import { getHostFromOrigin } from "../../util/domain";

type ChoiceOption =
  | "ShowError"
  | "ShowBrandPage"
  | "ShowLoginAndRedirectToSettings"
  | "ShowLoginAndRedirectToCustomURL";

interface RedirectURLFormState {
  public_origin: string;
  isCustomDomain: boolean;

  currentChoiceOption: ChoiceOption;

  ShowLoginAndRedirectToSettingsDisabled: boolean;
  ShowLoginAndRedirectToCustomURLDisabled: boolean;

  brand_page_uri: string;
  default_redirect_uri: string;
  default_post_logout_redirect_uri: string;
}

function checkIsCustomDomain(
  domains: Domain[],
  public_origin: string
): boolean {
  if (domains.length === 0) {
    return false;
  }
  const index = domains.findIndex((d) =>
    getHostFromOrigin(public_origin).includes(d.domain)
  );
  if (index < 0) {
    return false;
  }
  return domains[index].isCustom;
}

function makeConstructRedirectURLFormState(domains: Domain[]) {
  return (config: PortalAPIAppConfig) => {
    const public_origin = config.http?.public_origin ?? "";

    const isCustomDomain = checkIsCustomDomain(domains, public_origin);

    const direct_access_disabled = config.ui?.direct_access_disabled ?? false;
    const brand_page_uri = config.ui?.brand_page_uri ?? "";
    const default_redirect_uri = config.ui?.default_redirect_uri ?? "";
    const default_post_logout_redirect_uri =
      config.ui?.default_post_logout_redirect_uri ?? "";

    const currentChoiceOption: ChoiceOption =
      !isCustomDomain || direct_access_disabled
        ? brand_page_uri === ""
          ? "ShowError"
          : "ShowBrandPage"
        : default_redirect_uri === ""
        ? "ShowLoginAndRedirectToSettings"
        : "ShowLoginAndRedirectToCustomURL";

    return {
      public_origin,
      isCustomDomain,

      currentChoiceOption,

      ShowLoginAndRedirectToSettingsDisabled: !isCustomDomain,
      ShowLoginAndRedirectToCustomURLDisabled: !isCustomDomain,

      brand_page_uri,
      default_redirect_uri,
      default_post_logout_redirect_uri,
    };
  };
}

function constructConfigFromRedirectURLFormState(
  config: PortalAPIAppConfig,
  _initialState: RedirectURLFormState,
  currentState: RedirectURLFormState
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    draft.ui ??= {};

    switch (currentState.currentChoiceOption) {
      case "ShowError":
        draft.ui.direct_access_disabled = true;
        draft.ui.brand_page_uri = undefined;
        draft.ui.default_redirect_uri = undefined;
        break;
      case "ShowBrandPage":
        draft.ui.direct_access_disabled = true;
        draft.ui.brand_page_uri = currentState.brand_page_uri;
        draft.ui.default_redirect_uri = undefined;
        break;
      case "ShowLoginAndRedirectToSettings":
        draft.ui.direct_access_disabled = undefined;
        draft.ui.brand_page_uri = undefined;
        draft.ui.default_redirect_uri = undefined;
        break;
      case "ShowLoginAndRedirectToCustomURL":
        draft.ui.direct_access_disabled = undefined;
        draft.ui.brand_page_uri = undefined;
        draft.ui.default_redirect_uri = currentState.default_redirect_uri;
        break;
    }

    if (currentState.default_post_logout_redirect_uri === "") {
      draft.ui.default_post_logout_redirect_uri = undefined;
    } else {
      draft.ui.default_post_logout_redirect_uri =
        currentState.default_post_logout_redirect_uri;
    }
  });
}

interface RedirectURLTextFieldProps {
  fieldName: string;
  value: string;
  onChangeValue: (value: string) => void;

  className?: string;
  label?: React.ReactNode;
  description?: React.ReactNode;
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
      <FormTextField
        id={id}
        fieldName={fieldName}
        parentJSONPointer="/ui"
        className={className}
        /* @ts-expect-error label can be React.ReactNode */
        label={label}
        /* @ts-expect-error description can be React.ReactNode */
        description={description}
        value={value}
        onChange={onChange}
        disabled={disabled}
        required={required}
        styles={textFieldStyles}
        placeholder="https://"
      />
    );
  };

// Workaround for hidden input of ChoiceGroupOption
// ref: https://github.com/microsoft/fluentui/issues/21252#issuecomment-1168690443
const WORKAROUND_HIDDEN_INPUT_OF_ChoiceGroupOption = {
  input: {
    zIndex: -1,
  },
};

interface EndpointDirectAccessConfigOptionSelectorProps {
  className?: string;
  form: AppConfigFormModel<RedirectURLFormState>;
  onChangeDirectAccessOption: (key: ChoiceOption) => void;
  onChangeBrandPageURL: (url: string) => void;
  onChangePostLoginURL: (url: string) => void;
  onChangePostLogoutURL: (url: string) => void;
}

const EndpointDirectAccessConfigOptionSelector: React.VFC<EndpointDirectAccessConfigOptionSelectorProps> =
  function EndpointDirectAccessConfigOptionSelector(props) {
    const {
      className,
      form,
      onChangeDirectAccessOption,
      onChangeBrandPageURL,
      onChangePostLoginURL,
    } = props;
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const {
      ShowLoginAndRedirectToSettingsDisabled,
      ShowLoginAndRedirectToCustomURLDisabled,
    } = form.state;

    const onOptionsChange = useCallback(
      (
        _?: React.FormEvent<HTMLElement | HTMLInputElement>,
        option?: IChoiceGroupOption
      ) => {
        if (!option?.key) return;
        onChangeDirectAccessOption(option.key as ChoiceOption);
      },
      [onChangeDirectAccessOption]
    );

    const ChoiceGroupStyles: Partial<IChoiceGroupStyles> = {
      flexContainer: {
        selectors: {
          ".ms-ChoiceField": {
            display: "block",
          },
        },
      },
    };

    const onRenderFieldShowBrandPage = useCallback(
      (
        props?: IChoiceGroupOption & IChoiceGroupOptionProps,
        render?: (
          props?: IChoiceGroupOption & IChoiceGroupOptionProps
        ) => JSX.Element | null
      ) => {
        const checked = props?.checked ?? false;
        return (
          <>
            {render?.(props)}
            {checked ? (
              <RedirectURLTextField
                className={cn(styles.textFieldInOption)}
                fieldName="brand_page_uri"
                label={
                  <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowBrandPage.input.label" />
                }
                value={form.state.brand_page_uri}
                onChangeValue={onChangeBrandPageURL}
              />
            ) : null}
          </>
        );
      },
      [form.state.brand_page_uri, onChangeBrandPageURL]
    );

    const onRenderFieldShowLoginAndRedirectToCustomURL = useCallback(
      (
        props?: IChoiceGroupOption & IChoiceGroupOptionProps,
        render?: (
          props?: IChoiceGroupOption & IChoiceGroupOptionProps
        ) => JSX.Element | null
      ) => {
        const checked = props?.checked ?? false;
        return (
          <>
            {render?.(props)}
            {checked ? (
              <RedirectURLTextField
                className={cn(styles.textFieldInOption)}
                fieldName="default_redirect_uri"
                label={
                  <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.input.label" />
                }
                description={
                  <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.input.description" />
                }
                value={form.state.default_redirect_uri}
                onChangeValue={onChangePostLoginURL}
              />
            ) : null}
          </>
        );
      },
      [form.state.default_redirect_uri, onChangePostLoginURL]
    );

    const options: IChoiceGroupOption[] = [
      {
        key: "ShowError",
        text: renderToString(
          "EndpointDirectAccessScreen.section1.option.ShowError.label"
        ),
      },
      {
        key: "ShowBrandPage",
        text: renderToString(
          "EndpointDirectAccessScreen.section1.option.ShowBrandPage.label"
        ),
        styles: WORKAROUND_HIDDEN_INPUT_OF_ChoiceGroupOption,
        onRenderField: onRenderFieldShowBrandPage,
      },
      {
        key: "ShowLoginAndRedirectToSettings",
        disabled: ShowLoginAndRedirectToSettingsDisabled,
        // @ts-expect-error text can be React.Element.
        text: ShowLoginAndRedirectToSettingsDisabled ? (
          <FormattedMessage
            id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToSettings.label--disabled"
            values={{
              href: `/project/${appID}/branding/custom-domains`,
            }}
          />
        ) : (
          <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToSettings.label" />
        ),
      },
      {
        key: "ShowLoginAndRedirectToCustomURL",
        disabled: ShowLoginAndRedirectToCustomURLDisabled,
        // @ts-expect-error text can be React.Element.
        text: ShowLoginAndRedirectToSettingsDisabled ? (
          <FormattedMessage
            id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.label--disabled"
            values={{
              href: `/project/${appID}/branding/custom-domains`,
            }}
          />
        ) : (
          <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.label" />
        ),
        onRenderField: onRenderFieldShowLoginAndRedirectToCustomURL,
      },
    ];

    return (
      <ChoiceGroup
        className={className}
        options={options}
        styles={ChoiceGroupStyles}
        selectedKey={form.state.currentChoiceOption}
        onChange={onOptionsChange}
      />
    );
  };

interface RedirectURLFormProps {
  form: AppConfigFormModel<RedirectURLFormState>;
}
const RedirectURLForm: React.VFC<RedirectURLFormProps> =
  function RedirectURLForm(props) {
    const { form } = props;
    const { public_origin: publicOrigin } = form.state;

    const { canSave, onSubmit } = useFormContainerBaseContext();

    const onChangeBrandPageURL = useCallback(
      (url: string) => {
        form.setState((prev) =>
          produce(prev, (draft) => {
            draft.brand_page_uri = url;
          })
        );
      },
      [form]
    );

    const onChangePostLoginURL = useCallback(
      (url: string) => {
        form.setState((prev) =>
          produce(prev, (draft) => {
            draft.default_redirect_uri = url;
          })
        );
      },
      [form]
    );

    const onChangePostLogoutURL = useCallback(
      (url: string) => {
        form.setState((prev) =>
          produce(prev, (draft) => {
            draft.default_post_logout_redirect_uri = url;
          })
        );
      },
      [form]
    );

    const onChangeDirectAccessOption = useCallback(
      (directAccessOption: ChoiceOption) => {
        form.setState((prev) =>
          produce(prev, (draft) => {
            draft.currentChoiceOption = directAccessOption;
          })
        );
      },
      [form]
    );

    return (
      <form
        className={cn(styles.widget, "flex flex-col gap-12")}
        onSubmit={onSubmit}
      >
        <div /* this div exists for gap to work, do not remove */>
          <div className="flex flex-col gap-6">
            <Text as="h2">
              <FormattedMessage
                id="EndpointDirectAccessScreen.section1.description"
                values={{
                  endpoint: publicOrigin,
                }}
              />
            </Text>
            <EndpointDirectAccessConfigOptionSelector
              form={form}
              onChangeDirectAccessOption={onChangeDirectAccessOption}
              onChangeBrandPageURL={onChangeBrandPageURL}
              onChangePostLoginURL={onChangePostLoginURL}
              onChangePostLogoutURL={onChangePostLogoutURL}
            />
          </div>

          <HorizontalDivider className="my-12" />

          <div className="flex flex-col gap-4">
            <Text as="h2" className="block">
              <FormattedMessage
                id="EndpointDirectAccessScreen.section2.description"
                values={{
                  endpoint: publicOrigin,
                }}
              />
            </Text>

            <RedirectURLTextField
              fieldName="default_post_logout_redirect_uri"
              value={form.state.default_post_logout_redirect_uri}
              label={
                <FormattedMessage id="EndpointDirectAccessScreen.section2.input.label" />
              }
              onChangeValue={onChangePostLogoutURL}
            />
          </div>
        </div>

        <PrimaryButton
          className="self-start"
          type="submit"
          disabled={!canSave}
          text={<FormattedMessage id="save" />}
        />
      </form>
    );
  };

interface EndpointDirectAccessContentProps {
  form: AppConfigFormModel<RedirectURLFormState>;
}

const EndpointDirectAccessContent: React.VFC<EndpointDirectAccessContentProps> =
  function EndpointDirectAccessContent(props) {
    const { form } = props;

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: ".",
          label: <FormattedMessage id="EndpointDirectAccessScreen.title" />,
        },
      ];
    }, []);

    return (
      <ScreenLayoutScrollView>
        <ScreenContent>
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <RedirectURLForm form={form} />
        </ScreenContent>
      </ScreenLayoutScrollView>
    );
  };

interface EndpointDirectAccessScreen1Props {
  appID: string;
  domains: Domain[];
}

function EndpointDirectAccessScreen1(props: EndpointDirectAccessScreen1Props) {
  const { appID, domains } = props;
  const constructFormState = useMemo(
    () => makeConstructRedirectURLFormState(domains),
    [domains]
  );

  const redirectURLForm = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig: constructConfigFromRedirectURLFormState,
  });

  if (redirectURLForm.isLoading) {
    return <ShowLoading />;
  }
  if (redirectURLForm.loadError) {
    return (
      <ShowError
        error={redirectURLForm.loadError}
        onRetry={redirectURLForm.reload}
      />
    );
  }

  return (
    <FormContainerBase form={redirectURLForm}>
      <EndpointDirectAccessContent form={redirectURLForm} />
    </FormContainerBase>
  );
}

const EndpointDirectAccessScreen: React.VFC =
  function EndpointDirectAccessScreen() {
    const { appID } = useParams() as { appID: string };
    const { domains, loading, error, refetch } = useDomainsQuery(appID);

    if (loading) {
      return <ShowLoading />;
    }
    if (error) {
      return <ShowError error={error} onRetry={refetch} />;
    }

    return (
      <EndpointDirectAccessScreen1 appID={appID} domains={domains ?? []} />
    );
  };

export default EndpointDirectAccessScreen;
