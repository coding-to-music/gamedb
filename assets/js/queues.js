if ($('#queue-page').length > 0) {

    // Pause on tab change
    var paused = false;
    var $badge = $('#live-badge');

    $(window).blur(function () {
        paused = true;
        $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
    }).focus(function () {
        paused = false;
        $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
    });

    // 5 second timer.
    var time = 1;
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
            var $body = $('tbody');
            $.ajax({
                async: true,
                cache: false,
                dataType: 'json',
                method: 'GET',
                url: "/queues/queues.json",
                success: function (data, status) {
                    $body.empty();
                    for (var i in data) {
                        $body.append($('<tr><td>' + data[i].Name + '</td><td>' + data[i].Messages + '</td><td>' + data[i].Rate + '/s</td></tr>'));
                    }
                }
            });
        }
    }
}
