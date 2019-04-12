package template

const templateWelcomeEmailTxt = `Hello {% if user.LoginIDs.username %}{{ user.LoginIDs.username }}{% else %}{{ user.LoginIDs.email }}{% endif %},

Welcome to Skygear.

Thanks.`
