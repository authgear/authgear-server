import React, { useCallback, useMemo, useRef } from "react";
import cn from "classnames";
import { Flex, RadioGroup, Text } from "@radix-ui/themes";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { FormattedMessage } from "../../intl";
import { produce } from "immer";
import FormContainer from "../../FormContainer";
import { PortalAPIAppConfig } from "../../types";
import styles from "./EndpointDirectAccessScreen.module.css";
import { useParams } from "react-router-dom";
import { useDomainsQuery } from "./query/domainsQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import { Domain } from "./globalTypes.generated";
import { getHostFromOrigin } from "../../util/domain";
import { TextField } from "../../components/v2/TextField/TextField";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import Link from "../../Link";

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

interface EndpointDirectAccessConfigOptionSelectorProps {
  form: AppConfigFormModel<RedirectURLFormState>;
  onChangeDirectAccessOption: (key: ChoiceOption) => void;
  onChangeBrandPageURL: (url: string) => void;
  onChangePostLoginURL: (url: string) => void;
}

const EndpointDirectAccessConfigOptionSelector: React.VFC<EndpointDirectAccessConfigOptionSelectorProps> =
  function EndpointDirectAccessConfigOptionSelector(props) {
    const { form, onChangeDirectAccessOption, onChangeBrandPageURL, onChangePostLoginURL } =
      props;
    const { appID } = useParams() as { appID: string };

    const {
      currentChoiceOption,
      ShowLoginAndRedirectToSettingsDisabled,
      ShowLoginAndRedirectToCustomURLDisabled,
    } = form.state;

    const onOptionsChange = useCallback(
      (value: string) => {
        onChangeDirectAccessOption(value as ChoiceOption);
      },
      [onChangeDirectAccessOption]
    );

    const onChangeBrandPage = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        onChangeBrandPageURL(e.target.value);
      },
      [onChangeBrandPageURL]
    );

    const onChangePostLogin = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        onChangePostLoginURL(e.target.value);
      },
      [onChangePostLoginURL]
    );

    const customDomainsLink = useCallback(
      (chunks: React.ReactNode) => (
        <Link to={`/project/${appID}/branding/custom-domains`}>{chunks}</Link>
      ),
      [appID]
    );

    return (
      <RadioGroup.Root
        value={currentChoiceOption}
        onValueChange={onOptionsChange}
      >
        <Flex direction="column" gap="3">
          <Text as="label" size="2">
            <Flex gap="2" align="start">
              <RadioGroup.Item value="ShowError" />
              <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowError.label" />
            </Flex>
          </Text>

          <div className="flex flex-col gap-2">
            <Text as="label" size="2">
              <Flex gap="2" align="start">
                <RadioGroup.Item value="ShowBrandPage" />
                <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowBrandPage.label" />
              </Flex>
            </Text>
            {currentChoiceOption === "ShowBrandPage" ? (
              <div className={styles.textFieldInOption}>
                <TextField
                  size="2"
                  labelSize="2"
                  type="text"
                  placeholder="https://"
                  label={
                    <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowBrandPage.input.label" />
                  }
                  value={form.state.brand_page_uri}
                  onChange={onChangeBrandPage}
                  parentJSONPointer="/ui"
                  fieldName="brand_page_uri"
                />
              </div>
            ) : null}
          </div>

          <div className="flex flex-col gap-2">
            <Flex gap="2" align="start">
              <RadioGroup.Item
                value="ShowLoginAndRedirectToSettings"
                disabled={ShowLoginAndRedirectToSettingsDisabled}
                className="shrink-0 mt-0.5"
              />
              <div className="flex flex-col gap-1 min-w-0">
                <Text
                  as="span"
                  size="2"
                  color={ShowLoginAndRedirectToSettingsDisabled ? "gray" : undefined}
                >
                  <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToSettings.label" />
                </Text>
                {ShowLoginAndRedirectToSettingsDisabled ? (
                  <Text as="p" size="1" color="gray">
                    <FormattedMessage
                      id="EndpointDirectAccessScreen.section1.option.requires-custom-domain.hint"
                      values={{
                        // eslint-disable-next-line react/no-unstable-nested-components
                        reactRouterLink: customDomainsLink,
                      }}
                    />
                  </Text>
                ) : null}
              </div>
            </Flex>
          </div>

          <div className="flex flex-col gap-2">
            <Flex gap="2" align="start">
              <RadioGroup.Item
                value="ShowLoginAndRedirectToCustomURL"
                disabled={ShowLoginAndRedirectToCustomURLDisabled}
                className="shrink-0 mt-0.5"
              />
              <div className="flex flex-col gap-1 min-w-0">
                <Text
                  as="span"
                  size="2"
                  color={ShowLoginAndRedirectToCustomURLDisabled ? "gray" : undefined}
                >
                  <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.label" />
                </Text>
                {ShowLoginAndRedirectToCustomURLDisabled ? (
                  <Text as="p" size="1" color="gray">
                    <FormattedMessage
                      id="EndpointDirectAccessScreen.section1.option.requires-custom-domain.hint"
                      values={{
                        // eslint-disable-next-line react/no-unstable-nested-components
                        reactRouterLink: customDomainsLink,
                      }}
                    />
                  </Text>
                ) : null}
              </div>
            </Flex>
            {currentChoiceOption === "ShowLoginAndRedirectToCustomURL" ? (
              <div className={styles.textFieldInOption}>
                <TextField
                  size="2"
                  labelSize="2"
                  type="text"
                  placeholder="https://"
                  label={
                    <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.input.label" />
                  }
                  hint={
                    <FormattedMessage id="EndpointDirectAccessScreen.section1.option.ShowLoginAndRedirectToCustomURL.input.description" />
                  }
                  value={form.state.default_redirect_uri}
                  onChange={onChangePostLogin}
                  parentJSONPointer="/ui"
                  fieldName="default_redirect_uri"
                />
              </div>
            ) : null}
          </div>
        </Flex>
      </RadioGroup.Root>
    );
  };

interface EndpointDirectAccessContentProps {
  form: AppConfigFormModel<RedirectURLFormState>;
}

const EndpointDirectAccessContent: React.VFC<EndpointDirectAccessContentProps> =
  function EndpointDirectAccessContent(props) {
    const { form } = props;
    const { public_origin: publicOrigin } = form.state;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

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
      (e: React.ChangeEvent<HTMLInputElement>) => {
        form.setState((prev) =>
          produce(prev, (draft) => {
            draft.default_post_logout_redirect_uri = e.target.value;
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
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={styles.contentWidthAnchor}
          aria-hidden
        />
        <div className={cn(styles.widget, styles.pageHeader)}>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="EndpointDirectAccessScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="EndpointDirectAccessScreen.description" />
          </Text>
        </div>

        <div
          className={cn(
            styles.widget,
            "border border-[var(--gray-5)] rounded-lg p-6 flex gap-8 bg-white",
            isDirty && styles.settingsCardSaveBarClearance
          )}
        >
          <Text as="p" size="3" weight="medium" className="shrink-0 w-[200px]">
            <FormattedMessage id="EndpointDirectAccessScreen.settings.label" />
          </Text>
          <div className="flex-1 flex flex-col gap-6 min-w-0">
            <div className="flex flex-col gap-4">
              <Text as="p" size="2" color="gray">
                <FormattedMessage
                  id="EndpointDirectAccessScreen.section1.description"
                  values={{
                    endpoint: publicOrigin,
                    // eslint-disable-next-line react/no-unstable-nested-components
                    strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
                  }}
                />
              </Text>
              <EndpointDirectAccessConfigOptionSelector
                form={form}
                onChangeDirectAccessOption={onChangeDirectAccessOption}
                onChangeBrandPageURL={onChangeBrandPageURL}
                onChangePostLoginURL={onChangePostLoginURL}
              />
            </div>

            <hr className="border-0 border-t border-[var(--gray-5)]" />

            <div className="flex flex-col gap-4">
              <Text as="p" size="2" color="gray">
                <FormattedMessage
                  id="EndpointDirectAccessScreen.section2.description"
                  values={{
                    endpoint: publicOrigin,
                    // eslint-disable-next-line react/no-unstable-nested-components
                    strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
                  }}
                />
              </Text>
              <TextField
                size="2"
                labelSize="2"
                type="text"
                placeholder="https://"
                label={
                  <FormattedMessage id="EndpointDirectAccessScreen.section2.input.label" />
                }
                value={form.state.default_post_logout_redirect_uri}
                onChange={onChangePostLogoutURL}
                parentJSONPointer="/ui"
                fieldName="default_post_logout_redirect_uri"
              />
            </div>
          </div>
        </div>

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
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
    <FormContainer form={redirectURLForm} hideFooterComponent={true}>
      <EndpointDirectAccessContent form={redirectURLForm} />
    </FormContainer>
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
