const $apiPage = $('#api-page');

if ($apiPage.length > 0) {

    $('#sidebar').stickySidebar({
        topSpacing: 0,
        bottomSpacing: 16,
    });

    $('.endpoint').on('mouseenter', function () {
        $(this).select();
    });
}
