if ($('#queues-page').length > 0) {

    // Pause on tab change
    let paused = false;
    const $badge = $('#live-badge');

    $(window).blur(function () {
        paused = true;
        $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
    }).focus(function () {
        paused = false;
        $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
    });

    // 5 second timer.
    let time = 1;
    setInterval(function () {

        time--;
        if (time < 1) {
            time = 5;
            update();
        }

    }, 1000);

    // Update the table
    function update() {
        if (paused === false) {
            const $body = $('tbody');
            $.ajax({
                async: true,
                cache: false,
                dataType: 'json',
                method: 'GET',
                url: "/queues/queues.json",
                success: function (data, status) {
                    $body.empty();
                    if (isIterable(data)) {
                        for (const v of data) {
                            $body.append($('<tr><td>' + v.Name + '</td><td>' + v.Messages + '</td><td>' + v.Rate + '/s</td></tr>'));
                        }
                    }
                }
            });
        }
    }
}
