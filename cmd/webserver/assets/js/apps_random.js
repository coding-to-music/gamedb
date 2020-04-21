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

}