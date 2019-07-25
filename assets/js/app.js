const $appPage = $('#app-page');

if ($appPage.length > 0) {

    const $modal = $('#news-modal');

    // Detials image click
    const $detailsImage = $('#details img');

    $detailsImage.on('click', function () {
        $('.card-header-tabs a[href="#media"]').tab('show');
    });
    $detailsImage.on("error", function () {
        $(this).attr('src', '/assets/img/no-app-image-banner.jpg');
        $(this).hide();
    });

    // Show dev raw row
    $('#dev table.table tbody').on('click', 'td i, td svg', function () {

        const table = $(this).closest('table').DataTable()
        const $tr = $(this).closest('tr');
        const row = table.row($tr);

        if (row.child.isShown()) {

            row.child.hide();
            $tr.removeClass('shown');

        } else {

            row.child(function () {
                console.log($tr.data('raw'));
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
        if (to.attr('href') === '#media') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadMedia();
            }
        }
        if (to.attr('href') === '#news') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadNews();
            }
        }
        if (to.attr('href') === '#prices') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadPriceChart();
            }
        }
        if (to.attr('href') === '#players') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadAppPlayersChart();
                loadAppPlayerTimes();
            }
        }
        if (to.attr('href') === '#reviews') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadAppReviewsChart();
            }
        }

        // On leaving tab
        if (from.attr('href') === '#media') {
            resetVideos();
        }

        // On any tab
        $.each(dataTables, function (index, value) {
            // noinspection JSUnresolvedFunction
            value.fixedHeader.adjust();
        });
    });

    function resetVideos() {
        $('video').each(function (index) {
            $(this)[0].pause();
            $(this)[0].currentTime = 0;
        });
    }

    // Websockets
    websocketListener('app', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $appPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This app has been updated', -1, 'refresh');
        }
    });

    // Media carousels
    function loadMedia() {

        $('#carousel1 img, #carousel2 img').each(function (index) {
            loadImage($(this));
        });

        const $carousel1 = $('#carousel1');
        const $carousel2 = $('#carousel2');

        // noinspection JSUnresolvedFunction
        $carousel1.slick({
            waitForAnimate: false,
            arrows: false,
            autoplay: false,
            dots: false,
            asNavFor: $carousel2,
            adaptiveHeight: true,
            lazyLoad: 'ondemand',
        });

        // noinspection JSUnresolvedFunction
        $carousel2.slick({
            waitForAnimate: false,
            arrows: false,
            slidesToShow: 15,
            autoplay: false,
            dots: false,
            variableWidth: true,
            asNavFor: $carousel1,
            focusOnSelect: true,
            centerMode: true,
            infinite: true,
        });

        $carousel1.on('afterChange', function (event, slick, currentSlide) {

            // Stop all videos
            resetVideos();

            // Auto play current video
            const $video = $carousel1.find('div[data-slick-index=' + currentSlide + '] video');
            if ($video.length > 0) {
                $video[0].play();
            }
        });

        $(document).on('keydown', function (e) {
            if ($('a.active[href="#media"]').length > 0) {
                if (e.keyCode === 37) {
                    // noinspection JSUnresolvedFunction
                    $carousel1.slick('slickPrev');
                }
                if (e.keyCode === 39) {
                    // noinspection JSUnresolvedFunction
                    $carousel1.slick('slickNext');
                }
            }
        });

        // Fix layout when images lazy load
        $('#carousel1 img').on('load', function () {
            $carousel1.slick('setPosition');
            $carousel2.slick('setPosition');
        });
        $('#carousel2 img').on('load', function () {
            $carousel1.slick('setPosition');
            $carousel2.slick('setPosition');
        });
    }

    // News data table
    function loadNews() {

        const $newstable = $('#news-table');

        const table = $newstable.DataTable($.extend(true, {}, dtDefaultOptions, {
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
                    "orderable": false
                },
                // Author
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[2];
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
                }
            ]
        }));

        dataTables.push(table);

        $newstable.on('click', 'tr[role=row]', function () {

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

        const table = $('#top-players-table').DataTable($.extend(true, {}, dtDefaultOptions, {
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
                            return '<img data-lazy="' + row[3] + '" data-lazy-alt="' + row[7] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[7] + '" class="rounded">';
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
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[5] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
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
        }));

        dataTables.push(table);
    }
}
