const $appPage = $('#app-page');

if ($appPage.length > 0) {

    // Scroll to videos link
    $('#scroll-to-videos').on('mouseup', function (e) {

        const $video = $('#videos video').first();
        const offset = ($(window).height() / 2) - ($video.height() / 2);

        $('html, body').animate({scrollTop: $video.offset().top - offset}, 500);

        if ($video[0].paused) {
            $video.trigger('click');
        }
    });

    // Micro video link
    $('#details video').on('mouseup', function (e) {
        const video = $(this)[0];
        if (video.paused) {
            video.play();
        } else {
            $('a.nav-link[href="#media"]').tab('show');
            $('#scroll-to-videos').trigger('mouseup');
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

    loadAjaxOnObserve({
        'news': loadNews,
        'news-small': loadNewsSmall,
        'items': loadItems,
        'prices': loadPriceChart,
        'similar-owners': loadAppSimilar,
        'similar-owners-small': loadAppSimilarSmall,
        'reviews': loadAppReviewsChart,
        'achievements': loadAchievements,
        'dlc': loadDLC,
        'dev-localization': loadDevLocalization,
        'media': loadAppMediaTab,
        'tags-chart': loadAppTags,

        // Packages tab
        'bundles-table': loadAppBundlesTab,
        'packages-table': loadAppPackagesTab,

        // Players tab
        'players-chart': loadAppPlayersChart,
        'players-heatmap-chart': loadAppPlayersHeatmapChart,
        'group-chart': function () {
            loadGroupChart($appPage);
        },
        'top-players-table': loadAppPlayerTimes,
        'wishlists-chart': loadAppWishlist,
    });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
        if ($(e.relatedTarget).attr('href') === '#media') {
            pauseAllVideos();
        }
    });

    // Websockets
    if (user.toasts && user.toasts.length > 0) { // Only wait for update if an update was queued
        websocketListener('app', function (e) {

            const data = JSON.parse(e.data);
            if (data.Data.toString() === $appPage.attr('data-id')) {
                toast(true, 'Click to refresh', 'This app has been updated', 0, 'refresh');
            }
        });
    }

    // News data table
    function loadNews() {

        // Feed drop down
        const $feed = $('#article-feed');

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/news-feeds.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                $.each(data, function (i, item) {
                    $feed.append($('<option>', {
                        value: item.id,
                        text: item.name + ' (' + item.count + ')',
                    }));
                });

                if (!window.location.hash.includes(',')) {
                    $feed.val('steam_community_announcements');
                }

                $feed.chosen({
                    disable_search_threshold: 5,
                    allow_single_deselect: true,
                    max_selected_options: 10,
                });

                // Table
                const $newsTable = $('#news-table');
                const table = $newsTable.gdbTable({
                    searchFields: [
                        $('#article-search'),
                        $feed,
                    ],
                    tableOptions: {
                        'order': [[1, 'desc']],
                        'createdRow': function (row, data, dataIndex) {
                            // $(row).attr('data-id', data[0]);
                            $(row).attr('data-article-id', data[0]);
                            $(row).addClass('cursor-pointer');
                        },
                        'columnDefs': [
                            // Title
                            {
                                'targets': 0,
                                'render': function (data, type, row) {
                                    return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[10] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '<br /><small>' + row[2] + '</small></div></div>'
                                        + '<div class="d-none">' + row[5] + '</div>';
                                },
                                'createdCell': function (td, cellData, rowData, row, col) {
                                    $(td).attr('style', 'min-width: 300px;');
                                    $(td).addClass('img');
                                },
                                'orderable': false,
                            },
                            // Date
                            {
                                'targets': 1,
                                'render': function (data, type, row) {
                                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                                },
                                'createdCell': function (td, cellData, rowData, row, col) {
                                    $(td).attr('nowrap', 'nowrap');
                                },
                                'orderable': false,
                            },
                        ],
                    },
                });

                $newsTable.on('click', 'tbody tr[role=row]', function () {

                    const row = table.row($(this));

                    // noinspection JSUnresolvedFunction
                    if (row.child.isShown()) {

                        row.child.hide();
                        $(this).removeClass('shown');
                        history.replaceState(null, null, '#news');

                    } else {

                        row.child($('<div/>').html(row.data()[5])).show();
                        $(this).addClass('shown');
                        observeLazyImages($(this).next().find('img[data-lazy]'));
                        history.replaceState(null, null, '#news,' + $(this).attr('data-article-id'));
                    }
                });

                $newsTable.on('draw.dt', function (e, settings) {

                    const hash = window.location.hash;
                    if (hash) {
                        const id = hash.match(/\d{19}/)[0];
                        if (id) {
                            const $tr = $('tr[data-article-id=' + id + ']');
                            if ($tr.length > 0) {
                                $tr.trigger('click');
                                $('html, body').animate({scrollTop: $tr.offset().top - 100}, 200);
                            }
                        }
                    }
                });
            },
        });
    }

    function loadAppMediaTab() {

        $('#media #images img').each(function () {
            loadImage($(this));
        });

        $('#media #videos video').each(function () {
            loadVideo($(this));
        });
    }

    // News items
    function loadItems() {

        const options = {
            'order': [[2, 'desc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
                $(row).addClass('cursor-pointer');
            },
            'columnDefs': [
                // Description
                {
                    'targets': 0,
                    'render': function (data, type, row) {
                        return row[12];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                    },
                    'orderable': false,
                },
                // Icon / Article Name
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[25] + '" alt="" data-lazy-alt="' + row[16] + '"></div><div class="name">' + row[16] + '<br><small title="' + row[4] + '">' + row[29] + '</small></div></div>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderable': false,
                },
                // Link
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        if (row[28]) {
                            return '<a href="' + row[28] + '" data-src="/assets/img/no-app-image-square.jpg" target="_blank" rel="noopener" class="stop-prop"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    'orderable': false,
                },
            ],
        };

        const $itemsTable = $('#items-table');

        const searchFields = [
            $('#items-search'),
        ];

        const table = $itemsTable.gdbTable({tableOptions: options, searchFields: searchFields});

        $itemsTable.on('click', 'tbody tr[role=row]', function () {

                const row = table.row($(this));

                // noinspection JSUnresolvedFunction
                if (row.child.isShown()) {

                    row.child.hide();
                    $(this).removeClass('shown');

                } else {

                    const rowx = row.data();

                    const fields = [
                        {Name: 'App ID', Value: rowx[0]},
                        {Name: 'Bundle', Value: rowx[1]},
                        {Name: 'Commodity', Value: rowx[2]},
                        {Name: 'Date Created', Value: rowx[3]},
                        {Name: 'Description', Value: rowx[4]},
                        {Name: 'Display Type', Value: rowx[5]},
                        {Name: 'Drop Interval', Value: rowx[6]},
                        {Name: 'Drop Max Per Window', Value: rowx[7]},
                        {Name: 'Exchange', Value: rowx[8]},
                        {Name: 'Hash', Value: rowx[9]},
                        {Name: 'Icon URL', Value: '<a href="' + rowx[10] + '" target="_blank" rel="noopener">' + rowx[10] + '</a>'},
                        {Name: 'Icon URL Large', Value: '<a href="' + rowx[11] + '" target="_blank" rel="noopener">' + rowx[11] + '</a>'},
                        {Name: 'Item Def ID', Value: rowx[12]},
                        {Name: 'Item Quality', Value: rowx[13]},
                        {Name: 'Marketable', Value: rowx[14]},
                        {Name: 'Modified', Value: rowx[15]},
                        {Name: 'Name', Value: rowx[16]},
                        {Name: 'Price', Value: rowx[17]},
                        {Name: 'Promo', Value: rowx[18]},
                        {Name: 'Quantity', Value: rowx[19]},
                        {Name: 'Tags', Value: rowx[20]},
                        {Name: 'Timestamp', Value: rowx[21]},
                        {Name: 'Tradable', Value: rowx[22]},
                        {Name: 'Type', Value: rowx[23]},
                        {Name: 'Workshop ID', Value: rowx[24]},
                    ];

                    const html = json2html.transform(fields, {
                        '<>': 'tr', 'class': '', 'html': [
                            {'<>': 'th', 'html': '${Name}', 'class': 'nowrap'},
                            {'<>': 'td', 'html': '${Value}'},
                        ],
                    });

                    row.child('<table class="table table-hover table-striped table-sm mb-0">' + html + '</table>').show();
                    $(this).addClass('shown');
                }
            },
        );
    }

    function loadAppSimilarSmall() {

        const dt = $('#similar-owners-small').gdbTable({
            tableOptions: {
                'language': {
                    'zeroRecords': function () {
                        return 'This app is yet to be scanned';
                    },
                },
                'createdRow': function (row, data, dataIndex) {
                    $(row).attr('data-app-id', data[0]);
                    $(row).attr('data-link', data[1]);
                },
                'columnDefs': [
                    // App
                    {
                        'targets': 0,
                        'render': function (data, type, row) {
                            return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + row[3] + '</div></div>';
                        },
                        'createdCell': function (td, cellData, rowData, row, col) {
                            $(td).addClass('img');
                        },
                        'orderable': false,
                    },
                ],
            },
        });

        dt.on('draw.dt', function (e, settings) {
            if (!dt.page.info().recordsTotal) {
                $('#same-owners-small-wrapper div.card-footer').remove();
            }
        });
    }

    function loadNewsSmall() {

        const $table = $('#news-small');
        const dt = $table.gdbTable({
            tableOptions: {
                'language': {
                    'zeroRecords': function () {
                        return 'No community news yet';
                    },
                },
                'createdRow': function (row, data, dataIndex) {
                    $(row).attr('data-article-id', data[0]);
                    $(row).addClass('cursor-pointer');
                },
                'columnDefs': [
                    // Title
                    {
                        'targets': 0,
                        'render': function (data, type, row) {
                            return '<div class="icon-name">' + '<div class="name">' + row[1] + '<br /><small><span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span></small></div></div>';
                        },
                        'createdCell': function (td, cellData, rowData, row, col) {
                            $(td).attr('style', 'min-width: 300px;');
                            $(td).addClass('img');
                        },
                        'orderable': false,
                    },
                ],
            },
        });

        $table.on('click', 'tbody tr[role=row]', function () {

            history.pushState(null, null, '#news,' + $(this).attr('data-article-id'));
            $('a.nav-link[href="#news"]').tab('show');
        });
    }

    function loadAppSimilar() {

        $('#similar-owners').gdbTable({
            order: [[3, 'desc']],
            tableOptions: {
                'serverSide': false,
                'language': {
                    'zeroRecords': function () {
                        return 'This app is yet to be scanned';
                    },
                },
                'createdRow': function (row, data, dataIndex) {
                    $(row).attr('data-app-id', data[0]);
                    $(row).attr('data-link', data[1]);
                },
                'columnDefs': [
                    // App
                    {
                        'targets': 0,
                        'render': function (data, type, row) {
                            return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + row[3] + '</div></div>';
                        },
                        'createdCell': function (td, cellData, rowData, row, col) {
                            $(td).addClass('img');
                        },
                        'orderSequence': ['asc', 'desc'],
                    },
                    // Same Owners
                    {
                        'targets': 1,
                        'render': function (data, type, row) {
                            return row[5].toLocaleString();
                        },
                        'orderSequence': ['desc', 'asc'],
                    },
                    // Total Owners
                    {
                        'targets': 2,
                        'render': function (data, type, row) {
                            return row[6].toLocaleString();
                        },
                        'orderSequence': ['desc', 'asc'],
                    },
                    // Score
                    {
                        'targets': 3,
                        'render': function (data, type, row) {
                            return row[7].toFixed(2).toLocaleString();
                        },
                        'orderSequence': ['desc', 'asc'],
                    },
                    // External Link
                    {
                        'targets': 4,
                        'render': function (data, type, row) {
                            if (row[4]) {
                                return '<a href="' + row[4] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                            }
                            return '';
                        },
                        'orderable': false,
                    },
                ],
            },
        });

        const $tagsWrapper = $('#similar-tags');
        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/similar.html',
            dataType: 'html',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = '';
                }

                $tagsWrapper.html(data);

                observeLazyImages($tagsWrapper.find('img[data-lazy]'));
            },
        });
    }

    function loadAppReviewsChart() {

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/reviews.html',
            dataType: 'html',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = '';
                }

                $('#reviews-ajax').html(data);
            },
        });

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/reviews.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('reviews-chart', $.extend(true, {}, defaultChartOptions, {
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
                                },
                            },
                        },
                        {
                            allowDecimals: false,
                            title: {text: ''},
                            opposite: true,
                            // min: 0,
                        },
                    ],
                    tooltip: {
                        formatter: function () {

                            const time = moment(this.key).format('dddd DD MMM YYYY @ HH:mm');

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
                            marker: {symbol: 'circle'},
                        },
                        {
                            type: 'area',
                            name: 'Positive Reviews',
                            color: '#28a745',
                            data: data['mean_reviews_positive'],
                            yAxis: 1,
                            marker: {symbol: 'circle'},
                        },
                        {
                            type: 'area',
                            name: 'Negative Reviews',
                            color: '#e83e8c',
                            data: data['mean_reviews_negative'],
                            yAxis: 1,
                            marker: {symbol: 'circle'},
                        },
                    ],
                }));

            },
        });
    }

    function loadAppPlayersChart() {

        const d = new Date();
        d.setDate(d.getDate() - 7);

        const chartOptions = $.extend(true, {}, defaultChartOptions, {
            yAxis: [
                {
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
                    visible: true,
                },
            ],
            xAxis: {
                plotLines: plotline($appPage.attr('data-release'), 'Steam Release'),
            },
            plotOptions: {
                series: {
                    marker: {
                        enabled: false,
                    },
                },
            },
            tooltip: {
                formatter: function () {
                    switch (this.series.name) {
                        case 'Players Online':
                            return this.y.toLocaleString() + ' players on ' + moment(this.key).format('dddd DD MMM YYYY @ HH:mm');
                        case 'Players Online (Average)':
                            return this.y.toLocaleString() + ' average players on ' + moment(this.key).format('dddd DD MMM YYYY @ HH:mm');
                        case 'Twitch Viewers':
                            return this.y.toLocaleString() + ' Twitch viewers on ' + moment(this.key).format('dddd DD MMM YYYY @ HH:mm');
                        case 'YouTube Views':
                            return this.y.toLocaleString() + ' YouTube views on ' + moment(this.key).format('dddd DD MMM YYYY');
                        case 'YouTube Comments':
                            return this.y.toLocaleString() + ' YouTube comments on ' + moment(this.key).format('dddd DD MMM YYYY');
                    }
                },
            },
        });

        const series = function (data) {

            let series = [
                // {
                //     name: 'Twitch Viewers',
                //     color: '#6441A4', // Twitch purple
                //     data: data['max_twitch_viewers'],
                //     connectNulls: true,
                //     yAxis: 1,
                // },
                {
                    name: 'Players Online (Average)',
                    color: '#28a74544',
                    data: data['max_moving_average'],
                    connectNulls: true,
                    yAxis: 0,
                },
                {
                    name: 'Players Online',
                    color: '#28a745',
                    data: data['max_player_count'],
                    connectNulls: true,
                    yAxis: 0,
                },
            ];

            if (user.isLoggedIn) {
                series.unshift(
                    {
                        name: 'YouTube Comments',
                        color: '#ff0000',
                        data: data['max_youtube_comments'],
                        connectNulls: true,
                        type: 'line',
                        step: 'right',
                        visible: false,
                        yAxis: 1,
                    },
                    {
                        name: 'YouTube Views',
                        color: '#ff0000',
                        data: data['max_youtube_views'],
                        connectNulls: true,
                        type: 'line',
                        step: 'right',
                        visible: false,
                        yAxis: 1,
                    },
                );
            }

            return series;
        };

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/players.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                const $chart = $('#players-chart');

                if (!data || Object.keys(data).length === 0) {
                    $chart.html('No players.').css('height', 'auto');
                    return;
                }

                const start = d.getTime();

                let max = 0;
                if (isIterable(data['max_youtube_views'])) {
                    data['max_youtube_views'].forEach(function myFunction(value, index, array) {
                        if (value[0] > start && value[1] != null && value[1] > max) {
                            max = value[1];
                        }
                    });
                }
                $('#youtube-max-views').html(max.toLocaleString());

                //
                max = 0;
                if (isIterable(data['max_youtube_comments'])) {
                    data['max_youtube_comments'].forEach(function myFunction(value, index, array) {
                        if (value[0] > start && value[1] != null && value[1] > max) {
                            max = value[1];
                        }
                    });
                }
                $('#youtube-max-comments').html(max.toLocaleString());

                Highcharts.chart($chart[0], $.extend(true, {}, chartOptions, {
                    xAxis: {
                        min: d.getTime(),
                    },
                    series: series(data),
                }));
            },
        });

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/players2.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                const $chart = $('#players-chart2');

                if (!data || Object.keys(data).length === 0) {
                    $chart.html('No players.').css('height', 'auto');
                    return;
                }

                Highcharts.chart($chart[0], $.extend(true, {}, chartOptions, {
                    chart: {
                        zoomType: 'x',
                    },
                    series: series(data),
                }));
            },
        });
    }

    function loadAppPlayersHeatmapChart() {

        // Set default
        const $dd = $('#heatmap-timezones');
        let diff = Math.floor(moment().utcOffset() / 60);
        if (diff >= 0) {
            diff = '+' + diff;
        }
        $dd.val(diff);

        //
        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/players-heatmap.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                const $chart = $('#players-heatmap-chart');

                if (!data || Object.keys(data).length === 0) {
                    $chart.html('No players.').css('height', 'auto');
                    return;
                }

                let dt;
                $dd.on('change', function () {

                    if (dt) {
                        // Having to rebuild whole chart as just updating series data seems to break it
                        dt.destroy();
                    }

                    dt = Highcharts.chart($chart[0], $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            type: 'heatmap',
                        },
                        xAxis: {
                            title: null,
                            type: 'category',
                            lineColor: 'rgba(0,0,0,0)',
                        },
                        yAxis: {
                            categories: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],
                            title: null,
                            reversed: true,
                            gridLineColor: 'rgba(0,0,0,0)',
                        },
                        legend: {
                            enabled: true,
                            align: 'right',
                            layout: 'vertical',
                            verticalAlign: 'middle',
                        },
                        tooltip: {
                            formatter: function () {
                                const day = this.series.yAxis.categories[this.point.y];
                                const time = this.point.x;
                                return 'Average of last 4 ' + day + 's @ ' + pad(time, 2) + ':00-' + pad(time, 2) + ':59 ' + diff + ': ~'
                                    + Math.round(this.point.value).toLocaleString() + ' players';
                            },
                        },
                        colorAxis: {
                            minColor: darkMode ? '#212529' : '#FFFFFF',
                            maxColor: defaultChartOptions.colors[0],
                            reversed: false,
                        },
                        plotOptions: {
                            series: {
                                marker: {
                                    enabled: true,
                                },
                            },
                        },
                        series: [{
                            data: adjustZone(data.max_player_count, $dd.val()),
                            borderWidth: 0,
                        }],
                    }));
                });

                $dd.trigger('change');
            },
        });
    }

    function adjustZone(data, diff) {

        // Using extend to copy by value
        const data2 = $.extend(true, [], data);

        data2.forEach(function (hour, index) {

            diff = parseInt(diff);

            if (diff !== 0) {

                data2[index][0] += diff;

                // Hours
                if (data2[index][0] > 23) {
                    data2[index][0] -= 24;
                    data2[index][1]++;
                } else if (data2[index][0] < 0) {
                    data2[index][0] += 24;
                    data2[index][1]--;
                }

                // Days
                if (data2[index][1] > 6) {
                    data2[index][1] -= 7;
                } else if (data2[index][1] < 0) {
                    data2[index][1] += 7;
                }
            }
        });

        return data2;
    }

    function loadAppWishlist() {

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/wishlist.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = {};
                }

                Highcharts.chart('wishlists-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: [
                        {
                            title: {
                                text: '',
                            },
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        },
                        {
                            title: {
                                text: 'Wishlists',
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
                                text: 'Average Position',
                            },
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        },
                    ],
                    xAxis: {
                        plotLines: plotline($appPage.attr('data-release'), 'Steam Release'),
                    },
                    tooltip: {
                        formatter: function () {
                            switch (this.series.name) {
                                case 'Wishlists':
                                    return 'In ' + this.y.toLocaleString() + ' wishlists on ' + moment(this.key).format('dddd DD MMM YYYY');
                                case 'Wishlists %':
                                    return 'In ' + this.y.toFixed(7) + '% of wishlists on ' + moment(this.key).format('dddd DD MMM YYYY');
                                case 'Average Position':
                                    return 'Average position of ' + this.y.toFixed(2).toLocaleString() + ' on ' + moment(this.key).format('dddd DD MMM YYYY');
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
                }));
            },
        });
    }

    function loadAppPlayerTimes() {

        const options = {
            'order': [[3, 'desc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
                $(row).attr('data-link', data[6]);
            },
            'columnDefs': [
                // Rank
                {
                    'targets': 0,
                    'render': function (data, type, row) {
                        return row[4];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('font-weight-bold');
                    },
                    'orderable': false,
                },
                // Flag
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        if (row[3]) {
                            return '<img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[7] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[7] + '" class="rounded">';
                        }
                        return '';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderable': false,
                },
                // Player
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        return '<a href="' + row[6] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[5] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderable': false,
                },
                // Time
                {
                    'targets': 3,
                    'render': function (data, type, row) {
                        return row[2];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderable': false,
                },
            ],
        };

        $('#top-players-table').gdbTable({tableOptions: options});
    }

    function loadAchievements() {

        const appID = $appPage.attr('data-id');

        // Setup drop downs
        $('select#achievements-filter').chosen({
            disable_search_threshold: 5,
            allow_single_deselect: false,
            max_selected_options: 1,
        });

        loadFriends(appID, false, function ($chosen) {
            const val = $chosen.val();
            if (val) {
                window.location.href = '/games/' + appID + '/compare-achievements/' + user.playerID + ',' + val;
            }
        });

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/achievement-counts.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = {};
                }
                if (!data.hasOwnProperty('data')) {
                    data.data = [];
                }
                if (!data.data) {
                    $('#achievement-counts-chart').css('height', 'auto').html('No data');
                    return;
                }

                Highcharts.chart('achievement-counts-chart', $.extend(true, {}, defaultChartOptions, {
                    legend: {
                        enabled: false,
                    },
                    yAxis: {
                        // type: 'logarithmic',
                        allowDecimals: false,
                        title: {
                            text: '',
                        },
                        min: 0,
                    },
                    xAxis: {
                        type: 'category',
                        plotLines: plotline(data.marker, 'You are here!'),
                        min: data.marker === 0 ? 0 : 1,
                    },
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' players have ' + this.x.toLocaleString() + ' achievements';
                        },
                    },
                    series: [{
                        data: data.data,
                    }],

                }));
            },
        });

        $('#achievements-table').gdbTable({
            searchFields: [
                $('#achievements-filter'),
            ],
            tableOptions: {
                'pageLength': 1000,
                'order': [[2, 'desc']],
                'createdRow': function (row, data, dataIndex) {
                    if (user.isLocal) {
                        $(row).attr('data-ach-id', data[9]);
                    }
                },
                'columnDefs': [
                    // Name
                    {
                        'targets': 0,
                        'render': function (data, type, row) {

                            let name = row[0];

                            if (!name) {
                                name = row[9];
                            }

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

                            return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[0] + '"></div><div class="name">' + name + '</div></div>';
                        },
                        'createdCell': function (td, cellData, rowData, row, col) {
                            $(td).addClass('img');
                        },
                        'orderable': false,
                    },
                    // Completed Time
                    {
                        'targets': 1,
                        'render': function (data, type, row) {
                            if (row[7] && row[7] > 0) {
                                return '<span data-livestamp="' + row[7] + '"></span>'
                                    + '<br><small class="text-muted">' + row[8] + '</small>';
                            }
                            if (row[7] === -1) {
                                return 'Unknown';
                            }
                            return '';
                        },
                        'createdCell': function (td, cellData, rowData, row, col) {
                            $(td).attr('nowrap', 'nowrap');
                        },
                        'orderable': false,
                    },
                    // Completed Percent
                    {
                        'targets': 2,
                        'render': function (data, type, row) {
                            return row[3] + '%';
                        },
                        'createdCell': function (td, cellData, rowData, row, col) {
                            rowData[3] = Math.ceil(rowData[3]);
                            $(td).css('background', 'linear-gradient(to right, rgba(0,0,0,.15) ' + rowData[3] + '%, transparent ' + rowData[3] + '%)');
                            $(td).addClass('thin');
                        },
                        'orderSequence': ['desc', 'asc'],
                    },
                ],
            },
        });
    }

    function loadDLC() {

        const options = {
            'order': [[1, 'desc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[5]);
            },
            'columnDefs': [
                // Name
                {
                    'targets': 0,
                    'render': function (data, type, row) {
                        return '<a href="' + row[5] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Release Date
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        if (row[3]) {
                            return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                        } else {
                            return row[4];
                        }
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
            ],
        };

        const searchFields = [
            $('#dlc-search'),
        ];

        $('#dlc-table').gdbTable({
            tableOptions: options,
            searchFields: searchFields,
        });
    }

    function loadAppBundlesTab() {

        const options = {
            'serverSide': false,
            'order': [[0, 'asc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-link', data[1]);
            },
            'columnDefs': [
                // Name
                {
                    'targets': 0,
                    'render': function (data, type, row) {
                        return '<a href="' + row[1] + '" class="icon-name"><div class="icon"><img data-lazy="/assets/img/no-app-image-square.jpg" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></a>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Discount
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        return row[3] + '%';
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Apps Count
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Packages Count
                {
                    'targets': 3,
                    'render': function (data, type, row) {
                        return row[5].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Updated At
                {
                    'targets': 4,
                    'render': function (data, type, row) {
                        return row[6];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
            ],
        };

        $('#bundles-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadAppPackagesTab() {

        const options = {
            'serverSide': false,
            'order': [[0, 'asc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-link', data[1]);
            },
            'columnDefs': [
                // Name
                {
                    'targets': 0,
                    'render': function (data, type, row) {
                        return '<a href="' + row[1] + '" class="icon-name"><div class="icon"><img data-lazy="/assets/img/no-app-image-square.jpg" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></a>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Billing Type
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        return row[3];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // License Type
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        return row[4];
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Status
                {
                    'targets': 3,
                    'render': function (data, type, row) {
                        return row[5];
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Apps Count
                {
                    'targets': 4,
                    'render': function (data, type, row) {
                        return row[6];
                    },
                    'orderSequence': ['desc', 'asc'],
                },
            ],
        };

        $('#packages-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadDevLocalization() {

        $.ajax({
            type: 'GET',
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

    function loadAppTags() {

        $.ajax({
            type: 'GET',
            url: '/games/' + $appPage.attr('data-id') + '/tags.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                if (isIterable(data['order'])) {
                    for (const id of data['order']) {
                        series.push({
                            name: data['names'][id],
                            data: data['counts']['tag_' + id.toString()],
                        });
                    }
                }

                Highcharts.chart('tags-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: {
                        title: {
                            text: 'Votes',
                        },
                    },
                    series: series,
                }));
            },
        });
    }
}
