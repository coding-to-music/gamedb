const $packagePage = $('#package-page');

if ($packagePage.length > 0) {

    // Websockets
    websocketListener('package', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $packagePage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This package has been updated', 0, 'refresh');
        }

    });
}
