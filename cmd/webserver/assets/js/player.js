const $playerPage = $('#player-page');

if ($playerPage.length > 0) {

    // Update link
    $('#update-button').on('click', function (e) {

        e.preventDefault();

        const $link = $(this);

        $('i, svg', $link).addClass('fa-spin');

        $.ajax({
            url: '/players/' + $playerPage.attr('data-id') + '/update.json',
            data: {
                'csrf': $(this).attr('data-csrf'),
            },
            dataType: 'json',
            cache: false,
            success: function (data, textStatus, jqXHR) {

                toast(data.success, data.toast);

                $('i', $link).removeClass('fa-spin');

                $link.contents().last()[0].textContent = ' In Queue';

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
        if (!to.attr('loaded')) {
            to.attr('loaded', 1);
            switch (to.attr('href')) {
                case '#charts':
                    loadPlayerCharts();
                    break;
                case '#games':
                    loadPlayerGames();
                    break;
                case '#badges':
                    loadPlayerBadges();
                    break;
                case '#friends':
                    loadPlayerFriends();
                    break;
                case '#groups':
                    loadPlayerGroups();
                    break;
                case '#wishlist':
                    loadPlayerWishlist();
                    break;
                case '#achievements':
                    $('a.nav-link[href="#achievements-summary"]').tab('show');
                    break;
                case '#achievements-summary':
                    loadPlayerAchievementsSummary();
                    break;
                case '#achievements-latest':
                    loadPlayerAchievementsLatest();
                    break;
            }
        }
    });

    // Websockets
    websocketListener('profile', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $playerPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This player has been updated', 0, 'refresh');
        }
    });

    function loadPlayerGames() {

        const options = {
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
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
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
                    'orderSequence': ['desc', 'asc'],
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Price/Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    'orderSequence': ['desc', 'asc'],
                }
            ]
        };

        $('#games #all-games').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#player-games-search'),
            ],
        });

        //
        const recentOptions = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[5]);
            },
            "columnDefs": [
                // Icon / Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[1] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
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
        };

        $('#games #recent-games').gdbTable({tableOptions: recentOptions});
    }

    function loadPlayerFriends() {

        const options = {
            "order": [[1, 'desc'], [4, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[1]);
            },
            "columnDefs": [
                // Icon / Friend
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" data-src="/assets/img/no-player-image.jpg" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + row[3] + '</div></div>'
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

                        if (row[4] === '' || row[4] === '-') {
                            $('#add-missing-friends').removeClass('d-none');
                        }

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
                // Friend Since
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[7];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Link
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#friends table.table').gdbTable({tableOptions: options});
    }

    function loadPlayerGroups() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[3]);
                $(row).attr('data-group-id', data[0]);
            },
            "columnDefs": [
                // Group
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        let badge = '';
                        if (row[7]) {
                            badge = '<span class="badge badge-success float-right">Primary</span>';
                        }

                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[4] + '" data-src="/assets/img/no-player-image.jpg" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + badge + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc'],
                },
                // Members
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[5].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Official
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "orderable": false,
                },
                // Link
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#groups-table').gdbTable({tableOptions: options});
    }

    function loadPlayerWishlist() {

        const options = {
            "order": [[0, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[2]);
            },
            "columnDefs": [
                // Rank
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        if (row[4] === 0) {
                            return '-';
                        }
                        return ordinal(row[4]);
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('font-weight-bold')
                    },
                    'orderSequence': ['asc'],
                },
                // App
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" data-src="/assets/img/no-player-image.jpg" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc'],
                },
                // Release State
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Release Date
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Price
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[7];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
            ]
        };

        $('#wishlist-table').gdbTable({tableOptions: options});
    }

    function loadPlayerAchievementsSummary() {

        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[0] + '#achievements');
                $(row).attr('data-app-id', data[3]);
            },
            "columnDefs": [
                // App / Achievement
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderSequence": ['asc', 'desc'],
                },
                // Have
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString() + '<small>/' + row[5].toLocaleString() + '</small>';
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Percent
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[6] + '%';
                    },
                    "orderSequence": ['desc', 'asc'],
                },
            ]
        };

        $('#achievements-summary-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadPlayerAchievementsLatest() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[0] + '#achievements');
            },
            "columnDefs": [
                // App / Achievement
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + row[1] + ': ' + row[3] + '<br><small>' + row[5] + '</small></div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Release Date
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        if (row[6]) {
                            return '<span data-livestamp="' + row[6] + '"></span>';
                        } else {
                            return 'Unknown';
                        }
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
            ]
        };

        $('#achievements-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadPlayerBadges() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                if (data[0]) {
                    $(row).attr('data-app-id', data[0]);
                }
                $(row).attr('data-link', data[2]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        let name = row[1];
                        if (row[9]) {
                            name += '<span class="badge badge-primary float-right ml-1">Special</span>';
                        }
                        if (row[10]) {
                            name += '<span class="badge badge-warning float-right ml-1">Event</span>';
                        }
                        if (row[4]) {
                            name += '<span class="badge badge-success float-right ml-1">Foil</span>';
                        }

                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[5] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + name + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Level / XP
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[6].toLocaleString() + ' (' + row[8].toLocaleString() + 'xp)';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Scarcity
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[7].toLocaleString();
                    },
                    "orderSequence": ['asc', 'desc'],
                },
                // Completion Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[3].toLocaleString();
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
            ]
        };

        $('#badges-table').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#player-badge-search'),
            ],
        });
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
