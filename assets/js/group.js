const $groupPage = $('#group-page');

if ($groupPage.length > 0) {

    // Websockets
    websocketListener('group', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $groupPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This group has been updated', -1, 'refresh');
        }
    });

    // Load chart
    $.ajax({
        type: "GET",
        url: '/groups/' + $groupPage.attr('data-id') + '/time.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            if (data === null) {
                data = [];
            }

            const yAxisGroup = {
                allowDecimals: false,
                title: {
                    text: ''
                },
                labels: {
                    // enabled: false
                },
                // min: 0,
            };

            Highcharts.chart('chart', {

                chart: {
                    type: 'line',
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
                plotOptions: {},
                xAxis: {
                    title: {
                        text: ''
                    },
                    type: 'datetime'

                },
                yAxis: [
                    Object.assign({}, yAxisGroup),
                    Object.assign({}, yAxisGroup),
                    Object.assign({}, yAxisGroup),
                    Object.assign({}, yAxisGroup),
                ],
                tooltip: {
                    formatter: function () {
                        return this.y.toLocaleString() + ' members on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                    },
                },
                series: [
                    {
                        name: 'Members',
                        color: '#28a745',
                        data: data['max_members_count'],
                        marker: {symbol: 'circle'},
                        yAxis: 0,
                    },
                    // {
                    //     name: 'In Chat',
                    //     color: '#007bff',
                    //     data: data['max_members_in_chat'],
                    //     marker: {symbol: 'circle'},
                    //     yAxis: 1,
                    // },
                    // {
                    //     name: 'In Game',
                    //     color: '#e83e8c',
                    //     data: data['max_members_in_game'],
                    //     marker: {symbol: 'circle'},
                    //     yAxis: 2,
                    // },
                    // {
                    //     name: 'Online',
                    //     color: '#ffc107',
                    //     data: data['max_members_online'],
                    //     marker: {symbol: 'circle'},
                    //     yAxis: 3,
                    // },
                ],
            });

        },
    });

}