const $appPage = $('#app-page');

if ($appPage.length > 0) {

    // Play / Pause videos
    $('#media video').on('click', function (e) {
        const video = $(this)[0];
        if (video.paused) {
            pauseAllVideos();
            video.play()
        } else {
            video.pause()
        }
    });

    $('#details video').on('click', function (e) {
        const video = $(this)[0];
        if (video.paused) {
            video.play()
        } else {
            $('a.nav-link[href="#media"]').tab('show');
            $("#scroll-to-videos").trigger('click');
        }
    });

    // Followers link
    $('.followers-link').on('click', function (e) {

        $('a.nav-link[href="#players"]').tab('show');

        const st = $("#followers").offset().top;
        $('html, body').animate({scrollTop: st - 15}, 500);

        return false;
    });

    // Videos link
    $("#scroll-to-videos").on('click', function (e) {
        const st = $("#videos").offset().top;
        $('html, body').animate({scrollTop: st - 15}, 500);
    });

    // Show dev raw row
    $('#dev table.table tbody').on('click', 'td i, td svg', function () {

        const table = $(this).closest('table').DataTable();
        const $tr = $(this).closest('tr');
        const row = table.row($tr);

        if (row.child.isShown()) {

            row.child.hide();
            $tr.removeClass('shown');

        } else {

            row.child(function () {
                return '<div class="wbba">' + $tr.data('raw') + '</div>';
            }).show();
            $tr.addClass('shown');
        }
    });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (!to.attr('loaded')) {
            to.attr('loaded', 1);
            switch (to.attr('href')) {
                case '#news':
                    loadNews();
                    break;
                case '#items':
                    loadItems();
                    break;
                case '#prices':
                    loadPriceChart();
                    break;
                case '#players':
                    loadAppPlayersChart();
                    loadAppPlayerTimes();
                    loadGroupChart($appPage);
                    break;
                case '#reviews':
                    loadAppReviewsChart();
                    break;
            }
        }

        // On leaving tab
        if (from.attr('href') === '#media') {
            pauseAllVideos();
        }
    });

    function pauseAllVideos() {
        $('video').each(function (index) {
            $(this)[0].pause();
            // $(this)[0].currentTime = 0;
        });
    }

    // Websockets
    websocketListener('app', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $appPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This app has been updated', -1, 'refresh');
        }
    });

    // News data table
    function loadNews() {

        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
            },
            "columnDefs": [
                // Title
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div><i class="fas fa-newspaper"></i> ' + row[1] + '</div><div class="d-none">' + row[5] + '</div>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('style', 'min-width: 300px;')
                    },
                    "orderable": false
                },
                // Author
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[2];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                    },
                    "orderable": false
                },
                // Date
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
            ]
        };

        const $newsTable = $('#news-table');

        const table = $newsTable.gdbTable({tableOptions: options});

        $newsTable.on('click', 'tr[role=row]', function () {

            const row = table.row($(this));

            // noinspection JSUnresolvedFunction
            if (row.child.isShown()) {

                row.child.hide();
                $(this).removeClass('shown');

            } else {

                row.child($("<div/>").html(row.data()[5]).text()).show();
                $(this).addClass('shown');
            }
        });

        // Fix links
        $('#news a').each(function () {

            const href = $(this).attr('href');
            if (href && !(href.startsWith('http'))) {
                $(this).attr('href', 'http://' + href);
            }
        });
    }

    // News items
    function loadItems() {

        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
            },
            "columnDefs": [
                // Description
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return row[12];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                    },
                    "orderable": false
                },
                // Icon / Name
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[25] + '" alt="" data-lazy-alt="' + row[16] + '"></div><div class="name">' + row[16] + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Description
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[29];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                    },
                    "orderable": false
                },
                // Link
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[28]) {
                            return '<a href="' + row[28] + '" data-src="/assets/img/no-app-image-square.jpg" target="_blank" class="stop-prop"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        const $itemsTable = $('#items-table');

        const searchFields = [
            $('#items-search'),
        ];

        const table = $itemsTable.gdbTable({tableOptions: options, searchFields: searchFields});

        $itemsTable.on('click', 'tr[role=row]', function () {

                const row = table.row($(this));

                // noinspection JSUnresolvedFunction
                if (row.child.isShown()) {

                    row.child.hide();
                    $(this).removeClass('shown');

                } else {

                    const rowx = row.data();

                    const fields = [
                        {Name: "App ID", Value: rowx[0]},
                        {Name: "Bundle", Value: rowx[1]},
                        {Name: "Commodity", Value: rowx[2]},
                        {Name: "Date Created", Value: rowx[3]},
                        {Name: "Description", Value: rowx[4]},
                        {Name: "Display Type", Value: rowx[5]},
                        {Name: "Drop Interval", Value: rowx[6]},
                        {Name: "Drop Max Per Window", Value: rowx[7]},
                        {Name: "Exchange", Value: rowx[8]},
                        {Name: "Hash", Value: rowx[9]},
                        {Name: "Icon URL", Value: rowx[10]},
                        {Name: "Icon URL Large", Value: rowx[11]},
                        {Name: "Item Def ID", Value: rowx[12]},
                        {Name: "Item Quality", Value: rowx[13]},
                        {Name: "Marketable", Value: rowx[14]},
                        {Name: "Modified", Value: rowx[15]},
                        {Name: "Name", Value: rowx[16]},
                        {Name: "Price", Value: rowx[17]},
                        {Name: "Promo", Value: rowx[18]},
                        {Name: "Quantity", Value: rowx[19]},
                        {Name: "Tags", Value: rowx[20]},
                        {Name: "Timestamp", Value: rowx[21]},
                        {Name: "Tradable", Value: rowx[22]},
                        {Name: "Type", Value: rowx[23]},
                        {Name: "Workshop ID", Value: rowx[24]},
                    ];

                    const html = json2html.transform(fields, {
                        '<>': 'div', 'class': 'detail-row', 'html': [
                            {'<>': 'strong', 'html': '${Name}: '},
                            {'<>': 'span', 'html': '${Value}'},
                        ]
                    });

                    row.child('<img src="' + rowx[26] + '" alt="" class="float-right" />' + html).show();
                    $(this).addClass('shown');
                }
            }
        );
    }

    const defaultAppChartOptions = {
        chart: {
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
        plotOptions: {},
        xAxis: {
            title: {text: ''},
            type: 'datetime'
        },
    };

    function loadAppReviewsChart() {

        $.ajax({
            type: "GET",
            url: '/apps/' + $appPage.attr('data-id') + '/reviews.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('reviews-chart', $.extend(true, {}, defaultAppChartOptions, {
                    chart: {
                        type: 'line'
                    },
                    yAxis: [
                        {
                            allowDecimals: false,
                            title: {text: ''},
                            min: 0,
                            max: 100,
                            endOnTick: false,
                            labels: {
                                formatter: function () {
                                    return this.value + '%';
                                }
                            }
                        },
                        {
                            allowDecimals: false,
                            title: {text: ''},
                            opposite: true,
                            min: 0,
                        }
                    ],
                    legend: {
                        enabled: true
                    },
                    tooltip: {
                        formatter: function () {

                            const time = moment(this.key).format("DD MMM YYYY @ HH:mm");

                            if (this.series.name === 'score') {
                                return this.y.toLocaleString() + '% score on ' + time;
                            } else if (this.series.name === 'positive') {
                                return this.y.toLocaleString() + ' positive reviews on ' + time;
                            } else if (this.series.name === 'negative') {
                                return this.y.toLocaleString() + ' negative reviews on ' + time;
                            }
                        },
                    },
                    series: [
                        {
                            name: 'score',
                            color: '#28a745',
                            data: data['mean_reviews_score'],
                            yAxis: 0,
                            marker: {symbol: 'circle'}
                        },
                        {
                            name: 'positive',
                            color: '#e83e8c',
                            data: data['mean_reviews_positive'],
                            yAxis: 1,
                            marker: {symbol: 'circle'}
                        },
                        {
                            name: 'negative',
                            color: '#007bff',
                            data: data['mean_reviews_negative'],
                            yAxis: 1,
                            marker: {symbol: 'circle'}
                        },
                    ],
                }));

            },
        });
    }

    function loadAppPlayersChart() {

        const defaultAppChartOptions = {
            chart: {
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
            xAxis: {
                title: {text: ''},
                type: 'datetime'
            },
        };

        $.ajax({
            type: "GET",
            url: '/apps/' + $appPage.attr('data-id') + '/players.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    const now = Date.now();
                    data = {
                        "max_player_count": [[now, 0]],
                        "max_twitch_viewers": [[now, 0]],
                    };
                }

                Highcharts.chart('players-chart', $.extend(true, {}, defaultAppChartOptions, {
                    chart: {
                        type: 'area'
                    },
                    yAxis: [
                        {
                            allowDecimals: false,
                            title: {text: ''},
                            min: 0,
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        },
                        {
                            allowDecimals: false,
                            title: {text: ''},
                            min: 0,
                            opposite: true,
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        }
                    ],
                    legend: {
                        enabled: false
                    },
                    tooltip: {
                        formatter: function () {
                            if (this.series.name === 'Players') {
                                return this.y.toLocaleString() + ' players on ' + moment(this.key).format("DD MMM YYYY @ HH:mm");
                            } else {
                                return this.y.toLocaleString() + ' Twitch viewers on ' + moment(this.key).format("DD MMM YYYY @ HH:mm");
                            }
                        },
                    },
                    series: [
                        {
                            name: 'Players',
                            color: '#28a745',
                            data: data['max_player_count'],
                            yAxis: 0,
                        },
                        {
                            name: 'Viewers',
                            color: '#6441A4', // Twitch purple
                            data: data['max_twitch_viewers'],
                            yAxis: 1,
                            type: 'line',
                        }
                    ],
                }));

            },
        });
    }

    function loadAppPlayerTimes() {

        const options = {
            "order": [[3, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
                $(row).attr('data-link', data[6]);
            },
            "columnDefs": [
                // Rank
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return row[4];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('font-weight-bold')
                    },
                    "orderable": false,
                },
                // Flag
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        if (row[3]) {
                            return '<img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[7] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[7] + '" class="rounded">';
                        }
                        return '';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Player
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[5] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[2];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
            ]
        };

        $('#top-players-table').gdbTable({tableOptions: options});
    }
}
