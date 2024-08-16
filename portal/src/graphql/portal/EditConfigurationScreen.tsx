import React from "react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { FormattedMessage } from "@oursky/react-messageformat";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";

import styles from "./EditConfigurationScreen.module.css";

const SECTIONS_CONFIG_YAML: [EditTemplatesWidgetSection] = [
  {
    key: "authgear.yaml",
    title: null,
    items: [
      {
        key: "authgear.yaml",
        title: null,
        editor: "code",
        language: "yaml",

        // TODO: implement value & onchange
        value: "foobar",
        onChange: () => {},
      },
    ],
  },
];

const EditConfigurationScreen: React.VFC = function EditConfigurationScreen() {
  return (
    <>
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="EditConfigurationScreen.title" />
        </ScreenTitle>
        <EditTemplatesWidget
          className={styles.widget}
          sections={SECTIONS_CONFIG_YAML}
        />
      </ScreenContent>
    </>
  );
};

export default EditConfigurationScreen;
