if ($('#queues-page').length > 0 || $('#player-missing-page').length > 0) {

    const $playerPage = $('#player-missing-page');

    websocketListener('profile', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $playerPage.attr('data-id')) {

            toast(true, '', 'Player found!');

            location.reload();
        }
    });
}
