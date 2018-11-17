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
            }
        },
        yAxis: {
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
                    type: 'category',
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
                                    window.location.href = '/games?score-low=' + this.x + '&score-high=' + (this.x + 1);
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
                    type: 'category',
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
                                    window.location.href = '/games?types=' + (this.name.toLowerCase());
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
}
