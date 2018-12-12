package template

const templateWelcomeEmailTxt = `Hello {% if user.name %}{{ user.name }}{% else %}{{ user.email }}{% endif %},

Welcome to Skygear.

Thanks.`
