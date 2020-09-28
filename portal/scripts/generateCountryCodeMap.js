const readline = require("readline");
const fs = require("fs");
const googleLibphonenumber = require("google-libphonenumber");
const { countryCallingCodes } = require("../src/data/countryCallingCode.json");

(async () => {
  try {
    await main();
    process.exit(0);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();

async function main() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  const phoneUtil = googleLibphonenumber.PhoneNumberUtil.getInstance();
  const countryCodeMap = {};

  for (const callingCode of countryCallingCodes) {
    const countryCodes = phoneUtil.getRegionCodesForCountryCode(callingCode);
    if (countryCodes.length === 0) {
      await new Promise((resolve) => {
        rl.question(
          "\n" +
            `No country code is found for calling code ${callingCode}.` +
            "\n\n" +
            "Please enter custom country code (alpha-2 code) or press enter to discard: ",
          function (code) {
            if (code.trim() !== "") {
              countryCodeMap[callingCode] = [code];
            }
            resolve();
          }
        );
      });
    } else if (!countryCodes.every((code) => /[A-Z]{2}/.test(code))) {
      await new Promise((resolve) => {
        rl.question(
          "\n" +
            `For calling code ${callingCode}, country codes found ` +
            `(${countryCodes.join(",")}) ` +
            "contains code which is not alpha-2 code." +
            "\n\n" +
            "Please enter country code or press enter to ignore: ",
          function (code) {
            if (code.trim() === "") {
              countryCodeMap[callingCode] = countryCodes;
            } else {
              countryCodeMap[callingCode] = [code];
            }
            resolve();
          }
        );
      });
    } else {
      countryCodeMap[callingCode] = countryCodes;
    }
  }

  console.log(
    "\n[INFO]: Writing generated map to src/data/countryCodeMap.json"
  );

  countryCodeMapData = JSON.stringify(countryCodeMap, null, 2);
  fs.writeFileSync("./src/data/countryCodeMap.json", countryCodeMapData);

  console.log("\n[INFO]: Successfully written to src/data/countryCodeMap.json");
}
