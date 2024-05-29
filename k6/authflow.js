import { check, fail } from "k6";
import http from "k6/http";
import { URL } from "https://jslib.k6.io/url/1.0.0/index.js";

import {
  ENDPOINT,
  CLIENT_ID,
  REDIRECT_URI,
  FIXTURE_PASSWORD,
  FIXTURE_FIXED_OTP,
} from "./env.js";
import {
  makeAuthenticationURL,
  startAuthentication,
  extractAuthorizationCode,
  exchangeAuthorizationCode,
} from "./oauth.js";

function checkJSON(json) {
  return check(json, {
    [`no error`]: (json) =>
      !Object.prototype.hasOwnProperty.call(json, "error"),
    [`has result`]: (json) =>
      Object.prototype.hasOwnProperty.call(json, "result"),
  });
}

export function authflowCreate({ type, name, questionMarkQuery }) {
  const url = new URL("/api/v1/authentication_flows", ENDPOINT);
  url.search = questionMarkQuery;
  const payload = JSON.stringify({
    type,
    name,
  });
  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };
  const response = http.post(url.toString(), payload, params);
  const json = response.json();
  const checkResult = checkJSON(json);
  if (!checkResult) {
    fail(`failed to create authflow ${type}:${name}`);
  }
  return {
    response,
    json,
    result: json.result,
  };
}

export function authflowInput({ result, input }) {
  const { state_token } = result;
  const url = new URL("/api/v1/authentication_flows/states/input", ENDPOINT);
  const payload = JSON.stringify({
    state_token,
    input,
  });
  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };
  const response = http.post(url.toString(), payload, params);
  const json = response.json();
  const checkResult = checkJSON(json);
  if (!checkResult) {
    fail(`failed to input: ${state_token}`);
  }
  return {
    response,
    json,
    result: json.result,
  };
}

export function redirect(result) {
  const checkResult = check(result, {
    [`action type is finished`]: (r) => r.action.type === "finished",
    [`finish_redirect_uri exists`]: (r) =>
      typeof r.action.data.finish_redirect_uri === "string",
  });
  if (!checkResult) {
    fail(`unexpected action: ${result.action}`);
  }
  return http.get(result.action.data.finish_redirect_uri);
}

function authflowIdentify({ username, phone, email, result }) {
  const options = result.action.data.options;
  const first = options[0];
  switch (first.identification) {
    case "email":
      return authflowInput({
        result,
        input: {
          identification: "email",
          login_id: email,
        },
      });
    case "phone":
      return authflowInput({
        result,
        input: {
          identification: "phone",
          login_id: phone,
        },
      });
    case "username":
      return authflowInput({
        result,
        input: {
          identification: "username",
          login_id: username,
        },
      });
    default:
      throw new Error("unexpected identification: ", first.identification);
  }
}

function authflowVerify({ result }) {
  return authflowInput({
    result,
    input: {
      code: FIXTURE_FIXED_OTP,
    },
  });
}

function authflowCreateAuthenticator({ phone, email, result }) {
  const options = result.action.data.options;
  const first = options[0];
  switch (first.authentication) {
    case "primary_password":
      return authflowInput({
        result,
        input: {
          authentication: "primary_password",
          new_password: FIXTURE_PASSWORD,
        },
      });
    case "secondary_password":
      return authflowInput({
        result,
        input: {
          authentication: "secondary_password",
          new_password: FIXTURE_PASSWORD,
        },
      });
    case "primary_oob_otp_email":
      return authflowInput({
        result,
        input: {
          authentication: "primary_oob_otp_email",
          index: 0,
          channel: first.channels[0],
          code: FIXTURE_FIXED_OTP,
        },
      });
    case "primary_oob_otp_sms":
      return authflowInput({
        result,
        input: {
          authentication: "primary_oob_otp_sms",
          index: 0,
          channel: first.channels[0],
          code: FIXTURE_FIXED_OTP,
        },
      });
    default:
      throw new Error("unexpected authentication: ", result.action.data);
  }
}

function authflowAuthenticate({ result }) {
  const options = result.action.data.options;
  const first = options[0];
  switch (first.authentication) {
    case "primary_password":
      return authflowInput({
        result,
        input: {
          authentication: "primary_password",
          password: FIXTURE_PASSWORD,
        },
      });
    case "secondary_password":
      return authflowInput({
        result,
        input: {
          authentication: "secondary_password",
          password: FIXTURE_PASSWORD,
        },
      });
    case "primary_oob_otp_email": {
      let ret = authflowInput({
        result,
        input: {
          authentication: "primary_oob_otp_email",
          index: 0,
          channel: first.channels[0],
        },
      });
      return authflowInput({
        result: ret.result,
        input: {
          code: FIXTURE_FIXED_OTP,
        },
      });
    }
    case "primary_oob_otp_sms": {
      let ret = authflowInput({
        result,
        input: {
          authentication: "primary_oob_otp_sms",
          index: 0,
          channel: first.channels[0],
        },
      });
      return authflowInput({
        result: ret.result,
        input: {
          code: FIXTURE_FIXED_OTP,
        },
      });
    }
    default:
      throw new Error("unexpected authentication: ", result.action.data);
  }
}

export function authflowRun({ username, phone, email, type, name }) {
  const url = makeAuthenticationURL({
    endpoint: ENDPOINT,
    client_id: CLIENT_ID,
    redirect_uri: REDIRECT_URI,
  });

  const customUIURLString = startAuthentication(url);
  const customUIURL = new URL(customUIURLString);

  let ret;
  ret = authflowCreate({
    type,
    name,
    questionMarkQuery: customUIURL.search,
  });

  while (ret.result.action.type !== "finished") {
    switch (ret.result.action.type) {
      case "identify":
        ret = authflowIdentify({ username, phone, email, result: ret.result });
        break;
      case "verify":
        ret = authflowVerify({ result: ret.result });
        break;
      case "create_authenticator":
        ret = authflowCreateAuthenticator({ phone, email, result: ret.result });
        break;
      case "authenticate":
        ret = authflowAuthenticate({ result: ret.result });
        break;
      default:
        throw new Error("unexpected action type: ", ret);
    }
  }

  const response = redirect(ret.result);
  const code = extractAuthorizationCode(response);
  const tokenResponse = exchangeAuthorizationCode({
    endpoint: ENDPOINT,
    client_id: CLIENT_ID,
    redirect_uri: REDIRECT_URI,
    code,
  });
  const checkResult = check(tokenResponse, {
    access_token: (r) => typeof r.access_token === "string",
    refresh_token: (r) => typeof r.refresh_token === "string",
  });
  if (!checkResult) {
    fail("failed to exchange code");
  }
}
