/* global JSX */
import React, { useContext, useMemo, useCallback, useState } from "react";
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
  Dialog,
  DialogFooter,
  PrimaryButton,
  DefaultButton,
  IDialogContentProps,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import produce from "immer";
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
  UserProfileAttributesAccessControl,
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
  rawConfig: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  const modifiedEffectiveConfig = produce(
    effectiveConfig,
    (effectiveConfig) => {
      effectiveConfig.user_profile ??= {};
      effectiveConfig.user_profile.standard_attributes ??= {};
      for (const accessControl of effectiveConfig.user_profile
        .standard_attributes.access_control ?? []) {
        for (const item of currentState.standardAttributesItems) {
          if (accessControl.pointer === item.pointer) {
            accessControl.access_control = item.access_control;
          }
        }
      }
    }
  );

  const accessControl =
    modifiedEffectiveConfig.user_profile?.standard_attributes?.access_control;
  return produce(rawConfig, (rawConfig) => {
    rawConfig.user_profile ??= {};
    rawConfig.user_profile.standard_attributes ??= {};
    rawConfig.user_profile.standard_attributes.access_control = accessControl;
  });
}

function intOfAccessControlLevelString(
  level: AccessControlLevelString
): number {
  switch (level) {
    case "hidden":
      return 1;
    case "readonly":
      return 2;
    case "readwrite":
      return 3;
    default:
      throw new Error("unknown value: " + String(level));
  }
}

function accessControlLevelStringOfInt(
  value: number
): AccessControlLevelString {
  switch (value) {
    case 1:
      return "hidden";
    case 2:
      return "readonly";
    case 3:
      return "readwrite";
  }
  throw new Error("unknown value: " + String(value));
}

type StandardAttributesAccessControlAdjustment = [
  keyof UserProfileAttributesAccessControl,
  AccessControlLevelString
];

function adjustAccessControl(
  accessControl: UserProfileAttributesAccessControl,
  target: keyof UserProfileAttributesAccessControl,
  refValue: AccessControlLevelString
): StandardAttributesAccessControlAdjustment | undefined {
  const targetLevelInt = intOfAccessControlLevelString(accessControl[target]);
  const refLevelInt = intOfAccessControlLevelString(refValue);
  if (targetLevelInt <= refLevelInt) {
    return undefined;
  }

  return [target, accessControlLevelStringOfInt(refLevelInt)];
}

interface PendingUpdate {
  index: number;
  key: keyof UserProfileAttributesAccessControl;
  mainAdjustment: StandardAttributesAccessControlAdjustment;
  otherAdjustments: StandardAttributesAccessControlAdjustment[];
}

function makeUpdate(
  prev: FormState,
  index: number,
  key: keyof UserProfileAttributesAccessControl,
  newValue: AccessControlLevelString
): PendingUpdate {
  const accessControl = prev.standardAttributesItems[index].access_control;
  const mainAdjustment: StandardAttributesAccessControlAdjustment = [
    key,
    newValue,
  ];

  const adjustments: ReturnType<typeof adjustAccessControl>[] = [];
  switch (key) {
    case "end_user":
      break;
    case "bearer": {
      if (newValue === "hidden") {
        adjustments.push(
          adjustAccessControl(accessControl, "end_user", newValue)
        );
      }
      break;
    }
    case "portal_ui": {
      adjustments.push(adjustAccessControl(accessControl, "bearer", newValue));
      adjustments.push(
        adjustAccessControl(accessControl, "end_user", newValue)
      );
      break;
    }
  }

  const otherAdjustments: StandardAttributesAccessControlAdjustment[] =
    adjustments.filter(
      (a): a is StandardAttributesAccessControlAdjustment => a != null
    );

  return {
    index,
    key,
    mainAdjustment,
    otherAdjustments,
  };
}

function applyUpdate(prev: FormState, update: PendingUpdate): FormState {
  const { index, mainAdjustment, otherAdjustments } = update;
  let accessControl = prev.standardAttributesItems[index].access_control;
  const adjustments = [mainAdjustment, ...otherAdjustments];

  for (const adjustment of adjustments) {
    accessControl = {
      ...accessControl,
      [adjustment[0]]: adjustment[1],
    };
  }

  const newItems = [...prev.standardAttributesItems];
  newItems[index] = {
    ...newItems[index],
    access_control: accessControl,
  };

  return {
    ...prev,
    standardAttributesItems: newItems,
  };
}

const UserProfileConfigurationScreenContent: React.FC<UserProfileConfigurationScreenContentProps> =
  function UserProfileConfigurationScreenContent(props) {
    const items = props.form.state.standardAttributesItems;
    const { state, setState } = props.form;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
    const [pendingUpdate, setPendingUpdate] = useState<
      PendingUpdate | undefined
    >();
    const descriptionColor = themes.main.palette.neutralTertiary;

    const onClickConfirmPendingUpdate = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        if (pendingUpdate != null) {
          setState((prev) => applyUpdate(prev, pendingUpdate));
          setPendingUpdate(undefined);
        }
      },
      [setState, pendingUpdate]
    );

    const onDismissPendingUpdateDialog = useCallback(() => {
      setPendingUpdate(undefined);
    }, []);

    // title and subText are typed as string but they can actually be any JSX.Element.
    // @ts-expect-error
    const pendingUpdateDialogContentProps: IDialogContentProps = useMemo(() => {
      if (pendingUpdate == null) {
        return {
          title: "",
          subText: "",
        };
      }

      const { index } = pendingUpdate;

      const pointer = state.standardAttributesItems[index].pointer;
      const fieldName = parseJSONPointer(pointer)[0];

      const formattedLevel = renderToString(
        "standard-attribute.access-control-level." +
          pendingUpdate.mainAdjustment[1]
      );

      const affected =
        pendingUpdate.otherAdjustments.length === 1
          ? pendingUpdate.otherAdjustments[0][0]
          : "other";

      return {
        title: (
          <FormattedMessage
            id="UserProfileConfigurationScreen.dialog.title.pending-update"
            values={{
              fieldName,
              party: pendingUpdate.mainAdjustment[0],
            }}
          />
        ),
        subText: (
          <FormattedMessage
            id="UserProfileConfigurationScreen.dialog.description.pending-update"
            values={{
              fieldName,
              affected,
              party: pendingUpdate.mainAdjustment[0],
              level: formattedLevel,
            }}
          />
        ),
      };
    }, [renderToString, pendingUpdate, state]);

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

    const makeDropdownOnChange = useCallback(
      (index: number, key: keyof UserProfileAttributesAccessControl) => {
        return (
          _e: React.FormEvent<unknown>,
          option?: IDropdownOption<AccessControlLevelString>,
          _index?: number
        ) => {
          if (option == null) {
            return;
          }

          setState((prev) => {
            const pendingUpdate = makeUpdate(
              prev,
              index,
              key,
              option.key as AccessControlLevelString
            );

            if (pendingUpdate.otherAdjustments.length !== 0) {
              setPendingUpdate(pendingUpdate);
              return prev;
            }

            return applyUpdate(prev, pendingUpdate);
          });
        };
      },
      [setState]
    );

    const makeRenderDropdown = useCallback(
      (key: keyof UserProfileAttributesAccessControl) => {
        return (
          item?: StandardAttributesAccessControlConfig,
          index?: number,
          _column?: IColumn
        ) => {
          if (item == null || index == null) {
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

          let options: IDropdownOption<AccessControlLevelString>[] = [];
          let selectedKey: string | undefined;
          switch (key) {
            case "portal_ui":
              options = [optionHidden, optionReadonly, optionReadwrite];
              selectedKey = item.access_control.portal_ui;
              break;
            case "bearer":
              options = [optionHidden, optionReadonly];
              if (item.access_control.portal_ui === "hidden") {
                optionReadonly.disabled = true;
              }
              selectedKey = item.access_control.bearer;
              break;
            case "end_user":
              options = [optionHidden, optionReadonly, optionReadwrite];
              if (item.access_control.bearer === "hidden") {
                optionReadwrite.disabled = true;
                optionReadonly.disabled = true;
              }
              if (item.access_control.portal_ui === "hidden") {
                optionReadwrite.disabled = true;
                optionReadonly.disabled = true;
              }
              if (item.access_control.portal_ui === "readonly") {
                optionReadwrite.disabled = true;
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
              onChange={makeDropdownOnChange(index, key)}
            />
          );
        };
      },
      [renderToString, makeDropdownOnChange]
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
      <>
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
        <Dialog
          hidden={pendingUpdate == null}
          onDismiss={onDismissPendingUpdateDialog}
          dialogContentProps={pendingUpdateDialogContentProps}
        >
          <DialogFooter>
            <PrimaryButton onClick={onClickConfirmPendingUpdate}>
              <FormattedMessage id="confirm" />
            </PrimaryButton>
            <DefaultButton onClick={onDismissPendingUpdateDialog}>
              <FormattedMessage id="cancel" />
            </DefaultButton>
          </DialogFooter>
        </Dialog>
      </>
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
