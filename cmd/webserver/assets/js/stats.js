if ($('#genres-page').length > 0 || $('#developers-page').length > 0 || $('#publishers-page').length > 0 || $('#tags-page').length > 0 || $('#categories-page').length > 0) {

    const searchFields = [
        $('#search'),
    ];

    $('table.table').gdbTable({
        searchFields: searchFields
    });
}

if ($('#stats-page').length > 0) {

    (function ($, window) {
        'use strict';

        const config = {rootMargin: '50px 0px 50px 0px', threshold: 0};

        const callback1 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    statsAppTypes();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback1, config).observe(document.getElementById("app-types"));

        const callback2 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    statsReleaseDates();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback2, config).observe(document.getElementById("release-dates"));

        const callback3 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    statsPlayerLevels();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback3, config).observe(document.getElementById("player-levels"));

        const callback4 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    statsAppScores();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback4, config).observe(document.getElementById("scores"));

        const callback5 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    statsClientPlayers();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback5, config).observe(document.getElementById("client-players"));

        const callback6 = function (entries, self) {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    statsClientPlayers2();
                    self.unobserve(entry.target);
                }
            });
        };
        new IntersectionObserver(callback6, config).observe(document.getElementById("client-players2"));

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

                    const $container = $('#app-types tbody');

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

                    $('#app-types').gdbTable();
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

                    Highcharts.chart('release-dates', $.extend(true, {}, defaultChartOptions, {
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

                    Highcharts.chart('player-levels', $.extend(true, {}, defaultChartOptions, {
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
                            data: dataArray
                        }]
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

                    Highcharts.chart('scores', $.extend(true, {}, defaultChartOptions, {
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
