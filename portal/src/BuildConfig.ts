/* global process */
// @ts-expect-error
export const BUILD_TIME_VARIABLE: string = process.env.BUILD_TIME_VARIABLE;
// Parcel supports process.env.* to inject build time variable via environment variable.
// However, we are not going to use much build time variable.
// This above one is just for documentation purpose.
