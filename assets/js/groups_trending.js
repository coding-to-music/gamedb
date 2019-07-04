if ($('#groups-trending-page').length > 0) {

    const $trendingGroupsTable = $('table.table-datatable2');

    $('form').on('submit', function (e) {

        $trendingGroupsTable.DataTable().draw();
        return false;
    });

    $('#type, #errors').on('change', function (e) {

        $trendingGroupsTable.DataTable().draw();
        return false;
    });

    $trendingGroupsTable.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            data.search = {};
            data.search.search = $('#search').val();
            data.search.type = $('#type').val();
            data.search.errors = $('#errors').val();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
            if (data[7] === 'game' && !$('#type').val()) {
                $(row).addClass('table-primary');
            }
            if (data[9]) {
                $(row).addClass('table-danger');
            }
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img data-src="/assets/img/no-app-image-square.jpg" src="' + row[3] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Trend Value
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[10].toLocaleString();
                },
                "orderSequence": ["asc", "desc"],
            },
            // Trend Chart
            // {
            //     "targets": 3,
            //     "render": function (data, type, row) {
            //         return '<div data-group-id="' + row[0] + '"><i class="fas fa-spinner fa-spin"></i></div>';
            //     },
            //     "createdCell": function (td, cellData, rowData, row, col) {
            //         $(td).addClass('chart');
            //     },
            //     "orderable": false,
            // },
            // Link
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<a href="' + row[8] + '" target="_blank" rel="nofollow"><i class="fas fa-link" data-target="_blank"></i></a>';
                },
                "orderable": false,
            },
        ]
    }));

    $trendingAppsTable.on('draw.dt', function (e, settings, processing) {
        loadCharts();
    });

    function loadCharts() {

        const vals = $('td.chart div[data-group-id]')
            .map(function () {
                return $(this).attr('data-group-id');
            })
            .get()
            .join(',');

        $.ajax({
            type: "GET",
            url: '/groups/trending/charts.json?ids=' + vals,
            dataType: 'json',
            success: function (datas, textStatus, jqXHR) {

                if (datas === null) {
                    return
                }

                $('div[data-group-id]').each(function (index) {

                    let data = {};
                    const appID = $(this).attr('data-group-id');

                    if (datas !== null && appID in datas && 'max_members_count' in datas[appID]) {
                        data = datas[appID]['max_members_count'];
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
                            enabled: false
                        },
                        xAxis: {
                            title: {text: null},
                            labels: {enabled: false},
                            type: 'datetime',
                        },
                        yAxis: {
                            allowDecimals: false,
                            title: {text: null},
                            labels: {enabled: false},
                            // min: 0,
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
                                return this.y.toLocaleString() + ' members at ' + moment(this.x).format("DD MMM YYYY @ HH:mm");
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
