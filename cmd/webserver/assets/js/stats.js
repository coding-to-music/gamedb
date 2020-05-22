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

        const defaultStatsChartOptions = {
            chart: {
                type: 'column',
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
                enabled: false
            },
            xAxis: {
                title: {
                    text: ''
                },
                type: 'category'
            },
            yAxis: {
                allowDecimals: false,
                title: {
                    text: ''
                }
            },
            series: [{
                color: '#28a745',
            }],
            plotOptions: {
                series: {
                    pointPadding: 0,
                    groupPadding: 0,
                }
            }
        };

        $.ajax({
            type: "GET",
            url: '/stats/client-players.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('client-players', $.extend(true, {}, defaultStatsChartOptions, {
                    chart: {
                        type: 'area',
                    },
                    xAxis: {
                        type: 'datetime',
                        // tickInterval: 5,
                    },
                    legend: {
                        enabled: true,
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
                        },
                        {
                            name: 'Online',
                            marker: {symbol: 'circle'},
                            color: '#007bff',
                            data: data['max_player_online'],
                            type: 'line',
                        },
                    ]
                }));
            },
        });

        $.ajax({
            type: "GET",
            url: '/stats/client-players2.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('client-players2', $.extend(true, {}, defaultStatsChartOptions, {
                    chart: {
                        type: 'area',
                    },
                    xAxis: {
                        type: 'datetime',
                        // tickInterval: 5,
                    },
                    legend: {
                        enabled: true,
                    },
                    tooltip: {
                        formatter: function () {

                            const time = moment(this.key).format("dddd DD MMM YYYY @ HH:mm");

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
                        },
                        {
                            name: 'Online',
                            marker: {symbol: 'circle'},
                            color: '#007bff',
                            data: data['max_player_online'],
                            type: 'line',
                        },
                    ]
                }));
            },
        });

        $.ajax({
            type: "GET",
            url: '/stats/app-scores.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('scores', $.extend(true, {}, defaultStatsChartOptions, {
                    xAxis: {
                        tickInterval: 5,
                    },
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' apps have ' + this.x + '/100';
                        },
                    },
                    plotOptions: {
                        series: {
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

                Highcharts.chart('player-levels', $.extend(true, {}, defaultStatsChartOptions, {
                    xAxis: {
                        tickInterval: 5,
                        categories: categories,
                    },
                    yAxis: {
                        type: 'logarithmic',
                    },
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' players are level ' + this.x + '-' + (this.x + 9);
                        },
                    },
                    series: [{
                        data: dataArray
                    }]
                }));
            },
        });

        $.ajax({
            type: "GET",
            url: '/stats/release-dates.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                Highcharts.chart('release-dates', $.extend(true, {}, defaultStatsChartOptions, {
                    chart: {
                        type: 'column',
                    },
                    xAxis: {
                        type: 'datetime'
                    },
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' apps released on ' + moment(this.key).format("dddd DD MMM YYYY");
                        },
                    },
                    series: [{
                        data: data
                    }],
                    plotOptions: {
                        series: {
                            pointWidth: 2,
                        },
                        area: {
                            lineWidth: 1,
                            states: {
                                hover: {
                                    lineWidth: 1
                                }
                            },
                        }
                    },
                }));
            },
        });

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

    })(jQuery, window);
}
