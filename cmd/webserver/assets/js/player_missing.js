const $playerMissingPage = $('#player-missing-page');

if ($playerMissingPage.length > 0) {

    websocketListener('profile', function (e) {

        if (queue_current > 0) {
            queue_current--;
        }

        updateLoadingBar();

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $playerMissingPage.attr('data-id')) {

            toast(true, '', 'Player found!');

            location.reload();
        }
    });

    function updateLoadingBar() {

        //
        let percent = 100;
        if (queue_start > 0) {
            percent = queue_current / queue_start * 100;
            percent = Math.min(Math.max(percent, 0), 100);
        }
        percent = 100 - percent;

        //
        let text;
        if (queue_current > 0) {
            text = queue_current.toLocaleString() + ' / ' + queue_start.toLocaleString()
            text = percent.toLocaleString() + '%';
        } else {
            text = 'Next!';
        }

        //
        $('.progress .progress-bar').html(text).width(text);
    }

    $(updateLoadingBar);
}
