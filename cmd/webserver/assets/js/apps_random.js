if ($('#apps-random-page').length > 0) {

    // Find another button
    $('#find-another').on('click', function (e) {
        window.location.reload(true);
    });

    // Setup drop downs
    $('select.form-control-chosen').chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false,
        max_selected_options: 1
    });

    // Redirect on form change
    $('#search-card select').on('change', function (e) {
        const params = new URLSearchParams(location.search);
        if ($(this).val()) {
            params.set($(this).attr('name'), $(this).val());
        } else {
            params.delete($(this).attr('name'));
        }
        window.location.href = window.location.pathname + '?' + params.toString();
    })

    // Fill form on page load
    $(function () {
        const $selects = $('#search-card select');
        if (window.location.search) {
            $selects.deserialize(window.location.search.replace('?', ''));
        }
        $selects.trigger("chosen:updated");
    })
}