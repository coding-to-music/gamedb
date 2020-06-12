const $trendingAppsPage = $('#trending-apps-page');
const $trendingAppsTable = $('table.table');

if ($trendingAppsPage.length > 0) {

    const options = {
        "order": [[3, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / App Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img alt="" data-lazy="' + row[2] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Price
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Players
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Trend Value
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
            },
            // Chart
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<div data-app-id="' + row[0] + '"><i class="fas fa-spinner fa-spin"></i></div>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('chart');
                },
                "orderable": false,
            },
        ]
    };

    $trendingAppsTable.gdbTable({tableOptions: options});
}

if ($trendingAppsPage.length > 0 || $('#new-releases-page').length > 0) {

    $trendingAppsTable.on('draw.dt', function (e, settings, processing) {
        loadCharts();
    });

    function loadCharts() {

        const vals = $('td.chart div[data-app-id]')
            .map(function () {
                return $(this).attr('data-app-id');
            })
            .get()
            .join(',');

        $.ajax({
            type: "GET",
            url: '/games/trending/charts.json?ids=' + vals,
            dataType: 'json',
            success: function (datas, textStatus, jqXHR) {

                if (datas === null) {
                    return
                }

                $('div[data-app-id]').each(function (index) {

                    let data = {};
                    const appID = $(this).attr('data-app-id');

                    if (datas !== null && appID in datas && 'max_player_count' in datas[appID]) {
                        data = datas[appID]['max_player_count'];
                    } else {
                        data = [];
                    }

                    Highcharts.chart(this, {
                        chart: {
                            type: 'area',
                            margin: [0, 0, 0, 0],
                            skipClone: true,
                            height: 32,
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
                            enabled: false,
                            itemStyle: {
                                color: '#28a745',
                            },
                            itemHiddenStyle: {
                                color: '#666666',
                            },
                        },
                        xAxis: {
                            title: {text: null},
                            labels: {enabled: false},
                            type: 'datetime',
                        },
                        yAxis: {
                            title: {text: null},
                            labels: {enabled: false},
                            min: 0,
                        },
                        plotOptions: {
                            series: {
                                marker: {
                                    enabled: false
                                }
                            }
                        },
                        tooltip: {
                            hideDelay: 0,
                            outside: true,
                            shared: true,
                            formatter: function () {
                                return this.y.toLocaleString() + ' players on ' + moment(this.x).format("dddd DD MMM YYYY @ HH:mm");
                            },
                            style: {
                                'width': '500px',
                            }
                        },
                        series: [
                            {
                                color: '#28a745',
                                data: data,
                            },
                        ],
                    });
                });
            },
        });
    }
}
