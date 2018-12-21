if ($('#stats-page').length > 0) {

    const columnDefaults = {
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
        url: '/stats/app-scores',
        success: function (data, textStatus, jqXHR) {


            Highcharts.chart('scores', $.extend(true, {}, columnDefaults, {
                xAxis: {
                    tickInterval: 5,
                },
                tooltip: {
                    formatter: function () {
                        return this.y + ' apps have ' + this.x + '/100';
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
        dataType: 'json'
    });

    $.ajax({
        type: "GET",
        url: '/stats/app-types',
        success: function (data, textStatus, jqXHR) {

            Highcharts.chart('types', $.extend(true, {}, columnDefaults, {
                xAxis: {
                    labels: {
                        rotation: -20,
                    }
                },
                tooltip: {
                    formatter: function () {
                        return this.y + ' ' + this.key + ' apps';
                    },
                },
                plotOptions: {
                    series: {
                        cursor: 'pointer',
                        point: {
                            events: {
                                click: function () {
                                    console.log(this);
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
        dataType: 'json'
    });

    $.ajax({
        type: "GET",
        url: '/stats/ranked-countries',
        success: function (data, textStatus, jqXHR) {

            Highcharts.chart('countries', $.extend(true, {}, columnDefaults, {
                xAxis: {
                    tickInterval: 1,
                },
                tooltip: {
                    formatter: function () {
                        return this.y + ' ' + this.key + ' players';
                    },
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
        dataType: 'json'
    });

    $.ajax({
        type: "GET",
        url: '/stats/release-dates',
        success: function (data, textStatus, jqXHR) {

            Highcharts.chart('release-dates', $.extend(true, {}, columnDefaults, {
                chart: {
                    type: 'area',
                },
                xAxis: {
                    type: 'datetime'
                },
                tooltip: {
                    formatter: function () {
                        return this.y + ' apps released on ' + moment(this.key).format("DD MMM YYYY");
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
        dataType: 'json'
    });
}
