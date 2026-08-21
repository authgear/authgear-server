import React from "react";
import { Spinner } from "@radix-ui/themes";
import styles from "./ShowLoading.module.css";

// ShowLoading show a 100% width and 100% height spinner.
const ShowLoading: React.VFC = function ShowLoading() {
  return (
    <div className={styles.loading}>
      <Spinner size="3" />
    </div>
  );
};

export default ShowLoading;
