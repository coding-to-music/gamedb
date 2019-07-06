const $playerPage = $('#player-page');

if ($playerPage.length > 0) {

    // Update link
    $('a[data-update-id]').on('click', function (e) {

        e.preventDefault();

        const $link = $(this);

        $('i', $link).addClass('fa-spin');

        $.ajax({
            url: '/players/' + $(this).attr('data-update-id') + '/update.json',
            dataType: 'json',
            cache: false,
            success: function (data, textStatus, jqXHR) {

                toast(data.success, data.toast);

                $('i', $link).removeClass('fa-spin');

                if (data.log) {
                    console.log(data.log);
                }
            },
        });
    });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (to.attr('href') === '#charts') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadPlayerCharts();
            }
        }
        if (to.attr('href') === '#games') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadPlayerGames();
            }
        }
        if (to.attr('href') === '#badges') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadPlayerBadges();
            }
        }
        if (to.attr('href') === '#friends') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadPlayerFriends();
            }
        }

        // On any tab
        $.each(dataTables, function (index, value) {
            value.fixedHeader.adjust();
        });
    });

    // Websockets
    websocketListener('profile', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $playerPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This player has been updated', -1, 'refresh');
        }

    });

    function loadPlayerGames() {

        const dt = $('#games #all-games').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[7]);
            },
            "columnDefs": [
                // Icon / Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<img data-lazy="' + row[2] + '" class="rounded square" data-lazy-alt="' + row[1] + '"><span>' + row[1] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    }
                },
                // Price
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    }
                },
                // Price/Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                }
            ]
        }));

        dataTables.push(dt);

        const dt2 = $('#games #recent-games').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[5]);
            },
            "columnDefs": [
                // Icon / Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<img data-lazy="' + row[1] + '" class="rounded square" data-lazy-alt="' + row[2] + '"><span>' + row[2] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    }
                },
                // Price
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[3].toLocaleString();
                    },
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    }
                },
            ]
        }));

        dataTables.push(dt2);
    }

    function loadPlayerFriends() {

        const dt = $('#friends table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[1]);
            },
            "columnDefs": [
                // Icon / Friend
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<img data-lazy="' + row[2] + '" class="rounded square" data-src="/assets/img/no-player-image.jpg" data-lazy-alt="' + row[3] + '"><span>' + row[3] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Level
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Games
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        if (!row[5]) {
                            return '-';
                        } else if (row[6] === 0) {
                            return $lockIcon;
                        } else {
                            return row[6].toLocaleString();
                        }
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Logged Off
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[7];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Friend Since
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[8];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
            ]
        }));

        dataTables.push(dt);
    }

    function loadPlayerBadges() {

        const dt = $('#badges table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[2]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<img data-lazy="' + row[5] + '" class="rounded square" data-lazy-alt="' + row[1] + '"><span>' + row[1] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                        $(td).attr('nowrap', 'nowrap');
                    },
                },
                // Level / XP
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[6].toLocaleString() + ' (' + row[8].toLocaleString() + 'xp)';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    }
                },
                // Scarcity
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[7].toLocaleString();
                    },
                },
                // Completion Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[3].toLocaleString();
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    }
                },
            ]
        }));

        dataTables.push(dt);
    }

    function loadPlayerCharts() {

        const defaultPlayerChartOptions = {
            chart: {
                type: 'line',
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
                enabled: true
            },
            plotOptions: {},
            xAxis: {
                title: {
                    text: ''
                },
                type: 'datetime'
            },
        };

        $.ajax({
            type: "GET",
            url: '/players/' + $playerPage.attr('data-id') + '/history.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                const yAxisHistory = {
                    allowDecimals: false,
                    title: {
                        text: ''
                    },
                    labels: {
                        enabled: false
                    },
                };

                Highcharts.chart('history-chart', $.extend(true, {}, defaultPlayerChartOptions, {

                    yAxis: [
                        yAxisHistory,
                        yAxisHistory,
                        yAxisHistory,
                        yAxisHistory,
                        yAxisHistory,
                    ],
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' ' + this.series.name.toLowerCase() + ' on ' + moment(this.key).format("dddd DD MMM YYYY");
                        },
                    },
                    series: [
                        {
                            name: 'Level',
                            color: '#28a745',
                            data: data['mean_level'],
                            marker: {symbol: 'circle'},
                            yAxis: 0,
                        },
                        {
                            name: 'Games',
                            color: '#007bff',
                            data: data['mean_games'],
                            marker: {symbol: 'circle'},
                            yAxis: 1,
                        },
                        {
                            name: 'Badges',
                            color: '#e83e8c',
                            data: data['mean_badges'],
                            marker: {symbol: 'circle'},
                            yAxis: 2,
                        },
                        {
                            name: 'Playtime',
                            color: '#ffc107',
                            data: data['mean_playtime'],
                            marker: {symbol: 'circle'},
                            yAxis: 3,
                        },
                        {
                            name: 'Friends',
                            color: '#343a40',
                            data: data['mean_friends'],
                            marker: {symbol: 'circle'},
                            yAxis: 4,
                        },
                    ],
                }));

                const yAxisRanks = {
                    allowDecimals: false,
                    title: {
                        text: ''
                    },
                    reversed: true,
                    min: 1,
                    labels: {
                        enabled: false
                    },
                };

                Highcharts.chart('ranks-chart', $.extend(true, {}, defaultPlayerChartOptions, {
                    yAxis: [
                        yAxisRanks,
                        yAxisRanks,
                        yAxisRanks,
                        yAxisRanks,
                        yAxisRanks,
                    ],
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' rank ' + this.y.toLocaleString() + ' on ' + moment(this.key).format("dddd DD MMM YYYY");
                        },
                    },
                    series: [
                        {
                            name: 'Level',
                            color: '#28a745',
                            data: data['mean_level_rank'],
                            marker: {symbol: 'circle'},
                            yAxis: 0,
                        },
                        {
                            name: 'Games',
                            color: '#007bff',
                            data: data['mean_games_rank'],
                            marker: {symbol: 'circle'},
                            yAxis: 1,
                        },
                        {
                            name: 'Badges',
                            color: '#e83e8c',
                            data: data['mean_badges_rank'],
                            marker: {symbol: 'circle'},
                            yAxis: 2,
                        },
                        {
                            name: 'Playtime',
                            color: '#ffc107',
                            data: data['mean_playtime_rank'],
                            marker: {symbol: 'circle'},
                            yAxis: 3,
                        },
                        {
                            name: 'Friends',
                            color: '#343a40',
                            data: data['mean_friends_rank'],
                            marker: {symbol: 'circle'},
                            yAxis: 4,
                        }
                    ],
                }));

            },
        });

    }
}
