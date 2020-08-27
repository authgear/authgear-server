import React, { useContext } from "react";
import { Spinner } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import styles from "./ShowLoading.module.scss";

// ShowLoading show a 100% width and 100% height spinner.
// For better UX, please use Shimmer instead.
const ShowLoading: React.FC = function ShowLoading() {
  const { renderToString } = useContext(Context);

  return (
    <div className={styles.loading}>
      <Spinner label={renderToString("loading")} />
    </div>
  );
};

export default ShowLoading;
