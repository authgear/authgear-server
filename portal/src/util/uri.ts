export function base64EncodedDataToDataURI(base64EncodedData: string): string {
  return `data:;base64,${base64EncodedData}`;
}

export function dataURIToBase64EncodedData(dataURI: string): string {
  const idx = dataURI.indexOf(",");
  if (idx < 0) {
    throw new Error("not a data URI: " + dataURI);
  }
  return dataURI.slice(idx + 1);
}
