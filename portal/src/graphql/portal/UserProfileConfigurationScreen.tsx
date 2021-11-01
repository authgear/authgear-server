/* global JSX */
import React, { useContext, useMemo, useCallback } from "react";
import { useParams } from "react-router-dom";
import {
  DetailsList,
  SelectionMode,
  IColumn,
  IRenderFunction,
  IDetailsHeaderProps,
  DetailsHeader,
  IDetailsColumnRenderTooltipProps,
  DirectionalHint,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import LabelWithTooltip from "../../LabelWithTooltip";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";
import styles from "./UserProfileConfigurationScreen.module.scss";

interface FormState {}

interface UserProfileConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {};
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  _currentState: FormState
): PortalAPIAppConfig {
  return config;
}

interface Item {}

const UserProfileConfigurationScreenContent: React.FC<UserProfileConfigurationScreenContentProps> =
  function UserProfileConfigurationScreenContent(_props) {
    const items: Item[] = useMemo(() => [], []);
    const { renderToString } = useContext(Context);

    const columns: IColumn[] = useMemo(
      () => [
        {
          key: "pointer",
          minWidth: 200,
          name: renderToString(
            "UserProfileConfigurationScreen.header.label.attribute-name"
          ),
        },
        {
          key: "portal_ui",
          minWidth: 200,
          maxWidth: 200,
          name: "",
        },
        {
          key: "bearer",
          minWidth: 200,
          maxWidth: 200,
          name: "",
        },
        {
          key: "end_user",
          minWidth: 200,
          maxWidth: 200,
          name: "",
        },
      ],
      [renderToString]
    );

    const onRenderColumnHeaderTooltip: IRenderFunction<IDetailsColumnRenderTooltipProps> =
      useCallback(
        (
          props?: IDetailsColumnRenderTooltipProps,
          defaultRender?: (
            props: IDetailsColumnRenderTooltipProps
          ) => JSX.Element | null
        ) => {
          if (props == null || defaultRender == null) {
            return null;
          }
          if (props.column == null) {
            return null;
          }
          if (
            props.column.key === "portal_ui" ||
            props.column.key === "bearer" ||
            props.column.key === "end_user"
          ) {
            return (
              <LabelWithTooltip
                labelId={
                  "UserProfileConfigurationScreen.header.label." +
                  props.column.key
                }
                tooltipMessageId={
                  "UserProfileConfigurationScreen.header.tooltip." +
                  props.column.key
                }
                directionalHint={DirectionalHint.topCenter}
              />
            );
          }
          return defaultRender(props);
        },
        []
      );

    const onRenderDetailsHeader: IRenderFunction<IDetailsHeaderProps> =
      useCallback(
        (props?: IDetailsHeaderProps) => {
          if (props == null) {
            return null;
          }
          return (
            <DetailsHeader
              {...props}
              onRenderColumnHeaderTooltip={onRenderColumnHeaderTooltip}
            />
          );
        },
        [onRenderColumnHeaderTooltip]
      );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="UserProfileConfigurationScreen.title" />
        </ScreenTitle>
        <div className={styles.widget}>
          <DetailsList
            columns={columns}
            items={items}
            selectionMode={SelectionMode.none}
            onRenderDetailsHeader={onRenderDetailsHeader}
          />
        </div>
      </ScreenContent>
    );
  };

const UserProfileConfigurationScreen: React.FC =
  function UserProfileConfigurationScreen() {
    const { appID } = useParams();
    const form = useAppConfigForm(appID, constructFormState, constructConfig);

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form}>
        <UserProfileConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default UserProfileConfigurationScreen;
