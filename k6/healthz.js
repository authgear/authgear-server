import { check, fail } from "k6";
import http from "k6/http";
import { URL } from "https://jslib.k6.io/url/1.0.0/index.js";
import { ENDPOINT } from "./env.js";

export default function () {
  const url = new URL("/healthz", ENDPOINT);
  const response = http.get(url.toString());
  const checkResult = check(response, {
    ["200"]: (response) => response.status === 200,
  });
  if (!checkResult) {
    fail("unexpected status code: ", response.status);
  }
}
