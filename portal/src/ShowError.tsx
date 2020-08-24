import React from "react";

interface ShowErrorProps {
  error: unknown;
}

const ShowError: React.FC<ShowErrorProps> = function ShowError(
  props: ShowErrorProps
) {
  const { error } = props;
  if (error instanceof Error) {
    return (
      <div
        style={{
          whiteSpace: "pre",
        }}
      >
        {error.name}: {error.message}
        <br /> {error.stack}
      </div>
    );
  }
  return <div>Non-Error error: {String(error)}</div>;
};

export default ShowError;
