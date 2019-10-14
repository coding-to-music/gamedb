if ($('#genres-page').length > 0 || $('#developers-page').length > 0 || $('#publishers-page').length > 0 || $('#tags-page').length > 0 || $('#categories-page').length > 0) {

    const searchFields = [
        $('#search'),
    ];

    $('table.table').gdbTable({
        searchFields: searchFields
    });
}

if ($('#stats-page').length > 0) {

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
                tooltip: {
                    formatter: function () {

                        const time = moment(this.key).format("dddd DD MMM YYYY @ HH:mm");

                        if (this.series.name === 'ingame') {
                            return this.y.toLocaleString() + ' people in a game on ' + time;
                        } else {
                            return this.y.toLocaleString() + ' people online on ' + time;
                        }
                    },
                },
                plotOptions: {
                    series: {
                        cursor: 'pointer',
                        point: {
                            events: {
                                click: function () {
                                    window.location.href = '/apps?score-low=' + this.x + '&score-high=' + (this.x + 1);
                                }
                            }
                        }
                    }
                },
                series: [
                    {
                        name: 'ingame',
                        marker: {symbol: 'circle'},
                        data: data['max_player_count'],
                    },
                    {
                        name: 'online',
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
                                    window.location.href = '/apps?score-low=' + this.x + '&score-high=' + (this.x + 1);
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
        url: '/stats/app-types.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            if (data === null) {
                data = [];
            }

            Highcharts.chart('types', $.extend(true, {}, defaultStatsChartOptions, {
                xAxis: {
                    labels: {
                        rotation: -20,
                    }
                },
                tooltip: {
                    formatter: function () {
                        return this.y.toLocaleString() + ' ' + this.key + ' apps';
                    },
                },
                plotOptions: {
                    series: {
                        cursor: 'pointer',
                        point: {
                            events: {
                                click: function () {
                                    let name = this.name.toLowerCase();
                                    if (name === 'unknown') {
                                        name = '';
                                    }
                                    window.location.href = '/apps?types=' + name;
                                }
                            }
                        }
                    }
                },
                series: [{
                    data: data,
                    dataLabels: {
                        enabled: true,
                        formatter: function () {
                            return this.y.toLocaleString();
                        }
                    }
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
                    type: 'area',
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
}
