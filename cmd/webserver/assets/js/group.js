const $groupPage = $('#group-page');

if ($groupPage.length > 0) {

    // Websockets
    websocketListener('group', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $groupPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This group has been updated', 0, 'refresh');
        }
    });

    loadGroupChart();
    loadGroupPlayers();
}

function loadGroupPlayers() {

    $.ajax({
        type: "GET",
        url: '/groups/' + $groupPage.attr('data-group-id') + '/table.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            const options = {
                "order": [[0, 'asc']],
                "createdRow": function (row, data, dataIndex) {
                    $(row).attr('data-link', data[2]);
                    $(row).attr('data-player-id', data[0]);
                },
                "columnDefs": [
                    // Icon / Name
                    {
                        "targets": 0,
                        "render": function (data, type, row) {
                            return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                        },
                        "createdCell": function (td, cellData, rowData, row, col) {
                            $(td).addClass('img');
                        },
                        "orderable": false,
                    },
                ]
            };

            const searchFields = [
                $('#items-search'),
            ];

            $('#players').gdbTable({tableOptions: options, searchFields: searchFields});
        },
    });
}

function loadGroupChart($page = null) {

    const $groupChart = $('#group-chart');
    if ($groupChart.length === 0) {
        return
    }

    $page = $page || $groupPage;

    // Load chart
    $.ajax({
        type: "GET",
        url: '/groups/' + $page.attr('data-group-id') + '/members.json',
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

            Highcharts.chart('group-chart', {
                chart: {
                    type: 'spline',
                    backgroundColor: 'rgba(0,0,0,0)',
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
