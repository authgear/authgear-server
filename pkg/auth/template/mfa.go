package template

const templateMFAOOBCodeSMSText = `Your MFA code is: {{ .code }}`
const templateMFAOOBCodeEmailText = `Your MFA code is: {{ .code }}`
const templateMFAOOBCodeEmailHTML = `<!DOCTYPE html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
<p>Your MFA code is: {{ .code }}</p>
</body>
`
