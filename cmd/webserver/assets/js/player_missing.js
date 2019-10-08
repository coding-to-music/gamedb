const $playerMissingPage = $('#player-missing-page');

if ($playerMissingPage.length > 0) {

    websocketListener('profile', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $playerMissingPage.attr('data-id')) {

            toast(true, '', 'Player found!');

            location.reload();
        }
    });
}
