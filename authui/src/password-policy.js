import zxcvbn from "zxcvbn";

function setupPage() {
    function checkPasswordLength(value, el) {
        if (el == null) {
            return;
        }
        var minLength = parseInt(el.getAttribute("data-min-length"), 10);
        // .length is number of UTF-16 code units,
        // while the server is counting number of UTF-8 code units.
        if (value.length >= minLength) {
            el.classList.add("good-txt");
        }
    }

    function checkPasswordUppercase(value, el) {
        if (el == null) {
            return;
        }
        if (/[A-Z]/.test(value)) {
            el.classList.add("good-txt");
        }
    }

    function checkPasswordLowercase(value, el) {
        if (el == null) {
            return;
        }
        if (/[a-z]/.test(value)) {
            el.classList.add("good-txt");
        }
    }

    function checkPasswordDigit(value, el) {
        if (el == null) {
            return;
        }
        if (/[0-9]/.test(value)) {
            el.classList.add("good-txt");
        }
    }

    function checkPasswordSymbol(value, el) {
        if (el == null) {
            return;
        }
        if (/[^a-zA-Z0-9]/.test(value)) {
            el.classList.add("good-txt");
        }
    }

    function checkPasswordStrength(value) {
        var meter = document.querySelector("#password-strength-meter");
        var desc = document.querySelector("#password-strength-meter-description");
        if (meter == null || desc == null) {
            return;
        }

        meter.value = 0;
        desc.textContent = "";

        if (value === "") {
            return;
        }

        var result = zxcvbn(value);
        var score = Math.min(5, Math.max(1, result.score + 1));
        meter.value = score;
        desc.textContent = desc.getAttribute("data-desc-" + score);
    }

    function attachPasswordPolicyCheck() {
        var el = document.querySelector("[data-password-policy-password]");
        if (el == null ) {
            return;
        }
        el.addEventListener("input", function(e) {
            var value = e.currentTarget.value;
            var els = document.querySelectorAll(".password-policy");
            for (var i = 0; i < els.length; ++i) {
                els[i].classList.remove("error-txt", "good-txt");
            }
            checkPasswordLength(value, document.querySelector(".password-policy.length"));
            checkPasswordUppercase(value, document.querySelector(".password-policy.uppercase"));
            checkPasswordLowercase(value, document.querySelector(".password-policy.lowercase"));
            checkPasswordDigit(value, document.querySelector(".password-policy.digit"));
            checkPasswordSymbol(value, document.querySelector(".password-policy.symbol"));
            checkPasswordStrength(value);
        });
    }

    attachPasswordPolicyCheck();
}
window.addEventListener("load", setupPage);
