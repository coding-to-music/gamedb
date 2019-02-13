const $packagePage = $('#package-page');

if ($packagePage.length > 0) {

    // Link to dev tabs
    $(document).ready(function (e) {
        const hash = window.location.hash;
        if (hash.startsWith('#dev-')) {
            $('a.nav-link[href="#dev"]').tab('show');
            $('a.nav-link[href="' + hash + '"]').tab('show');
            window.location.hash = hash;
        }
    });

    // Websockets
    websocketListener('package', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $packagePage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This package has been updated', -1, 'refresh');
        }

    });
}
