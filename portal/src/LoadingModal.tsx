import React from "react";
import ShowLoading from "./ShowLoading";

import styles from "./LoadingModal.module.scss";

interface LoadingModalProps {
  loading: boolean;
}

const LoadingModal: React.FC<LoadingModalProps> = function LoadingModal(
  props: LoadingModalProps
) {
  if (!props.loading) {
    return null;
  }

  return (
    <div className={styles.root}>
      <ShowLoading />
    </div>
  );
};

export default LoadingModal;
