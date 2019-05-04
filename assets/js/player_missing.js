if ($('#queues-page').length > 0 || $('#player-missing-page').length > 0) {

    websocketListener('profile', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $playerPage.attr('data-id')) {

            toast(true, '', 'Player found!');

            location.reload();

        }

    });

}
