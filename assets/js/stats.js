if ($('#stats-page').length > 0) {

    const defaultStatsChartOptions = {
        chart: {
            type: 'column'
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
        url: '/stats/client-players',
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
                        return this.y.toLocaleString() + ' people logged into steam on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
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
                    data: data['max_player_count']
                }]
            }));
        },
    });

    $.ajax({
        type: "GET",
        url: '/stats/app-scores',
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
        url: '/stats/app-types',
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
                                    window.location.href = '/apps?types=' + (this.name.toLowerCase());
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
        url: '/stats/release-dates',
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
