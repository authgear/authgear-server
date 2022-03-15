import React, { useMemo, useContext } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import CommandBarContainer from "../../CommandBarContainer";
import { FormProvider } from "../../form";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ScreenContent from "../../ScreenContent";
import { useSystemConfig } from "../../context/SystemConfigContext";

import styles from "./EditPictureScreen.module.scss";
import { ICommandBarItemProps } from "@fluentui/react";

const EditPictureScreen: React.FC = function EditPictureScreen() {
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const isDirty = false;
  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="EditPictureScreen.title" /> },
    ];
  }, []);
  const initialItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "upload",
        text: renderToString("EditPictureScreen.upload-new-picture.label"),
        iconProps: { iconName: "Upload" },
      },
      {
        key: "remove",
        text: renderToString("EditPictureScreen.remove-picture.label"),
        iconProps: { iconName: "Delete" },
        theme: themes.destructive,
      },
    ];
  }, [renderToString, themes.destructive]);
  return (
    <FormProvider>
      <CommandBarContainer
        primaryItems={initialItems}
        messageBar={<FormErrorMessageBar />}
      >
        <form>
          <ScreenContent>
            <NavBreadcrumb
              className={styles.widget}
              items={navBreadcrumbItems}
            />
          </ScreenContent>
        </form>
      </CommandBarContainer>
      <NavigationBlockerDialog blockNavigation={isDirty} />
    </FormProvider>
  );
};

export default EditPictureScreen;
