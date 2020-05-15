if ($('#coop-page').length > 0) {

    $('form#add').on('submit', function (e) {

        e.preventDefault();

        let url = '';

        const val = $('input#id').val();

        if (window.location.href.indexOf("?") >= 0) {
            url = window.location.href + "&p=" + val;
        } else {
            url = window.location.href + "?p=" + val;
        }

        window.location = url;
    });

    $('#addme input').on('click', function (e) {

        const val = $(this).attr('data-id');

        $('input#id').val(val);
        $('form#add').trigger('submit');
    });
}
