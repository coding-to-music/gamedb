const $appsComparePage = $('#apps-compare-page');

if ($appsComparePage.length > 0) {

    loadAjaxOnObserve({
        'apps-table': loadCompareSearchTable,
        'players-chart': loadComparePlayersChart,
        'group-chart': loadCompareFollowersChart,
        'score-chart': loadCompareScoreChart,
        'wishlists-chart': loadCompareWishlistChart,
        'price-chart': loadComparePriceChart,
    });

    function loadCompareSearchTable() {

        const options = {
            "order": [[0, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[1]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[4] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></a>'
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
                        return row[9];
                    },
                    "orderable": false,
                    "visible": user.isLocal,
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
            cache: true,
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
            cache: true,
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
            cache: true,
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
                        // min: 0,
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
            cache: true,
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

    function loadComparePriceChart() {

        if ($.isEmptyObject(appNames)) {
            return;
        }

        $.ajax({
            type: "GET",
            url: '/games/compare/' + $appsComparePage.attr('data-id') + '/prices.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['price'],
                        type: 'line',
                        step: 'left',
                    });
                }

                Highcharts.chart('price-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: {
                        title: {
                            text: 'Price (' + user.userCurrencySymbol + ')'
                        },
                        allowDecimals: true,
                        min: 0,
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
            cache: true,
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