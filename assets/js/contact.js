if ($('#contact-page').length > 0) {

    const $name = $('#name');
    const $email = $('#email');
    const $message = $('#message');

    $name.val(user.session['login-name']);
    $email.val(user.session['login-email']);
    $message.val(user.session['login-message']);

    if (!$email.val()) {
        $email.val(user.userEmail);
    }

}
