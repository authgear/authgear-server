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
  Text,
  Dropdown,
  IDropdownOption,
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
import {
  PortalAPIAppConfig,
  StandardAttributesAccessControlConfig,
  StandardAttributesAccessControl,
  AccessControlLevelString,
} from "../../types";
import { parseJSONPointer } from "../../util/jsonpointer";
import styles from "./UserProfileConfigurationScreen.module.scss";
import { useSystemConfig } from "../../context/SystemConfigContext";

interface FormState {
  standardAttributesItems: StandardAttributesAccessControlConfig[];
}

interface UserProfileConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const naturalOrder = [
  "/name",
  "/given_name",
  "/family_name",
  "/middle_name",
  "/nickname",
  "/profile",
  "/picture",
  "/website",
  "/gender",
  "/birthdate",
  "/zoneinfo",
  "/locale",
  "/address",
];

function constructFormState(config: PortalAPIAppConfig): FormState {
  const items = config.user_profile?.standard_attributes?.access_control ?? [];
  const listedItems = items.filter((a) => naturalOrder.indexOf(a.pointer) >= 0);
  listedItems.sort((a, b) => {
    const ia = naturalOrder.indexOf(a.pointer);
    const ib = naturalOrder.indexOf(b.pointer);
    return ia - ib;
  });
  return {
    standardAttributesItems: listedItems,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  _currentState: FormState
): PortalAPIAppConfig {
  return config;
}

const UserProfileConfigurationScreenContent: React.FC<UserProfileConfigurationScreenContentProps> =
  function UserProfileConfigurationScreenContent(props) {
    const items = props.form.state.standardAttributesItems;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
    const descriptionColor = themes.main.palette.neutralTertiary;

    const onRenderPointer = useCallback(
      (
        item?: StandardAttributesAccessControlConfig,
        _index?: number,
        _column?: IColumn
      ) => {
        if (item == null) {
          return null;
        }
        const { pointer } = item;
        const fieldName = parseJSONPointer(pointer)[0];
        return (
          <div>
            <Text className={styles.fieldName} block={true}>
              <FormattedMessage id={"standard-attribute." + fieldName} />
            </Text>
            <Text
              variant="small"
              block={true}
              style={{
                color: descriptionColor,
              }}
            >
              <FormattedMessage
                id={"standard-attribute.description." + fieldName}
              />
            </Text>
          </div>
        );
      },
      [descriptionColor]
    );

    const makeRenderDropdown = useCallback(
      (key: keyof StandardAttributesAccessControl) => {
        return (
          item?: StandardAttributesAccessControlConfig,
          _index?: number,
          _column?: IColumn
        ) => {
          if (item == null) {
            return null;
          }

          const optionHidden: IDropdownOption = {
            key: "hidden",
            text: renderToString(
              "standard-attribute.access-control-level.hidden"
            ),
          };

          const optionReadonly: IDropdownOption = {
            key: "readonly",
            text: renderToString(
              "standard-attribute.access-control-level.readonly"
            ),
          };

          const optionReadwrite: IDropdownOption = {
            key: "readwrite",
            text: renderToString(
              "standard-attribute.access-control-level.readwrite"
            ),
          };

          const options: IDropdownOption<AccessControlLevelString>[] = [
            optionHidden,
            optionReadonly,
            optionReadwrite,
          ];

          let selectedKey: string | undefined;
          switch (key) {
            case "portal_ui":
              selectedKey = item.access_control.portal_ui;
              break;
            case "bearer":
              if (item.access_control.portal_ui === "readonly") {
                optionReadwrite.disabled = true;
              }
              if (item.access_control.portal_ui === "hidden") {
                optionReadwrite.disabled = true;
                optionReadonly.disabled = true;
              }
              selectedKey = item.access_control.bearer;
              break;
            case "end_user":
              if (item.access_control.bearer === "readonly") {
                optionReadwrite.disabled = true;
              }
              if (item.access_control.bearer === "hidden") {
                optionReadwrite.disabled = true;
                optionReadonly.disabled = true;
              }
              selectedKey = item.access_control.end_user;
              break;
          }

          const disabledOptionCount = options.reduce((a, b) => {
            return a + (b.disabled === true ? 1 : 0);
          }, 0);
          const dropdownIsDisabled = options.length - disabledOptionCount <= 1;

          return (
            <Dropdown
              options={options}
              selectedKey={selectedKey}
              disabled={dropdownIsDisabled}
            />
          );
        };
      },
      [renderToString]
    );

    const columns: IColumn[] = useMemo(
      () => [
        {
          key: "pointer",
          minWidth: 200,
          name: renderToString(
            "UserProfileConfigurationScreen.header.label.attribute-name"
          ),
          onRender: onRenderPointer,
          isMultiline: true,
        },
        {
          key: "portal_ui",
          minWidth: 200,
          maxWidth: 200,
          name: "",
          onRender: makeRenderDropdown("portal_ui"),
        },
        {
          key: "bearer",
          minWidth: 200,
          maxWidth: 200,
          name: "",
          onRender: makeRenderDropdown("bearer"),
        },
        {
          key: "end_user",
          minWidth: 200,
          maxWidth: 200,
          name: "",
          onRender: makeRenderDropdown("end_user"),
        },
      ],
      [renderToString, makeRenderDropdown, onRenderPointer]
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
