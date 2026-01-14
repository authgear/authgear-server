import React, { useContext } from "react";
import { Spinner } from "@fluentui/react";
import { Context } from "./intl";
import styles from "./ShowLoading.module.css";

interface ShowLoadingProps {
  label?: string;
}

// ShowLoading show a 100% width and 100% height spinner.
// For better UX, please use Shimmer instead.
const ShowLoading: React.VFC<ShowLoadingProps> = function ShowLoading({
  label,
}) {
  const { renderToString } = useContext(Context);

  return (
    <div className={styles.loading}>
      <Spinner label={label ?? renderToString("loading")} />
    </div>
  );
};

export default ShowLoading;
