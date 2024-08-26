/* global it, expect */
import {
  cleanRawInputValue,
  trimCountryCallingCode,
  makePartialValue,
  validatePhoneNumber,
} from "./phone";

it("cleanRawInputValue", () => {
  expect(cleanRawInputValue("1234 1234")).toEqual("12341234");
  expect(cleanRawInputValue("asdf")).toEqual("");
  expect(cleanRawInputValue("我")).toEqual("");
  expect(cleanRawInputValue("+852+852")).toEqual("+852852");
});

it("trimCountryCallingCode", () => {
  expect(trimCountryCallingCode("+", "852")).toEqual("");
  expect(trimCountryCallingCode("+852", "852")).toEqual("");
  expect(trimCountryCallingCode("+85298", "852")).toEqual("98");
  expect(trimCountryCallingCode("98765432", "852")).toEqual("98765432");
});

it("makePartialValue", () => {
  expect(makePartialValue("+", "852")).toEqual("+852");
  expect(makePartialValue("+852", "852")).toEqual("+852");
  expect(makePartialValue("123", "852")).toEqual("+852123");
});

it("validatePhoneNumber", () => {
  // invalid
  expect(validatePhoneNumber("")).toEqual(null); // need digits
  expect(validatePhoneNumber("+")).toEqual(null); // need digits
  expect(validatePhoneNumber("+  ")).toEqual(null); // need digits
  expect(validatePhoneNumber("+852 9 9 9  99 999")).toEqual(null); // no space in middle

  // valid
  expect(validatePhoneNumber("+-")).toEqual("+-"); // edge case, expected
  expect(validatePhoneNumber("+852")).toEqual("+852"); // valid
  expect(validatePhoneNumber("   +81     ")).toEqual("+81"); // trimmed
  expect(validatePhoneNumber("\t\r\n+85299999999\t\r\n")).toEqual(
    "+85299999999"
  ); // trimmed
  expect(validatePhoneNumber("+852-9999-9999")).toEqual("+852-9999-9999"); // dashes in middle allowed
});
