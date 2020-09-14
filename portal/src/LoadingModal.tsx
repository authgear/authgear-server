import React from "react";
import { Modal } from "@fluentui/react";
import ShowLoading from "./ShowLoading";

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
    <Modal
      isOpen={true}
      styles={{
        main: { display: "flex", minWidth: "130px", minHeight: "130px" },
      }}
    >
      <ShowLoading />
    </Modal>
  );
};

export default LoadingModal;
