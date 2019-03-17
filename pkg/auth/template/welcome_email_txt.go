package template

const templateWelcomeEmailTxt = `Hello {% if user.metadata.name %}{{ user.metadata.name }}{% else %}{{ user.metadata.email }}{% endif %},

Welcome to Skygear.

Thanks.`
