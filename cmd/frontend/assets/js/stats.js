if ($('#stats-page').length > 0) {

    (function ($, window) {
        'use strict';

        loadAjaxOnObserve({
            // Steam
            "client-players": statsClientPlayers,
            "client-players2": statsClientPlayers2,
            "scores": statsAppScores,
            "release-dates": statsReleaseDates,
            "app-types": statsAppTypes,

            // Global Steam
            "player-countries": playerCountries,
            "player-levels": statsPlayerLevels,
            "player-update-days": statsPlayerUpdateDays,
        });

        //
        function statsClientPlayers() {

            $.ajax({
                type: "GET",
                url: '/stats/client-players.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    if (data === null) {
                        data = [];
                    }

                    Highcharts.chart('client-players', $.extend(true, {}, defaultChartOptions, {
                        yAxis: {
                            allowDecimals: false,
                            title: {
                                text: ''
                            }
                        },
                        tooltip: {
                            formatter: function () {

                                const time = moment(this.key).format("dddd DD MMM YYYY @ HH:mm");

                                if (this.series.name === 'Ingame') {
                                    return this.y.toLocaleString() + ' people in a game on ' + time;
                                } else {
                                    return this.y.toLocaleString() + ' people online on ' + time;
                                }
                            },
                        },
                        series: [
                            {
                                name: 'In Game',
                                marker: {symbol: 'circle'},
                                data: data['max_player_count'],
                                type: 'area',
                            },
                            {
                                name: 'Online',
                                marker: {symbol: 'circle'},
                                data: data['max_player_online'],
                                type: 'line',
                            },
                        ]
                    }));
                },
            });
        }

        function statsClientPlayers2() {

            $.ajax({
                type: "GET",
                url: '/stats/client-players2.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    if (data === null) {
                        data = [];
                    }

                    Highcharts.chart('client-players2', $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            zoomType: 'x',
                        },
                        yAxis: {
                            allowDecimals: false,
                            title: {
                                text: ''
                            }
                        },
                        tooltip: {
                            formatter: function () {

                                const time = moment(this.key).format("dddd DD MMM YYYY");

                                if (this.series.name === 'In Game') {
                                    return this.y.toLocaleString() + ' people in a game on ' + time;
                                } else {
                                    return this.y.toLocaleString() + ' people online on ' + time;
                                }
                            },
                        },
                        series: [
                            {
                                name: 'In Game',
                                marker: {symbol: 'circle'},
                                data: data['max_player_count'],
                                type: 'area',
                            },
                            {
                                name: 'Online',
                                marker: {symbol: 'circle'},
                                data: data['max_player_online'],
                                type: 'line',
                            },
                        ]
                    }));
                },
            });
        }

        function statsAppTypes() {

            $.ajax({
                type: "GET",
                url: '/stats/app-types.json',
                dataType: 'json',
                cache: true,
                success: function (data, textStatus, jqXHR) {

                    const $container = $('#app-types-table tbody');

                    $container.empty();

                    $container.json2html(
                        data.rows,
                        {
                            '<>': 'tr', 'data-link': '/games?types=${type}', 'html': [
                                {
                                    '<>': 'td', 'html': '${typef}',
                                },
                                {
                                    '<>': 'td', 'html': '${countf}'
                                },
                                {
                                    '<>': 'td', 'html': '${totalf}'
                                },
                            ]
                        },
                        {
                            prepend: false,
                        }
                    );

                    $('#total-price').text(data.total);

                    $('#app-types-table').gdbTable();
                },
            });
        }

        function statsReleaseDates() {

            $.ajax({
                type: "GET",
                url: '/stats/release-dates.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    if (data === null) {
                        data = [];
                    }

                    Highcharts.chart('release-dates-chart', $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            zoomType: 'x',
                        },
                        legend: {
                            enabled: false,
                        },
                        yAxis: {
                            allowDecimals: false,
                            title: {
                                text: ''
                            }
                        },
                        tooltip: {
                            formatter: function () {
                                return this.y.toLocaleString() + ' apps released on ' + moment(this.key).format("dddd DD MMM YYYY");
                            },
                        },
                        series: [{
                            data: data,
                        }],
                    }));
                },
            });
        }

        function statsPlayerLevels() {

            $.ajax({
                type: "GET",
                url: '/stats/player-levels.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    if (data === null) {
                        // data = [];
                    }

                    let categories = [];
                    let dataArray = [];

                    data.forEach(function (value, index, array) {
                        categories.push(data[index]['id']);
                        dataArray.push(value['count']);
                    });

                    Highcharts.chart('player-levels-chart', $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            type: 'column',
                        },
                        legend: {
                            enabled: false,
                        },
                        xAxis: {
                            type: 'category',
                            tickInterval: 5,
                            categories: categories,
                            labels: {
                                formatter: function () {
                                    return this.value.toLocaleString();
                                },
                            },
                        },
                        yAxis: {
                            type: 'logarithmic',
                            allowDecimals: false,
                            title: {
                                text: ''
                            },
                        },
                        tooltip: {
                            formatter: function () {
                                return this.y.toLocaleString() + ' players are level ' + this.x + '-' + (this.x + 9);
                            },
                        },
                        plotOptions: {
                            series: {
                                pointPadding: 0,
                                groupPadding: 0,
                            }
                        },
                        series: [{
                            data: dataArray,
                        }]
                    }));
                },
            });
        }

        function statsPlayerUpdateDays() {

            $.ajax({
                type: "GET",
                url: '/stats/player-update-dates.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    if (data === null) {
                        data = [];
                    }

                    Highcharts.chart('player-update-days-chart', $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            type: 'column',
                        },
                        legend: {
                            enabled: false,
                        },
                        yAxis: {
                            allowDecimals: false,
                            title: {
                                text: ''
                            },
                        },
                        tooltip: {
                            formatter: function () {
                                return this.y.toLocaleString() + ' players updated on ' + moment(this.x).format("dddd DD MMM YYYY");
                            },
                        },
                        plotOptions: {
                            series: {
                                pointPadding: 0,
                                groupPadding: 0,
                            }
                        },
                        series: [{
                            data: data,
                        }],
                    }));
                },
            });
        }

        function playerCountries() {

            $.ajax({
                type: "GET",
                url: '/stats/player-countries.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    Highcharts.chart('player-countries-chart', $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            type: 'column',
                        },
                        legend: {
                            enabled: false,
                        },
                        xAxis: {
                            type: 'category',
                        },
                        yAxis: {
                            type: 'logarithmic',
                            title: null,
                        },
                        plotOptions: {
                            series: {
                                pointPadding: 0,
                                groupPadding: 0,
                            }
                        },
                        series: [{
                            // name: 'Countries',
                            data: data['series'],
                        }],
                        drilldown: {series: data['drilldown']},
                    }));
                },
            });
        }

        function statsAppScores() {

            $.ajax({
                type: "GET",
                url: '/stats/app-scores.json',
                dataType: 'json',
                success: function (data, textStatus, jqXHR) {

                    if (data === null) {
                        data = [];
                    }

                    Highcharts.chart('scores-chart', $.extend(true, {}, defaultChartOptions, {
                        chart: {
                            type: 'column',
                        },
                        legend: {
                            enabled: false,
                        },
                        xAxis: {
                            type: 'category',
                            tickInterval: 5,
                        },
                        yAxis: {
                            allowDecimals: false,
                            title: {
                                text: ''
                            }
                        },
                        tooltip: {
                            formatter: function () {
                                return this.y.toLocaleString() + ' apps have ' + this.x + '/100';
                            },
                        },
                        plotOptions: {
                            series: {
                                pointPadding: 0,
                                groupPadding: 0,
                                cursor: 'pointer',
                                point: {
                                    events: {
                                        click: function () {
                                            window.location.href = '/games?score=' + this.x + '&score=' + (this.x + 1);
                                        }
                                    }
                                }
                            }
                        },
                        series: [{
                            data: data
                        }]
                    }));
                },
            });
        }

    })(jQuery, window);
}
