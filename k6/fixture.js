import exec from "k6/execution";
import { md5 } from "k6/crypto";
import {
  FIXTURE_EMAIL_DOMAIN,
  FIXTURE_PHONE_NUMBER_COUNTRY_CALLING_CODE,
  FIXTURE_PHONE_NUMBER_LOW,
  FIXTURE_PHONE_NUMBER_HIGH,
} from "./env.js";

export function makeLoginIDs(nationalPhone) {
  const username = `user${nationalPhone}`;
  const email = `${username}@${FIXTURE_EMAIL_DOMAIN}`;
  const phone = `${FIXTURE_PHONE_NUMBER_COUNTRY_CALLING_CODE}${nationalPhone}`;
  return {
    username,
    email,
    phone,
  };
}

function parseInteger(valueString, name) {
  const value = parseInt(valueString);
  if (isNaN(value)) {
    throw new Error(`${name} must be an integer: ${valueString}`);
  }
  if (value < 0) {
    throw new Error(`${name} must be non-negative: ${valueString}`);
  }
  return value;
}

function parseRange(lowValueString, lowName, highValueString, highName) {
  const low = parseInteger(lowValueString, lowName);
  const high = parseInteger(highValueString, highName);
  if (low >= high) {
    throw new Error(
      `${lowName} must be less than ${highName}: ${lowValueString} ${highValueString}`,
    );
  }
  return [low, high];
}

function getMinimumNumberOfDecimalDigits(n) {
  // Or n.String().length
  return Math.ceil(Math.log10(n));
}

function decimalDropRight(n, x) {
  return Math.floor(n / Math.pow(10, x));
}

function decimalPadZeroLeft(n, x) {
  const s = Array(x).fill("0").join("") + n.toString();
  return s.slice(-x);
}

export function makeNationalPhoneNumberForLogin({ vu }) {
  const [low, high] = parseRange(
    FIXTURE_PHONE_NUMBER_LOW,
    "FIXTURE_PHONE_NUMBER_LOW",
    FIXTURE_PHONE_NUMBER_HIGH,
    "FIXTURE_PHONE_NUMBER_HIGH",
  );
  const rangeValue = high - low;

  if (vu < 0 || vu > rangeValue) {
    throw new Error(`vu ${vu} is not in [0, ${rangeValue}]`);
  }

  return String(low + vu);
}

export function makeNationalPhoneNumberForSignup() {
  const totalVU = exec.instance.vusActive;
  const currentVU = exec.vu.idInTest;
  const iteration = exec.vu.iterationInInstance;

  const [low, high] = parseRange(
    FIXTURE_PHONE_NUMBER_LOW,
    "FIXTURE_PHONE_NUMBER_LOW",
    FIXTURE_PHONE_NUMBER_HIGH,
    "FIXTURE_PHONE_NUMBER_HIGH",
  );
  const rangeValue = high - low;

  const minimumNumberOfDecimalDigitsForVU =
    getMinimumNumberOfDecimalDigits(totalVU);
  const minimumNumberOfDecimalDigitsForRange =
    getMinimumNumberOfDecimalDigits(rangeValue);
  const minimumNumberOfDecimalDigitsForIteration =
    minimumNumberOfDecimalDigitsForRange - minimumNumberOfDecimalDigitsForVU;

  const maximumIteration = decimalDropRight(
    rangeValue,
    minimumNumberOfDecimalDigitsForVU,
  );

  if (iteration > maximumIteration) {
    throw new Error(
      `iteration exceeds maximum iteration: ${iteration} > ${maximumIteration}`,
    );
  }

  const iterationPart = decimalPadZeroLeft(
    iteration,
    minimumNumberOfDecimalDigitsForIteration,
  );
  const vuPart = decimalPadZeroLeft(
    currentVU,
    minimumNumberOfDecimalDigitsForVU,
  );
  const valueStr = iterationPart + vuPart;
  const value = parseInt(valueStr, 10);
  if (value < 0 || value > rangeValue) {
    throw new Error(`${value} is not in [0, ${rangeValue}]`);
  }

  return String(low + value);
}

export function makeFixedLoginLinkCode(userID) {
  return md5(userID, "hex");
}

export function getIndex() {
  const scenario = exec.test.options.scenarios[exec.scenario.name];
  const vus = scenario.vus;
  const vuID = exec.vu.idInTest;
  const index = (vuID - 1) % vus;
  return index;
}

export function getVU() {
  return getIndex() + 1;
}
