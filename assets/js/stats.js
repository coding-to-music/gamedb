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
                pointWidth: 10,
                cursor: 'pointer',
                point: {
                    events: {
                        click: function () {
                            window.location.href = '/games#price-low=' + this.x + '&price-high=' + this.x;
                        }
                    }
                }
            }
        },
        series: [{
            //data: [49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0, 176.0, 135.6, 148.5, 216.4, 194.1, 95.6, 54.4, 49.9, 71.5, 106.4, 129.2, 144.0],
            data: scores,
            color: '#28a745',
        }]
    });
}
