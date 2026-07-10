import React, { useCallback, useRef } from "react";
import { useParams, useNavigate } from "react-router-dom";
import cn from "classnames";
import { FormattedMessage } from "../../intl";
import { produce } from "immer";
import { Text } from "@radix-ui/themes";
import { PlusIcon } from "@radix-ui/react-icons";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import UserProfileAttributesList, {
  ItemComponentProps,
} from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  CustomAttributesAttributeConfig,
} from "../../types";
import { parseJSONPointer } from "../../util/jsonpointer";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import styles from "./CustomAttributesConfigurationScreen.module.css";

interface FormState {
  items: CustomAttributesAttributeConfig[];
}

interface CustomAttributesConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const items = config.user_profile?.custom_attributes?.attributes ?? [];
  return {
    items,
  };
}

function constructConfig(
  rawConfig: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(rawConfig, (rawConfig) => {
    rawConfig.user_profile ??= {};
    rawConfig.user_profile.custom_attributes ??= {};
    rawConfig.user_profile.custom_attributes.attributes = currentState.items;
  });
}

function EmptyState() {
  const navigate = useNavigate();
  const onClick = useCallback(
    (e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./add");
    },
    [navigate]
  );
  return (
    <div className={styles.emptyState}>
      <div className={styles.emptyStateInner}>
        <div className={styles.emptyStateIcon}>
          <svg
            width="32"
            height="32"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <rect
              x="9"
              y="1"
              width="6"
              height="5"
              rx="1"
              fill="currentColor"
              opacity="0.4"
            />
            <rect
              x="11"
              y="4"
              width="2"
              height="3"
              fill="currentColor"
              opacity="0.4"
            />
            <rect x="2" y="5" width="20" height="16" rx="2" fill="currentColor" opacity="0.15" />
            <rect x="2" y="5" width="20" height="16" rx="2" stroke="currentColor" strokeWidth="1.5" />
            <circle cx="8" cy="12" r="2.5" fill="currentColor" opacity="0.6" />
            <line
              x1="12.5"
              y1="10.5"
              x2="18"
              y2="10.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
            <line
              x1="12.5"
              y1="13.5"
              x2="18"
              y2="13.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
        </div>
        <Text
          as="p"
          size="3"
          weight="bold"
          className={styles.emptyStateHeading}
        >
          <FormattedMessage id="CustomAttributesConfigurationScreen.empty-heading" />
        </Text>
        <Text
          as="p"
          size="2"
          color="gray"
          className={styles.emptyStateDescription}
        >
          <FormattedMessage id="CustomAttributesConfigurationScreen.empty-description" />
        </Text>
      </div>
      <PrimaryButton
        size="2"
        onClick={onClick}
        text={
          <span className={styles.addButtonContent}>
            <PlusIcon width="1rem" height="1rem" />
            <FormattedMessage id="CustomAttributesConfigurationScreen.label.create-custom-attributes" />
          </span>
        }
      />
    </div>
  );
}

function ItemComponent(
  props: ItemComponentProps<CustomAttributesAttributeConfig>
) {
  const { className, item } = props;
  const { pointer } = item;
  const fieldName = parseJSONPointer(pointer)[0];
  return (
    <Text as="p" size="2" weight="medium" className={className}>
      {fieldName}
    </Text>
  );
}

const CustomAttributesConfigurationScreenContent: React.VFC<CustomAttributesConfigurationScreenContentProps> =
  function CustomAttributesConfigurationScreenContent(props) {
    const navigate = useNavigate();
    const { state, setState } = props.form;
    const { items } = state;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const isEmpty = items.length === 0;

    const onAddNewAttribute = useCallback(() => {
      navigate("./add");
    }, [navigate]);

    const onChangeItems = useCallback(
      (newItems: CustomAttributesAttributeConfig[]) => {
        setState((prev) => {
          return {
            ...prev,
            items: newItems,
          };
        });
      },
      [setState]
    );

    const onEditButtonClick = useCallback(
      (index: number) => {
        navigate(`./${index}/edit`);
      },
      [navigate]
    );

    return (
      <ScreenContent
        layout="list"
        className={cn(
          styles.screenContent,
          isDirty ? styles.contentWithSaveBar : null
        )}
      >
        <div ref={contentWidthAnchorRef} className={styles.widget}>
          <div className={styles.header}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="CustomAttributesConfigurationScreen.title" />
            </Text>
            {!isEmpty ? (
              <PrimaryButton
                size="2"
                onClick={onAddNewAttribute}
                text={
                  <span className={styles.addButtonContent}>
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="CustomAttributesConfigurationScreen.label.add-new-attribute" />
                  </span>
                }
              />
            ) : null}
          </div>
        </div>
        <div className={styles.widget}>
          {isEmpty ? (
            <EmptyState />
          ) : (
            <UserProfileAttributesList
              items={items}
              onChangeItems={onChangeItems}
              onReorderItems={onChangeItems}
              onEditButtonClick={onEditButtonClick}
              ItemComponent={ItemComponent}
            />
          )}
        </div>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const CustomAttributesConfigurationScreen: React.VFC =
  function CustomAttributesConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer
        form={form}
        hideFooterComponent={true}
        canSave={true}
      >
        <CustomAttributesConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default CustomAttributesConfigurationScreen;
