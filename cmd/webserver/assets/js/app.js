const $appPage = $('#app-page');

if ($appPage.length > 0) {

    // Scroll to videos link
    $("#scroll-to-videos").on('mouseup', function (e) {
        const $videosDiv = $("#videos");
        $('html, body').animate({scrollTop: $videosDiv.offset().top - 15}, 500);

        const $videos = $videosDiv.find('video')
        if ($videos[0].paused) {
            $videos.first().trigger('click');
        }
    });

    // Micro video link
    $('#details video').on('mouseup', function (e) {
        const video = $(this)[0];
        if (video.paused) {
            video.play();
        } else {
            $('a.nav-link[href="#media"]').tab('show');
            $("#scroll-to-videos").trigger('mouseup');
        }
    });

    // Show dev raw row
    $('#dev-info').on('mouseup', 'tr', function () {

        const $tr = $(this);
        const row = $(this).closest('table').DataTable().row($tr);

        if (row.child.isShown()) {

            row.child.hide();
            $tr.removeClass('shown');

        } else {

            row.child(function () {
                return '<code class="wbba">' + $tr.data('raw') + '</code>';
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
                    loadPlayersTab();
                    break;
                case '#reviews':
                    loadAppReviewsChart();
                    break;
                case '#achievements':
                    loadAchievements();
                    break;
                case '#dlc':
                    loadDLC();
                    break;
                case '#dev-localization':
                    loadDevLocalization();
                    break;
                case '#media':
                    loadAppMediaTab();
                    break;
            }
        }

        // On leaving tab
        if (from.attr('href') === '#media') {
            pauseAllVideos();
        }
    });

    // Websockets
    websocketListener('app', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $appPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This app has been updated', 0, 'refresh');
        }
    });

    // News data table
    function loadNews() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
                $(row).addClass('cursor-pointer');
            },
            "columnDefs": [
                // Title
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[10] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '<br /><small>' + row[2] + '</small></div></div><div class="d-none">' + row[5] + '</div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('style', 'min-width: 300px;')
                        $(td).addClass('img');
                    },
                    "orderable": false
                },
                // Date
                {
                    "targets": 1,
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
        const searchFields = [
            $('#article-search'),
        ];

        const table = $newsTable.gdbTable({tableOptions: options, searchFields: searchFields});

        $newsTable.on('click', 'tr[role=row]', function () {

            const row = table.row($(this));

            // noinspection JSUnresolvedFunction
            if (row.child.isShown()) {

                row.child.hide();
                $(this).removeClass('shown');

            } else {

                row.child($("<div/>").html(row.data()[5])).show();
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

    function loadAppMediaTab() {

        $('#media #images img').each(function () {
            loadImage($(this));
        })

        $('#media #videos video').each(function () {
            loadVideo($(this));
        })
    }

    // News items
    function loadItems() {

        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
                $(row).addClass('cursor-pointer');
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
                // Icon / Article Name
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[25] + '" alt="" data-lazy-alt="' + row[16] + '"></div><div class="name">' + row[16] + '<br><small title="' + row[4] + '">' + row[29] + '</small></div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Link
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        if (row[28]) {
                            return '<a href="' + row[28] + '" data-src="/assets/img/no-app-image-square.jpg" target="_blank" rel="noopener" class="stop-prop"><i class="fas fa-link"></i></a>';
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

                    row.child('<img src="' + rowx[26] + '" alt="" class="float-right rounded" />' + html).show();
                    $(this).addClass('shown');
                }
            }
        );
    }

    function loadAppReviewsChart() {

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/reviews.html',
            dataType: 'html',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = '';
                }

                $('#reviews-ajax').html(data);
            },
        });

        const defaultAppChartOptions = {
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
            plotOptions: {},
            xAxis: {
                title: {text: ''},
                type: 'datetime'
            },
        };

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/reviews.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('reviews-chart', $.extend(true, {}, defaultAppChartOptions, {
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
                            // min: 0,
                        }
                    ],
                    // xAxis: {
                    //     gridLineWidth: 1,
                    // },
                    legend: {
                        enabled: true,
                        itemStyle: {
                            color: '#28a745',
                        },
                        itemHiddenStyle: {
                            color: '#666666',
                        },
                    },
                    tooltip: {
                        formatter: function () {

                            const time = moment(this.key).format("dddd DD MMM YYYY @ HH:mm");

                            if (this.series.name === 'Score') {
                                return this.y.toLocaleString() + '% Review score on ' + time;
                            } else if (this.series.name === 'Positive Reviews') {
                                return this.y.toLocaleString() + ' Positive reviews on ' + time;
                            } else if (this.series.name === 'Negative Reviews') {
                                return Math.abs(this.y).toLocaleString() + ' Negative reviews on ' + time;
                            }
                        },
                    },
                    series: [
                        {
                            type: 'line',
                            name: 'Score',
                            color: '#007bff',
                            data: data['mean_reviews_score'],
                            yAxis: 0,
                            marker: {symbol: 'circle'}
                        },
                        {
                            type: 'area',
                            name: 'Positive Reviews',
                            color: '#28a745',
                            data: data['mean_reviews_positive'],
                            yAxis: 1,
                            marker: {symbol: 'circle'}
                        },
                        {
                            type: 'area',
                            name: 'Negative Reviews',
                            color: '#e83e8c',
                            data: data['mean_reviews_negative'],
                            yAxis: 1,
                            marker: {symbol: 'circle'}
                        },
                    ],
                }));

            },
        });
    }

    function loadPlayersTab() {

        const config = {rootMargin: '50px 0px 50px 0px', threshold: 0};

        const playersCallback = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadAppPlayersChart();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(playersCallback, config).observe(document.getElementById("players-chart"));

        const youtubeCallback = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadAppYoutubeChart();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(youtubeCallback, config).observe(document.getElementById("youtube-chart"));

        const groupChart = document.getElementById("group-chart");
        if (groupChart) {
            const groupCallback = function (entries, self) {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        loadGroupChart($appPage);
                        self.unobserve(entry.target);
                    }
                });
            };
            new IntersectionObserver(groupCallback, config).observe(groupChart);
        }

        const timesCallback = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadAppPlayerTimes();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(timesCallback, config).observe(document.getElementById("top-players-table"));

        const wishlistCallback = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadAppWishlist();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(wishlistCallback, config).observe(document.getElementById("wishlists-chart"));
    }

    function loadAppPlayersChart() {

        const d = new Date();
        d.setDate(d.getDate() - 7);

        const defaultAppChartOptions = {
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
                enabled: false,
            },
            legend: {
                enabled: true,
                itemStyle: {
                    color: '#28a745',
                },
                itemHiddenStyle: {
                    color: '#666666',
                },
            },
            xAxis: {
                title: {text: ''},
                type: 'datetime',
            },
            yAxis: {
                allowDecimals: false,
                title: {text: ''},
                min: 0,
                opposite: false,
                labels: {
                    formatter: function () {
                        return this.value.toLocaleString();
                    },
                },
                visible: true,
            },
            plotOptions: {
                series: {
                    marker: {
                        enabled: false
                    }
                }
            },
            tooltip: {
                formatter: function () {
                    switch (this.series.name) {
                        case 'Players Online':
                            return this.y.toLocaleString() + ' players on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                        case 'Players Online (Average)':
                            return this.y.toLocaleString() + ' average players on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                        case 'Twitch Viewers':
                            return this.y.toLocaleString() + ' Twitch viewers on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                    }
                },
            },
        };

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/players.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    const now = Date.now();
                    data = {
                        "max_player_count": [[now, 0]],
                        "max_twitch_viewers": [[now, 0]],
                        "max_youtube_views": [[now, 0]],
                    };
                }

                Highcharts.chart('players-chart', $.extend(true, {}, defaultAppChartOptions, {
                    xAxis: {
                        min: d.getTime(),
                    },
                    series: [
                        {
                            name: 'Twitch Viewers',
                            color: '#6441A4', // Twitch purple
                            data: data['max_twitch_viewers'],
                            connectNulls: true,
                        },
                        {
                            name: 'Players Online (Average)',
                            color: '#28a74544',
                            data: data['max_moving_average'],
                            connectNulls: true,
                        },
                        {
                            name: 'Players Online',
                            color: '#28a745',
                            data: data['max_player_count'],
                            connectNulls: true,
                        },
                    ],
                }));

            },
        });

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/players2.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    const now = Date.now();
                    data = {
                        "max_twitch_viewers": [[now, 0]],
                        "max_moving_average": [[now, 0]],
                        "max_player_count": [[now, 0]],
                    };
                }

                Highcharts.chart('players-chart2', $.extend(true, {}, defaultAppChartOptions, {
                    series: [
                        {
                            name: 'Twitch Viewers',
                            color: '#6441A4', // Twitch purple
                            data: data['max_twitch_viewers'],
                            connectNulls: true,
                        },
                        {
                            name: 'Players Online (Average)',
                            color: '#28a74544',
                            data: data['max_moving_average'],
                            connectNulls: true,
                        },
                        {
                            name: 'Players Online',
                            color: '#28a745',
                            data: data['max_player_count'],
                            connectNulls: true,
                        },
                    ],
                }));

            },
        });
    }

    function loadAppYoutubeChart() {

        const d = new Date();
        d.setDate(d.getDate() - 7);

        const defaultAppChartOptions = {
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
                enabled: false,
            },
            legend: {
                enabled: true,
                itemStyle: {
                    color: '#28a745',
                },
                itemHiddenStyle: {
                    color: '#666666',
                },
            },
            xAxis: {
                title: {text: ''},
                type: 'datetime',
            },
            yAxis: [
                {
                    allowDecimals: false,
                    title: {text: 'Views'},
                    min: 0,
                    opposite: false,
                    labels: {
                        formatter: function () {
                            return this.value.toLocaleString();
                        },
                    },
                    visible: true,
                },
                {
                    allowDecimals: false,
                    title: {text: 'Comments'},
                    min: 0,
                    opposite: true,
                    labels: {
                        formatter: function () {
                            return this.value.toLocaleString();
                        },
                    },
                    visible: true,
                },
            ],
            plotOptions: {
                series: {
                    marker: {
                        enabled: false
                    }
                }
            },
            tooltip: {
                formatter: function () {
                    switch (this.series.name) {
                        case 'Youtube Views':
                            return this.y.toLocaleString() + ' Youtube views on ' + moment(this.key).format("dddd DD MMM YYYY");
                        case 'Youtube Comments':
                            return this.y.toLocaleString() + ' Youtube comments on ' + moment(this.key).format("dddd DD MMM YYYY");
                    }
                },
            },
        };

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/youtube.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    const now = Date.now();
                    data = {
                        "max_youtube_views": [[now, 0]],
                        "max_youtube_comments": [[now, 0]],
                    };
                }

                Highcharts.chart('youtube-chart', $.extend(true, {}, defaultAppChartOptions, {
                    xAxis: {
                        min: d.getTime(),
                    },
                    series: [
                        {
                            name: 'Youtube Comments',
                            color: '#007bff',
                            data: data['max_youtube_comments'],
                            connectNulls: true,
                            type: 'line',
                            step: 'right',
                            yAxis: 1,
                        },
                        {
                            name: 'Youtube Views',
                            color: '#28a745',
                            data: data['max_youtube_views'],
                            connectNulls: true,
                            type: 'line',
                            step: 'right',
                            yAxis: 0,
                        },
                    ],
                }));

            },
        });
    }

    function loadAppWishlist() {

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/wishlist.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = {};
                }

                Highcharts.chart('wishlists-chart', {
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
                        enabled: false,
                    },
                    legend: {
                        enabled: true,
                        itemStyle: {
                            color: '#28a745',
                        },
                        itemHiddenStyle: {
                            color: '#666666',
                        },
                    },
                    xAxis: {
                        title: {
                            text: ''
                        },
                        type: 'datetime'

                    },
                    yAxis: [
                        {
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
                            title: {
                                text: 'Wishlists'
                            },
                            labels: {
                                formatter: function () {
                                    return this.value.toFixed(7) + '%';
                                },
                            },
                        },
                        {
                            opposite: true,
                            reversed: true,
                            title: {
                                text: 'Average Position'
                            },
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        }
                    ],
                    tooltip: {
                        formatter: function () {
                            switch (this.series.name) {
                                case 'Wishlists':
                                    return 'In ' + this.y.toLocaleString() + ' wishlists on ' + moment(this.key).format("dddd DD MMM YYYY");
                                case 'Wishlists %':
                                    return 'In ' + this.y.toFixed(7) + '% of wishlists on ' + moment(this.key).format("dddd DD MMM YYYY");
                                case 'Average Position':
                                    return 'Average position of ' + this.y.toFixed(2).toLocaleString() + ' on ' + moment(this.key).format("dddd DD MMM YYYY");
                            }
                        },
                    },
                    series: [
                        {
                            name: 'Wishlists',
                            color: '#007bff',
                            data: data['mean_wishlist_count'],
                            marker: {symbol: 'circle'},
                            yAxis: 0,
                        },
                        {
                            name: 'Wishlists %',
                            color: '#28a745',
                            data: data['mean_wishlist_percent'],
                            marker: {symbol: 'circle'},
                            yAxis: 1,
                        },
                        {
                            name: 'Average Position',
                            color: '#e83e8c',
                            data: data['mean_wishlist_avg_position'],
                            marker: {symbol: 'circle'},
                            yAxis: 2,
                        },
                    ],
                });
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

    function loadAchievements() {

        const options = {
            "pageLength": 100,
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                if (data[7]) {
                    $(row).addClass('font-weight-bold');
                }
            },
            "columnDefs": [
                // Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        let name = row[0];
                        if (!row[4]) {
                            name += '<span class="badge badge-danger float-right ml-1">Inactive</span>';
                        }
                        if (row[5]) {
                            row[1] = '<em>&lt;Hidden&gt;</em> ' + row[1];
                        }
                        if (row[6]) {
                            name += '<span class="badge badge-danger float-right ml-1">Deleted</span>';
                        }

                        name += '<br><small>' + row[1] + '</small>';

                        return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[0] + '"></div><div class="name">' + name + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Description
                // {
                //     "targets": 1,
                //     "render": function (data, type, row) {
                //         return row[1];
                //     },
                //     "orderable": false,
                // },
                // Completed
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[3] + '%';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        rowData[3] = Math.ceil(rowData[3]);
                        $(td).css('background', 'linear-gradient(to right, rgba(0,0,0,.15) ' + rowData[3] + '%, transparent ' + rowData[3] + '%)');
                        $(td).addClass('thin');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
            ]
        };

        $('#achievements-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadDLC() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[5]);
            },
            "columnDefs": [
                // Name
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
                // Release Date
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        if (row[3]) {
                            return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                        } else {
                            return row[4];
                        }
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
            ]
        };

        const searchFields = [
            $('#dlc-search'),
        ];

        $('#dlc-table').gdbTable({
            tableOptions: options,
            searchFields: searchFields
        });
    }

    function loadDevLocalization() {

        $.ajax({
            type: "GET",
            url: '/games/' + $appPage.attr('data-id') + '/localization.html',
            dataType: 'html',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = '';
                }

                $('#dev-localization').html(data);
            },
        });

    }
}
