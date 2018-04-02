if ($('#coop-page').length > 0) {

    $('form#add').submit(function (e) {

        e.preventDefault();

        var url = '';

        var val = $('input#id').val();

        if (document.location.href.indexOf("?") >= 0) {
            url = document.location.href + "&p=" + val;
        } else {
            url = document.location.href + "?p=" + val;
        }

        document.location = url;
    });

    $('input#addme').click(function (e) {

        var id = $(this).attr('data-id');

        $('input#id').val(id);
        $('form#add').submit();
    });
}
