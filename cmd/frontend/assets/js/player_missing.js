const $playerMissingPage = $('#player-missing-page');

if ($playerMissingPage.length > 0) {

    websocketListener('profile', function (e) {

        if (queue_current > 0) {
            queue_current--;
        }

        updateLoadingBar();

        const data = JSON.parse(e.data);
        if (data.Data['queue'] === 'player' && data.Data['id'] === $playerMissingPage.attr('data-id')) {

            toast(true, '', 'Player found!');

            location.reload();
        }
    });

    function updateLoadingBar() {

        if (queue_current > 0) {

            let percent = 0;
            if (queue_start > 0) {
                percent = 100 - (queue_current / queue_start * 100);
            }

            if (percent < 5) {
                percent = 5;
            }

            let text = queue_current.toLocaleString() + ' / ' + queue_start.toLocaleString();

            $('.progress .progress-bar')
                .html(text)
                .width(percent + '%');

        } else {

            $('.progress .progress-bar')
                .html('You\'re Next!')
                .width('100%');
        }
    }

    updateLoadingBar();
}
