// This file is not intended to be imported.
// This file shows eslint-plugin-compat is working as intended.
// It is unlikely that all browsers we declare in browserslist will support PaymentRequest
// so this test should serve us well in a few years or so.

// @ts-expect-error
// eslint-disable-next-line compat/compat
const _ = new PaymentRequest();
