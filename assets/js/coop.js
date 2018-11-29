if ($('#coop-page').length > 0) {

    if (user.isLoggedIn) {
        $('#addme').removeClass('d-none')
    }

    $('form#add').submit(function (e) {

        e.preventDefault();

        let url = '';

        const val = $('input#id').val();

        if (document.location.href.indexOf("?") >= 0) {
            url = document.location.href + "&p=" + val;
        } else {
            url = document.location.href + "?p=" + val;
        }

        document.location = url;
    });

    $('#addme input').click(function (e) {

        $('input#id').val(user.userID);
        $('form#add').submit();
    });
}
