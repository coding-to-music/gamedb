const $appsComparePage = $('#apps-compare-page');

if ($appsComparePage.length > 0) {

    const config = {rootMargin: '50px 0px 50px 0px', threshold: 0};

    const callback1 = function (entries, self) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                loadCompareSearchTable();
                self.unobserve(entry.target);
            }
        });
    };
    new IntersectionObserver(callback1, config).observe(document.getElementById('apps-table'));

    const playersChart = document.getElementById('players-chart');
    if (playersChart) {
        const callback2 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadComparePlayersChart();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback2, config).observe(playersChart);
    }

    const groupsChart = document.getElementById('group-chart');
    if (groupsChart) {
        const callback3 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadCompareFollowersChart();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback3, config).observe(groupsChart);
    }

    const scoreChart = document.getElementById('score-chart');
    if (scoreChart) {
        const callback4 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadCompareScoreChart();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback4, config).observe(scoreChart);
    }

    const wishlistChart = document.getElementById('wishlists-chart');
    if (wishlistChart) {
        const callback5 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    loadCompareWishlistChart();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback5, config).observe(wishlistChart);
    }

    function loadCompareSearchTable() {

        const options = {
            "order": [[0, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[1]);
                $(row).attr('data-link', data[4]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
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
                        return row[6];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Action
                {
                    "targets": 2,
                    "render": function (data, type, row) {

                        if (row[8]) {
                            return '<a href="' + row[7] + '" ><i class="fas fa-minus"></i> Remove</a>';
                        } else {
                            return '<a href="' + row[7] + '" ><i class="fas fa-plus"></i> Add</a>';
                        }
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Community Link
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[5]) {
                            return '<a href="' + row[5] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
                // Search Score
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[9].toLocaleString();
                    },
                    "orderable": false,
                    "visible": false,
                },
            ]
        };

        const $ids = $('#ids');
        const $search = $('#search');

        const dt = $('#search-table').gdbTable({
            tableOptions: options,
            searchFields: [$ids, $search],
        });

        dt.on('draw.dt', function (e, settings) {
            if ($search.val()) {
                $('#search-results').show();
            } else {
                $('#search-results').hide();
            }
        });

        $('#apps-table').gdbTable({
            tableOptions: options,
            searchFields: [$ids],
        });
    }

    function loadComparePlayersChart() {

        if ($.isEmptyObject(appNames)) {
            return;
        }

        const chartOptions = $.extend(true, {}, defaultChartOptions, {
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
                    },
                }
            },
            tooltip: {
                formatter: function () {
                    return this.series.name + ' had ' + this.y.toLocaleString() + ' players on '
                        + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                },
            },
        });

        $.ajax({
            type: "GET",
            url: '/games/compare/' + $appsComparePage.attr('data-id') + '/players.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['max_player_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('players-chart', $.extend(true, {}, chartOptions, {
                    series: series,
                }));
            },
        });

        $.ajax({
            type: "GET",
            url: '/games/compare/' + $appsComparePage.attr('data-id') + '/players2.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['max_player_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('players-chart2', $.extend(true, {}, chartOptions, {
                    series: series,
                }));

            },
        });
    }

    function loadCompareFollowersChart($page = null) {

        if ($.isEmptyObject(groupNames)) {
            return;
        }

        $.ajax({
            type: "GET",
            url: '/games/compare/' + $appsComparePage.attr('data-group-id') + '/members.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: groupNames[datum.key],
                        data: datum['value']['max_members_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('group-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: {
                        allowDecimals: false,
                        title: {
                            text: ''
                        },
                        labels: {
                            formatter: function () {
                                return this.value.toLocaleString();
                            },
                        },
                        min: 0,
                    },
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' had members on '
                                + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                        },
                    },
                    series: series,
                }));
            },
        });
    }

    function loadCompareScoreChart() {

        if ($.isEmptyObject(appNames)) {
            return;
        }

        $.ajax({
            type: "GET",
            url: '/games/compare/' + $appsComparePage.attr('data-id') + '/reviews.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['mean_reviews_score'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('score-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: {
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
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' had a review score of ' + this.y.toLocaleString() + '% on '
                                + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                        },
                    },
                    series: series,
                }));

            },
        });
    }

    function loadCompareWishlistChart() {

        if ($.isEmptyObject(appNames)) {
            return;
        }

        $.ajax({
            type: "GET",
            url: '/games/compare/' + $appsComparePage.attr('data-id') + '/wishlists.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['mean_wishlist_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('wishlists-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: {
                        allowDecimals: false,
                        title: {text: ''},
                    },
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' is in ' + this.y.toLocaleString() + ' wishlists on '
                                + moment(this.key).format("dddd DD MMM YYYY");
                        },
                    },
                    series: series,
                }));
            },
        });
    }
}