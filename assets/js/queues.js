if ($('#queues-page').length > 0) {

    Highcharts.chart('chart', {
        chart: {
            type: 'spline'
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
                text: 'Time'
            },
            labels: {
                step: 1
            }
        },
        yAxis: {
            title: {
                text: 'Queue Size'
            }
        },
        series: [{
            color: '#28a745',
        }],
        tooltip: {
            formatter: function () {
                return this.y + ' items in the queue';
            },
        },
        data: {
            rowsURL: location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '') + '/queues/ajax.json',
            enablePolling: true,
            dataRefreshRate: 60
        }
    });
}
