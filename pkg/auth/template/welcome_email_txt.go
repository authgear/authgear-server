package template

const templateWelcomeEmailTxt = `Hello {% if user_metadata.name %}{{ user_metadata.name }}{% else %}{{ user_metadata.email }}{% endif %},

Welcome to Skygear.

Thanks.`
