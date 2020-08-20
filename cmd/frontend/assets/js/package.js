const $packagePage = $('#package-page');

if ($packagePage.length > 0) {

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (!to.attr('loaded')) {
            to.attr('loaded', 1);
            switch (to.attr('href')) {
                case '#prices':
                    loadPriceChart();
                    break;
            }
        }
    });

    // Websockets
    websocketListener('package', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $packagePage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This package has been updated', 0, 'refresh');
        }
    });
}
