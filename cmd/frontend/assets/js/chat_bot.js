if ($('#chat-bot-page').length > 0) {

    $('#commands-docs').gdbTable({
        order: [[0, "asc"], [1, "asc"]],
        tableOptions: {
            "drawCallback": function (settings) {
                const api = this.api();
                if (api.order()[0] && api.order()[0][0] === 0) {
                    const rows = api.rows({page: 'current'}).nodes();

                    let last = null;
                    api.rows().every(function (rowIdx, tableLoop, rowLoop) {
                        let group = this.data()[0].display;
                        if (last !== group) {
                            $(rows).eq(rowLoop).before(
                                '<tr class="table-success"><td colspan="4">' + group + '</td></tr>'
                            );
                            last = group;
                        }
                    });
                }
            },
        },
    });

    loadAjaxOnObserve({
        'chart': loadChart,
        'recent-table': loadLatest,
    });

    function loadChart() {

        $.ajax({
            type: "GET",
            url: '/discord-bot/chart.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: [
                        {
                            allowDecimals: false,
                            title: {
                                text: ''
                            },
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        },
                        {
                            allowDecimals: false,
                            title: {
                                text: ''
                            },
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                            opposite: true,
                        }
                    ],
                    tooltip: {
                        formatter: function () {
                            switch (this.series.name) {
                                case 'Requests':
                                    return this.y.toLocaleString() + ' requests at ' + moment(this.key).format("dddd DD MMM YYYY @ HH:00");
                                case 'Servers':
                                    return this.y.toLocaleString() + ' guilds on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:00");
                            }
                        },
                    },
                    plotOptions: {
                        series: {
                            fillColor: {
                                linearGradient: {x1: 0, x2: 0, y1: 0, y2: 1},
                                stops: [
                                    [0, defaultChartOptions.colors[0] + 'FF'],
                                    [1, defaultChartOptions.colors[0] + '00']
                                ]
                            }
                        }
                    },
                    series: [
                        {
                            name: 'Servers',
                            data: data['max_guilds'],
                            yAxis: 1,
                            color: defaultChartOptions.colors[1],
                            marker: {symbol: 'circle'},
                        },
                        {
                            name: 'Requests',
                            data: data['sum_request'],
                            yAxis: 0,
                            color: defaultChartOptions.colors[0],
                            marker: {symbol: 'circle'},
                            type: 'area',
                            step: 'left',
                        },
                    ],
                }));
            },
        });
    }

    function loadLatest() {

        const options = {
            "order": [[2, 'desc']],
            "columnDefs": [
                // Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name">' +
                            '<div class="icon"><img class="tall" alt="" data-lazy="https://cdn.discordapp.com/avatars/' + row[0] + '/' + row[2] + '.png?size=64" data-lazy-alt="' + row[1] + '"></div>' +
                            '<div class="name nowrap">' + row[1] + '<br><small>' + row[6] + '</small></div>' +
                            '</div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img thin');
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Message
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[3];
                    },
                    "orderable": false,
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        if (row[4] && row[4] > 0) {
                            return '<span data-livestamp="' + row[4] + '"></span>'
                                + '<br><small class="text-muted">' + row[5] + '</small>';
                        }
                        return '';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('thin');
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
            ]
        };

        const $table = $('#recent-table');
        const dt = $table.gdbTable({
            tableOptions: options,
        });

        websocketListener('chat-bot', function (e) {

            const info = dt.page.info();
            if (info.page === 0) { // Page 1

                const data = JSON.parse(e.data);
                addDataTablesRow(options, data.Data['row_data'], info.length, $table);
            }
        });
    }
}
