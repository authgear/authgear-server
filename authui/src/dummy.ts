// On parcel@2.10.2, its swc-optimizer will crash if the function body is empty.
// So we add a dummy console.log function call.
import("@hotwired/stimulus").then((stimulus) => {
  console.log(stimulus);
});
import("@hotwired/turbo").then((turbo) => {
  console.log(turbo);
});
import("zxcvbn").then((zxcvbn) => {
  console.log(zxcvbn);
});
import("axios").then((axios) => {
  console.log(axios);
});
