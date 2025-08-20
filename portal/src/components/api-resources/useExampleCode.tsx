export enum ExampleCodeVariant {
  curl = "curl",
  Python = "Python",
  Go = "Go",
  NodeJS = "NodeJS",
}

interface TokenExchangeExampleCodeProps {
  variant: ExampleCodeVariant;
  resourceURI: string;
  tokenEndpoint: string;
  clientID: string;
  clientSecret: string | null;
}

export function useExampleCode({
  variant,
  tokenEndpoint,
  resourceURI,
  clientSecret,
  clientID,
}: TokenExchangeExampleCodeProps): string {
  switch (variant) {
    case ExampleCodeVariant.curl:
      return `curl --request POST \\
  --url ${tokenEndpoint} \\
  --header 'Content-Type: application/x-www-form-urlencoded' \\
  --data grant_type=client_credentials \\
  --data resource=${resourceURI} \\
  --data client_id=${clientID} \\
  --data client_secret=${clientSecret ?? "********"}
`;
    case ExampleCodeVariant.Python:
      return `import urllib.parse
import urllib.request
import json

url = "${tokenEndpoint}"
headers = {
    "Content-Type": "application/x-www-form-urlencoded"
}
data = {
    "grant_type": "client_credentials",
    "resource": "${resourceURI}",
    "client_id": "${clientID}",
    "client_secret": "${clientSecret ?? "********"}"
}

encoded_data = urllib.parse.urlencode(data).encode('utf-8')
req = urllib.request.Request(url, data=encoded_data, headers=headers, method='POST')

with urllib.request.urlopen(req) as response:
    response_status = response.getcode()
    response_body = response.read().decode('utf-8')
    print("Response Status Code:", response_status)
    print("Response Body:", json.loads(response_body))
`;
    case ExampleCodeVariant.Go:
      return `package main

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"
)

func main() {
  data := url.Values{}
  data.Set("grant_type", "client_credentials")
  data.Set("resource", "${resourceURI}")
  data.Set("client_id", "${clientID}")
  data.Set("client_secret", "${clientSecret ?? "********"}")

  req, _ := http.NewRequest("POST", "${tokenEndpoint}", strings.NewReader(data.Encode()))
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

  resp, _ := http.DefaultClient.Do(req)
  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  fmt.Println("Response Status Code:", resp.StatusCode)
  fmt.Println("Response Body:", string(body))
}
`;
    case ExampleCodeVariant.NodeJS:
      return `async function makeRequest() {
  const url = "${tokenEndpoint}";
  const data = new URLSearchParams();
  data.append("grant_type", "client_credentials");
  data.append("resource", "${resourceURI}");
  data.append("client_id", "${clientID}");
  data.append("client_secret", "${clientSecret ?? "********"}");

  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: data,
  });

  const responseBody = await response.json();
  console.log("Response Status Code:", response.status);
  console.log("Response Body:", responseBody);
}

makeRequest();
`;
    default:
      return "";
  }
}
