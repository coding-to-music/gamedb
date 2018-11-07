if ($('#app-page').length > 0) {

    // Background
    const background = $('.container[data-bg]').attr('data-bg');
    if (background !== '') {
        $('body').css("background-image", 'url(' + background + ')');
    }

    // News
    const $collapseBoxes = $('#news .collapse');

    $collapseBoxes.collapse();
    $collapseBoxes.first().collapse('show');

    // Fix links
    $('#news a').each(function () {

        const href = $(this).attr('href');
        if (href && !(href.startsWith('http'))) {
            $(this).attr('href', 'http://' + href);
        }
    });
}
