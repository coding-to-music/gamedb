if ($('#apps-page').length > 0) {

    $('select.form-control-chosen').chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false
    });
}

if ($('#app-page').length > 0) {

    $('.collapse').collapse();
}
