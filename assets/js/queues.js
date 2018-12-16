if ($('#queues-page').length > 0) {

    $(document).ready(function (e) {

        const full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');

        Highcharts.chart('chart', {
            chart: {
                type: 'line'
            },
            title: {
                text: ''
            },
            subtitle: {
                text: ''
            },
            credits: {
                enabled: false
            },
            legend: {
                enabled: false
            },
            xAxis: {
                title: {
                    text: 'Time'
                },
                labels: {
                    step: 1
                }
            },
            yAxis: {
                title: {
                    text: 'Items in the queue'
                }
            },
            series: [{
                color: '#28a745',
            }],
            data: {
                rowsURL: full + '/queues/ajax.json',
                enablePolling: true,
                dataRefreshRate: 60
            }
        });

    });

    // // Pause on tab change
    // let paused = false;
    // const $badge = $('#live-badge');
    //
    // $(window).blur(function () {
    //     paused = true;
    //     $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
    // }).focus(function () {
    //     paused = false;
    //     $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
    // });
    //
    // // 5 second timer.
    // let time = 1;
    // setInterval(function () {
    //
    //     time--;
    //     if (time < 1) {
    //         time = 5;
    //         update();
    //     }
    //
    // }, 1000);
    //
    // // Update the table
    // function update() {
    //     if (paused === false) {
    //         const $body = $('tbody');
    //         $.ajax({
    //             async: true,
    //             cache: false,
    //             dataType: 'json',
    //             method: 'GET',
    //             url: "/queues/ajax",
    //             success: function (data, status) {
    //                 $body.empty();
    //                 if (isIterable(data)) {
    //                     for (const v of data) {
    //                         $body.append($('<tr><td>' + v.Name + '</td><td>' + v.Messages + '</td><td>' + v.Rate + '/s</td></tr>'));
    //                     }
    //                 }
    //             }
    //         });
    //     }
    // }
}
