if ($('#contact-page').length > 0) {

    const $name = $('#name');
    const $email = $('#email');
    const $message = $('#message');

    $name.val(user.contactPage['name']);
    $email.val(user.contactPage['email']);
    $message.val(user.contactPage['message']);

    if (!$email.val()) {
        $email.val(user.userEmail);
    }
}
