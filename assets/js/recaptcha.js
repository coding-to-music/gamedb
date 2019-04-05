function recaptchaCallback(code) {

    $('form[data-recaptcha] button[type=submit]').prop("disabled", false);


    const inputs = $('form[data-recaptcha] input[type=text], form input[type=email], form textarea').filter(function () {
        return $(this).val() === '';
    });

    if (inputs.length > 0) {
        inputs.get(0).focus();
    } else {
        // $('form[data-recaptcha]').submit();
    }
}
