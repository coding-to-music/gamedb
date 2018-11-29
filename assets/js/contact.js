if ($('#contact-page').length > 0) {

    $('#name').val(user.session['login-name']);
    $('#email').val(user.session['login-email']);
    $('#message').val(user.session['login-message']);

}
