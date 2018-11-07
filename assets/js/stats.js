if ($('#stats-page').length > 0) {

    Highcharts.chart('scores', {
        chart: {
            type: 'column'
        },
        title: {
            text: ''
        },
        subtitle: {
            text: ''
        },
        xAxis: {
            categories: [1, 2, 3], // Only need a few to start the pattern off at 1
            crosshair: true,
            title: {
                text: ''
            }
        },
        yAxis: {
            min: 0,
            max: 100,
            title: {
                text: ''
            }
        },
        credits: {
            enabled: false
        },
        legend: {
            enabled: false
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
                            window.location.href = '/games#score-low=' + this.x + '&score-high=' + this.x;
                        }
                    }
                }
            }
        },
        series: [{
            data: scores,
            color: '#28a745',
        }]
    });
}
