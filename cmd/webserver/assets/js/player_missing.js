const $playerMissingPage = $('#player-missing-page');

if ($playerMissingPage.length > 0) {

    websocketListener('profile', function (e) {

        // noinspection JSConstantReassignment
        queue_current--;

        updateLoadingBar();

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $playerMissingPage.attr('data-id')) {

            toast(true, '', 'Player found!');

            location.reload();
        }
    });

    function updateLoadingBar() {

        let p = 0;
        if (queue_start > 0) {
            p = queue_current / queue_start * 100;
            p = Math.min(Math.max(p, 0), 100);
            p = 100 - p;
        }
        p = p.toString() + '%';

        logLocal('total:', queue_start, 'on:', queue_current);

        $('.progress .progress-bar').html(queue_current.toString() + ' / ' + queue_start.toString()).width(p);
    }

    $(updateLoadingBar);
}
