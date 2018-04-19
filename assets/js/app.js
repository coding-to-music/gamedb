if ($('#app-page').length > 0) {

    // Background
    var background = $('.container[data-bg]').attr('data-bg');
    if (background !== '') {
        $('body').css("background-image", 'url(' + background + ')').css("background-color", '#1b2738');
        $('nav.navbar').removeClass('bg-dark');
    }

    // News
    var $collapseBoxes = $('#news .collapse');

    $collapseBoxes.collapse();
    $collapseBoxes.first().collapse('show');

    // Fix links
    $('#news a').each(function () {

        var href = $(this).attr('href');
        if (!href.startsWith('http')) {
            $(this).attr('href', 'http://' + href);
        }

    });
}
